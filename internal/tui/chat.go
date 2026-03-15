package tui

import (
	"io"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/goldenfocus/multitab/internal/commander"
	"github.com/goldenfocus/multitab/internal/queue"
)

// ─────────────────────────────────────────────────
// Voice mode
// ─────────────────────────────────────────────────

type voiceMode int

const (
	voiceAuto   voiceMode = iota // speak every response automatically
	voiceManual                  // press 'v' to hear the last response
	voiceOff                     // no voice at all
)

func (v voiceMode) String() string {
	switch v {
	case voiceAuto:
		return "AUTO"
	case voiceManual:
		return "MANUAL"
	case voiceOff:
		return "OFF"
	}
	return ""
}

func (v voiceMode) Next() voiceMode {
	return (v + 1) % 3
}

// ─────────────────────────────────────────────────
// Chat messages (tea.Msg types)
// ─────────────────────────────────────────────────

type chatStreamStartMsg struct {
	reader io.ReadCloser
	proc   *exec.Cmd
}

type chatChunkMsg struct {
	text string
}

type chatStreamDoneMsg struct{}

type chatStreamErrMsg struct {
	err error
}

type chatSpeakDoneMsg struct{}

// ─────────────────────────────────────────────────
// Streaming commands
// ─────────────────────────────────────────────────

// startChatStreamCmd launches claude --print with full agent context.
func startChatStreamCmd(repoRoot string, state *queue.State, history []commander.Message, userMsg string) tea.Cmd {
	return func() tea.Msg {
		prompt := commander.BuildPrompt(state, history, userMsg)
		reader, proc, err := commander.StartStream(repoRoot, prompt)
		if err != nil {
			return chatStreamErrMsg{err: err}
		}
		return chatStreamStartMsg{reader: reader, proc: proc}
	}
}

// readChatChunkCmd reads the next chunk from the stream.
func readChatChunkCmd(reader io.ReadCloser) tea.Cmd {
	return func() tea.Msg {
		buf := make([]byte, 256)
		n, err := reader.Read(buf)
		if err != nil {
			return chatStreamDoneMsg{}
		}
		return chatChunkMsg{text: string(buf[:n])}
	}
}

// speakCmd speaks text aloud using macOS say.
func speakCmd(text, voice string) tea.Cmd {
	return func() tea.Msg {
		proc, err := commander.Speak(text, voice)
		if err != nil {
			return chatSpeakDoneMsg{}
		}
		if proc != nil {
			proc.Wait()
		}
		return chatSpeakDoneMsg{}
	}
}

// ─────────────────────────────────────────────────
// Key handling
// ─────────────────────────────────────────────────

func handleChatKeys(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc":
		// Unfocus chat — return keyboard to dashboard shortcuts
		if m.chatProc != nil {
			m.chatProc.Process.Kill()
			m.chatProc.Wait()
			m.chatProc = nil
			m.chatReader = nil
		}
		if m.sayProc != nil {
			commander.KillSpeech(m.sayProc)
			m.sayProc = nil
			m.speaking = false
		}
		m.chatStreaming = false
		m.chatFocused = false
		m.chatInput.Blur()
		return m, nil

	case "enter":
		if m.chatStreaming {
			return m, nil // ignore while streaming
		}
		userMsg := m.chatInput.Value()
		if userMsg == "" {
			return m, nil
		}

		// Capture prior history (before appending current message)
		// BuildPrompt adds userMsg separately, so don't include it in history
		priorHistory := make([]commander.Message, len(m.chatHistory))
		copy(priorHistory, m.chatHistory)

		// Add user message to display history
		m.chatHistory = append(m.chatHistory, commander.Message{
			Role:    "user",
			Content: userMsg,
		})
		m.chatInput.SetValue("")
		m.chatStreaming = true
		m.chatStreamBuf = ""

		return m, startChatStreamCmd(m.repoRoot, m.state, priorHistory, userMsg)

	case "ctrl+v":
		// Cycle voice mode: AUTO → MANUAL → OFF → AUTO
		m.voice = m.voice.Next()
		return m, nil

	case "ctrl+p":
		// Play/speak the last commander response (manual mode)
		if !m.speaking && len(m.chatHistory) > 0 {
			// Find the last commander message
			for i := len(m.chatHistory) - 1; i >= 0; i-- {
				if m.chatHistory[i].Role == "commander" {
					m.speaking = true
					return m, speakCmd(m.chatHistory[i].Content, m.voiceID)
				}
			}
		}
		// If currently speaking, stop it
		if m.speaking && m.sayProc != nil {
			commander.KillSpeech(m.sayProc)
			m.sayProc = nil
			m.speaking = false
		}
		return m, nil

	case "ctrl+l":
		// Clear chat history
		if !m.chatStreaming {
			m.chatHistory = nil
			m.chatStreamBuf = ""
		}
		return m, nil
	}

	// Forward all other keys to the text input
	var cmd tea.Cmd
	m.chatInput, cmd = m.chatInput.Update(msg)
	return m, cmd
}

