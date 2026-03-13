package agent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// SpawnResult holds the outcome of launching a new agent.
type SpawnResult struct {
	Name     string
	Path     string
	Branch   string
	Pid      int
	LogFile  string
	Err      error
}

// Spawn creates a new worktree and launches a Claude Code session in it.
// The prompt can be raw text or a path to an .md file.
func Spawn(repoRoot, prompt string) SpawnResult {
	// Derive a short name from the prompt
	name := deriveName(prompt)
	branch := name
	worktreePath := filepath.Join(repoRoot, ".claude", "worktrees", name)
	logFile := filepath.Join(repoRoot, ".claude", "worktrees", name+".log")

	// Create worktree
	wtCmd := exec.Command("git", "worktree", "add", worktreePath, "-b", branch, "origin/main")
	wtCmd.Dir = repoRoot
	if out, err := wtCmd.CombinedOutput(); err != nil {
		return SpawnResult{Name: name, Err: fmt.Errorf("worktree: %s", string(out))}
	}

	// Resolve prompt — if it looks like a file path, read it
	resolvedPrompt := resolvePrompt(repoRoot, prompt)

	// Open log file
	logF, err := os.Create(logFile)
	if err != nil {
		return SpawnResult{Name: name, Path: worktreePath, Err: fmt.Errorf("log file: %w", err)}
	}

	// Launch claude in the worktree
	claudeCmd := exec.Command("claude", "--print", resolvedPrompt)
	claudeCmd.Dir = worktreePath
	claudeCmd.Stdout = logF
	claudeCmd.Stderr = logF

	if err := claudeCmd.Start(); err != nil {
		logF.Close()
		return SpawnResult{Name: name, Path: worktreePath, Err: fmt.Errorf("claude: %w", err)}
	}

	// Detach — the process runs in background
	go func() {
		claudeCmd.Wait()
		logF.Close()
	}()

	return SpawnResult{
		Name:    name,
		Path:    worktreePath,
		Branch:  branch,
		Pid:     claudeCmd.Process.Pid,
		LogFile: logFile,
	}
}

// resolvePrompt checks if the input is a file path and reads it, otherwise returns as-is.
func resolvePrompt(repoRoot, prompt string) string {
	prompt = strings.TrimSpace(prompt)

	// Check if it's a file path (absolute or relative)
	candidates := []string{
		prompt,
		filepath.Join(repoRoot, prompt),
	}

	for _, path := range candidates {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			if content, err := os.ReadFile(path); err == nil {
				return string(content)
			}
		}
	}

	return prompt
}

// deriveName creates a short kebab-case name from a prompt.
func deriveName(prompt string) string {
	prompt = strings.TrimSpace(prompt)

	// If it's a file path, use the filename
	if strings.HasSuffix(prompt, ".md") || strings.Contains(prompt, "/") {
		base := filepath.Base(prompt)
		base = strings.TrimSuffix(base, filepath.Ext(base))
		return sanitize(base)
	}

	// Take first few meaningful words
	words := strings.Fields(strings.ToLower(prompt))
	if len(words) > 4 {
		words = words[:4]
	}

	name := strings.Join(words, "-")
	name = sanitize(name)

	if len(name) > 30 {
		name = name[:30]
	}

	if name == "" {
		name = fmt.Sprintf("agent-%d", time.Now().Unix()%10000)
	}

	return name
}

func sanitize(s string) string {
	var result []byte
	for _, c := range []byte(s) {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			result = append(result, c)
		} else if c == ' ' || c == '_' || c == '/' {
			result = append(result, '-')
		}
	}
	// Remove leading/trailing/double dashes
	s = string(result)
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	return s
}
