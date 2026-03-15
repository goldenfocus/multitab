package commander

import (
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/goldenfocus/multitab/internal/git"
	"github.com/goldenfocus/multitab/internal/queue"
)

// Message represents one chat exchange.
type Message struct {
	Role    string // "user" or "commander"
	Content string
}

// BuildPrompt creates the full prompt for claude --print with live agent context.
func BuildPrompt(state *queue.State, history []Message, userMsg string) string {
	var sb strings.Builder

	sb.WriteString("You are COMMANDER — the AI brain of Multitab, a multi-agent deployment cockpit.\n")
	sb.WriteString("You have full real-time visibility into all running Claude Code agents.\n")
	sb.WriteString("Be concise, direct, and sharp. You are mission control — not a chatbot.\n")
	sb.WriteString("Keep responses under 3 sentences unless the user asks for detail.\n")
	sb.WriteString("Your responses will be spoken aloud via TTS, so write naturally — no markdown, no bullet lists, no code blocks unless asked.\n\n")

	if state != nil {
		sb.WriteString("══ LIVE AGENT STATUS ══\n")
		for _, a := range state.Agents {
			sb.WriteString(fmt.Sprintf("  %-22s  %s  %d commits  %d files",
				a.Name, a.Status.String(), a.Commits, a.Files))
			if a.LastPrompt != "" {
				prompt := a.LastPrompt
				if len(prompt) > 80 {
					prompt = prompt[:80] + "..."
				}
				sb.WriteString(fmt.Sprintf("  last: %q", prompt))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")

		// System alerts
		var alerts []string
		if len(state.Conflicts) > 0 {
			alerts = append(alerts, fmt.Sprintf("%d file conflicts across agents", len(state.Conflicts)))
		}
		if state.HasMigrations {
			alerts = append(alerts, "pending migrations")
		}
		staleCount := 0
		for _, a := range state.Agents {
			if a.Status == git.StatusStale || a.Status == git.StatusAbandoned {
				staleCount++
			}
		}
		if staleCount > 0 {
			alerts = append(alerts, fmt.Sprintf("%d stale/abandoned worktrees", staleCount))
		}
		if len(alerts) > 0 {
			sb.WriteString("⚠ ALERTS: " + strings.Join(alerts, " | ") + "\n")
		}

		sb.WriteString(fmt.Sprintf("Deploy queue: %d/%d ready (%d%%)\n\n",
			state.ReadyCount, state.TotalCount,
			safePct(state.ReadyCount, state.TotalCount)))
	}

	if len(history) > 0 {
		sb.WriteString("══ CONVERSATION ══\n")
		for _, m := range history {
			prefix := "USER"
			if m.Role == "commander" {
				prefix = "COMMANDER"
			}
			sb.WriteString(fmt.Sprintf("%s: %s\n", prefix, m.Content))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("USER: " + userMsg)

	return sb.String()
}

// StartStream launches claude --print and returns a reader for streaming output.
func StartStream(repoRoot, prompt string) (io.ReadCloser, *exec.Cmd, error) {
	cmd := exec.Command("claude", "--print", prompt)
	cmd.Dir = repoRoot

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("stdout pipe: %w", err)
	}

	// Merge stderr into stdout so we catch errors too
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("start claude: %w", err)
	}

	return stdout, cmd, nil
}

// Speak runs macOS `say` to speak text aloud. Returns the process handle.
func Speak(text, voice string) (*exec.Cmd, error) {
	if voice == "" {
		voice = "Samantha"
	}
	// Clean text for speech — strip markdown artifacts
	text = cleanForSpeech(text)
	if text == "" {
		return nil, nil
	}

	cmd := exec.Command("say", "-v", voice, "-r", "210", text)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

// KillSpeech stops any active speech process.
func KillSpeech(proc *exec.Cmd) {
	if proc != nil && proc.Process != nil {
		proc.Process.Kill()
		proc.Wait()
	}
}

// cleanForSpeech removes markdown and code artifacts that sound bad when spoken.
func cleanForSpeech(text string) string {
	text = strings.TrimSpace(text)
	// Remove code blocks
	for strings.Contains(text, "```") {
		start := strings.Index(text, "```")
		end := strings.Index(text[start+3:], "```")
		if end == -1 {
			text = text[:start]
		} else {
			text = text[:start] + text[start+3+end+3:]
		}
	}
	// Remove inline code backticks
	text = strings.ReplaceAll(text, "`", "")
	// Remove markdown headers
	lines := strings.Split(text, "\n")
	var clean []string
	for _, line := range lines {
		line = strings.TrimLeft(line, "# ")
		clean = append(clean, line)
	}
	text = strings.Join(clean, "\n")
	// Remove bullet markers
	text = strings.ReplaceAll(text, "- ", "")
	text = strings.ReplaceAll(text, "* ", "")
	return strings.TrimSpace(text)
}

func safePct(a, b int) int {
	if b == 0 {
		return 0
	}
	return (a * 100) / b
}