// ─────────────────────────────────────────────────
// Chat message rendering (used by renderChatPanel in view.go)
// ─────────────────────────────────────────────────

func renderChatMessages(history []commander.Message, streamBuf string, streaming bool, width, tick int) string {
	if len(history) == 0 && streamBuf == "" {
		var lines []string
		lines = append(lines, "")
		lines = append(lines, chatSystemStyle.Render("  Commander standing by."))
		lines = append(lines, "")
		lines = append(lines, chatSystemStyle.Render("  I can see all your agents, conflicts,"))
		lines = append(lines, chatSystemStyle.Render("  migrations, and deploy status in real-time."))
		lines = append(lines, "")
		lines = append(lines, chatSystemStyle.Render("  Ask me anything. Or just talk."))
		lines = append(lines, "")
		return strings.Join(lines, "\n")
	}

	var lines []string
	lines = append(lines, "")

	for _, msg := range history {
		if msg.Role == "user" {
			lines = append(lines, chatUserLabelStyle.Render("  YOU"))
			// Word-wrap user message
			wrapped := wordWrap(msg.Content, width-4)
			for _, wl := range strings.Split(wrapped, "\n") {
				lines = append(lines, chatUserStyle.Render("  "+wl))
			}
			lines = append(lines, "")
		} else {
			lines = append(lines, chatCmdLabelStyle.Render("  COMMANDER"))
			wrapped := wordWrap(msg.Content, width-4)
			for _, wl := range strings.Split(wrapped, "\n") {
				lines = append(lines, chatCmdStyle.Render("  "+wl))
			}
			lines = append(lines, "")
		}
	}

	// Streaming buffer (response still coming in)
	if streaming && streamBuf != "" {
		lines = append(lines, chatCmdLabelStyle.Render("  COMMANDER"))
		wrapped := wordWrap(streamBuf, width-4)
		for _, wl := range strings.Split(wrapped, "\n") {
			lines = append(lines, chatCmdStyle.Render("  "+wl))
		}
		cursor := []string{"█", "▓", "▒", "░"}
		lines = append(lines, chatCmdStyle.Render("  "+cursor[tick%len(cursor)]))
		lines = append(lines, "")
	} else if streaming && streamBuf == "" {
		spinner := spinnerFrames[tick%len(spinnerFrames)]
		lines = append(lines, chatCmdLabelStyle.Render("  COMMANDER"))
		lines = append(lines, chatCmdStyle.Render("  "+spinner+" ..."))
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func renderVoiceIndicator(voice voiceMode, speaking bool, tick int) string {
	var indicator string
	switch {
	case speaking:
		frames := []string{"🔊", "🔉", "🔈", "🔉"}
		indicator = frames[tick%len(frames)] + " speaking..."
		return "  " + pushStepActiveStyle.Render(indicator)
	case voice == voiceAuto:
		indicator = "🔊 voice: AUTO"
	case voice == voiceManual:
		indicator = "🔇 voice: MANUAL (ctrl+p to play)"
	case voice == voiceOff:
		indicator = "🔇 voice: OFF"
	}
	return "  " + statusIndicatorStyle.Render(indicator)
}

// wordWrap breaks text into lines of at most maxWidth characters.
func wordWrap(text string, maxWidth int) string {
	if maxWidth <= 0 {
		maxWidth = 60
	}
	var result []string
	for _, paragraph := range strings.Split(text, "\n") {
		if paragraph == "" {
			result = append(result, "")
			continue
		}
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			result = append(result, "")
			continue
		}
		line := words[0]
		for _, w := range words[1:] {
			if len(line)+1+len(w) > maxWidth {
				result = append(result, line)
				line = w
			} else {
				line += " " + w
			}
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

// chatTickCmd returns a faster tick for the streaming cursor animation.
func chatTickCmd() tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(t time.Time) tea.Msg {
		return chatTickMsg(t)
	})
}

type chatTickMsg time.Time
