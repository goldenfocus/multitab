package agent

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ConversationInfo holds extracted data from a Claude session JSONL.
type ConversationInfo struct {
	SessionID    string
	LastPrompt   string    // last human message text
	LastPromptAt time.Time // when it was sent
	MessageCount int       // total human messages
	StartedAt    time.Time
	Branch       string
	CWD          string
}

// Turn represents one human↔assistant exchange in the conversation.
type Turn struct {
	HumanText     string
	HumanTime     string // local HH:MM
	AssistantText string
	AssistantTime string // local HH:MM
	Number        int
}

// jsonlEntry represents a single line in a Claude JSONL file.
type jsonlEntry struct {
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp"`
	CWD       string          `json:"cwd"`
	GitBranch string          `json:"gitBranch"`
	SessionID string          `json:"sessionId"`
	Message   json.RawMessage `json:"message"`
}

// messageWrapper holds the message envelope.
type messageWrapper struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// contentBlock represents one item in the content array.
type contentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// FindConversation locates and parses the Claude session JSONL for a given worktree path.
// Strategy:
//  1. Check for a worktree-specific project dir (~/.claude/projects/<encoded-worktree-path>/)
//  2. If not found, scan the parent repo's project dir for sessions that reference this worktree in their cwd
func FindConversation(worktreePath string) (*ConversationInfo, error) {
	jsonlFile, err := findBestJSONL(worktreePath)
	if err != nil || jsonlFile == "" {
		return nil, err
	}
	return parseJSONL(jsonlFile)
}

// ReadFullChat extracts all human messages from the session for display in the log viewer.
func ReadFullChat(worktreePath string) (string, error) {
	jsonlFile, err := findBestJSONL(worktreePath)
	if err != nil || jsonlFile == "" {
		return "No conversation found.", nil
	}
	return formatChat(jsonlFile)
}

// findBestJSONL finds the most relevant JSONL file for a worktree.
// First tries the worktree-specific project dir, then falls back to scanning
// the parent repo's project dir for sessions that reference this worktree path.
func findBestJSONL(worktreePath string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	projectsDir := filepath.Join(homeDir, ".claude", "projects")

	// Strategy 1: Direct match — worktree-specific project dir
	encoded := encodeProjectPath(worktreePath)
	directDir := filepath.Join(projectsDir, encoded)
	if jsonl := newestJSONL(directDir); jsonl != "" {
		return jsonl, nil
	}

	// Strategy 2: Scan parent project dir for sessions referencing this worktree in cwd.
	// Walk up from the worktree path to find the parent repo root
	// e.g., /Users/vibeyang/p69/.claude/worktrees/recep-override → /Users/vibeyang/p69
	parentDir := findParentRepoDir(worktreePath)
	if parentDir == "" {
		return "", nil
	}
	parentEncoded := encodeProjectPath(parentDir)
	parentProjectDir := filepath.Join(projectsDir, parentEncoded)

	entries, err := os.ReadDir(parentProjectDir)
	if err != nil {
		return "", nil
	}

	var bestFile string
	var bestTime time.Time

	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		fullPath := filepath.Join(parentProjectDir, e.Name())

		// Quick scan: check first 100 lines for cwd matching the worktree path
		if sessionMatchesWorktree(fullPath, worktreePath) {
			info, _ := e.Info()
			if info != nil && info.ModTime().After(bestTime) {
				bestFile = fullPath
				bestTime = info.ModTime()
			}
		}
	}

	return bestFile, nil
}

// newestJSONL returns the most recently modified JSONL file in a directory.
func newestJSONL(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	var jsonlFiles []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".jsonl") {
			jsonlFiles = append(jsonlFiles, filepath.Join(dir, e.Name()))
		}
	}
	if len(jsonlFiles) == 0 {
		return ""
	}

	sort.Slice(jsonlFiles, func(i, j int) bool {
		si, _ := os.Stat(jsonlFiles[i])
		sj, _ := os.Stat(jsonlFiles[j])
		if si == nil || sj == nil {
			return false
		}
		return si.ModTime().Before(sj.ModTime())
	})

	return jsonlFiles[len(jsonlFiles)-1]
}

// sessionMatchesWorktree checks if a JSONL file references the given worktree path.
func sessionMatchesWorktree(jsonlPath, worktreePath string) bool {
	f, err := os.Open(jsonlPath)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 5*1024*1024)

	for i := 0; i < 200 && scanner.Scan(); i++ {
		var entry jsonlEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}
		if entry.CWD == worktreePath {
			return true
		}
	}
	return false
}

// findParentRepoDir walks up from a worktree path to find the repo root.
// e.g., /a/b/.claude/worktrees/foo → /a/b
func findParentRepoDir(worktreePath string) string {
	// Pattern: .claude/worktrees/<name> is always inside the repo root
	parts := strings.Split(worktreePath, string(filepath.Separator))
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == ".claude" && i > 0 {
			return strings.Join(parts[:i], string(filepath.Separator))
		}
	}
	// Fallback: try going up directories looking for .git
	dir := worktreePath
	for {
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		if _, err := os.Stat(filepath.Join(parent, ".git")); err == nil {
			return parent
		}
		dir = parent
	}
}

func parseJSONL(path string) (*ConversationInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info := &ConversationInfo{
		SessionID: strings.TrimSuffix(filepath.Base(path), ".jsonl"),
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024) // 10MB buffer for large lines

	for scanner.Scan() {
		var entry jsonlEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		// Capture metadata from first entry
		if info.CWD == "" && entry.CWD != "" {
			info.CWD = entry.CWD
		}
		if info.Branch == "" && entry.GitBranch != "" {
			info.Branch = entry.GitBranch
		}

		// Track start time
		if !info.StartedAt.IsZero() {
		} else if entry.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339Nano, entry.Timestamp); err == nil {
				info.StartedAt = t
			}
		}

		if entry.Type != "user" {
			continue
		}

		text := extractUserText(entry.Message)
		if text == "" {
			continue
		}

		info.MessageCount++
		info.LastPrompt = text

		if entry.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339Nano, entry.Timestamp); err == nil {
				info.LastPromptAt = t
			}
		}
	}

	if info.MessageCount == 0 {
		return nil, nil
	}

	return info, nil
}

func formatChat(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var sb strings.Builder
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	humanNum := 0
	for scanner.Scan() {
		var entry jsonlEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		ts := ""
		if entry.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339Nano, entry.Timestamp); err == nil {
				ts = t.Local().Format("15:04")
			}
		}

		switch entry.Type {
		case "user":
			text := extractUserText(entry.Message)
			if text == "" {
				continue
			}
			humanNum++

			sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			if ts != "" {
				sb.WriteString("  HUMAN #" + itoa(humanNum) + "  [" + ts + "]\n")
			} else {
				sb.WriteString("  HUMAN #" + itoa(humanNum) + "\n")
			}
			sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

			display := text
			if len(display) > 2000 {
				display = display[:2000] + "\n  ... (truncated)"
			}
			sb.WriteString(display)
			sb.WriteString("\n\n")

		case "assistant":
			text := extractAssistantText(entry.Message)
			if text == "" {
				continue
			}

			sb.WriteString("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄\n")
			if ts != "" {
				sb.WriteString("  CLAUDE  [" + ts + "]\n")
			} else {
				sb.WriteString("  CLAUDE\n")
			}
			sb.WriteString("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄\n")

			display := text
			if len(display) > 3000 {
				display = display[:3000] + "\n  ... (truncated)"
			}
			sb.WriteString(display)
			sb.WriteString("\n\n")
		}
	}

	if humanNum == 0 {
		return "No conversation found in this session.", nil
	}

	return sb.String(), nil
}

// extractAssistantText pulls readable text from assistant messages, skipping thinking blocks.
func extractAssistantText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var wrapper messageWrapper
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return ""
	}

	if wrapper.Role != "assistant" {
		return ""
	}

	return extractAssistantContentBlocks(wrapper.Content)
}

// extractAssistantContentBlocks filters to only text blocks (no thinking, no tool_use).
func extractAssistantContentBlocks(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}

	var blocks []contentBlock
	if err := json.Unmarshal(raw, &blocks); err == nil {
		var texts []string
		for _, b := range blocks {
			if b.Type == "text" && b.Text != "" {
				texts = append(texts, b.Text)
			}
			// Skip thinking, tool_use, tool_result — noise for the curated view
		}
		return strings.Join(texts, "\n")
	}

	return ""
}

// extractUserText pulls readable text from a user message's content field.
func extractUserText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	// First try: content is a wrapper with role + content
	var wrapper messageWrapper
	if err := json.Unmarshal(raw, &wrapper); err == nil && wrapper.Role == "user" {
		return extractContentBlocks(wrapper.Content)
	}

	// Fallback: content might be a direct string
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}

	return ""
}

// extractContentBlocks handles content as either a string or array of blocks.
func extractContentBlocks(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	// Try as string first
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}

	// Try as array of content blocks
	var blocks []contentBlock
	if err := json.Unmarshal(raw, &blocks); err == nil {
		var texts []string
		for _, b := range blocks {
			if b.Type == "text" && b.Text != "" {
				texts = append(texts, b.Text)
			} else if b.Type == "image" {
				texts = append(texts, "[image]")
			} else if b.Type == "tool_result" {
				// Skip tool results — they're system noise
			}
		}
		return strings.Join(texts, "\n")
	}

	return ""
}

// encodeProjectPath converts a filesystem path to Claude's project directory encoding.
// /Users/vibeyang/p69/.claude/worktrees/fix-history → -Users-vibeyang-p69--claude-worktrees-fix-history
func encodeProjectPath(path string) string {
	// Remove trailing slash
	path = strings.TrimRight(path, "/")
	// Replace / with -
	return strings.ReplaceAll(path, "/", "-")
}

// ParseTurns extracts conversation turns from the most recent JSONL session for a worktree.
func ParseTurns(worktreePath string) ([]Turn, error) {
	jsonlFile, err := findBestJSONL(worktreePath)
	if err != nil || jsonlFile == "" {
		return nil, err
	}
	return parseTurnsFromFile(jsonlFile)
}

func parseTurnsFromFile(path string) ([]Turn, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	var turns []Turn
	var current Turn
	turnNum := 0
	inTurn := false

	for scanner.Scan() {
		var entry jsonlEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue
		}

		ts := ""
		if entry.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339Nano, entry.Timestamp); err == nil {
				ts = t.Local().Format("15:04")
			}
		}

		switch entry.Type {
		case "user":
			text := extractUserText(entry.Message)
			if text == "" {
				continue
			}

			// If we already have a turn in progress, save it
			if inTurn {
				turns = append(turns, current)
			}

			turnNum++
			current = Turn{
				HumanText: text,
				HumanTime: ts,
				Number:    turnNum,
			}
			inTurn = true

		case "assistant":
			if !inTurn {
				continue
			}
			text := extractAssistantText(entry.Message)
			if text == "" {
				continue
			}
			// Append to current turn's assistant text (there can be multiple assistant entries per turn)
			if current.AssistantText != "" {
				current.AssistantText += "\n" + text
			} else {
				current.AssistantText = text
				current.AssistantTime = ts
			}
		}
	}

	// Don't forget the last turn
	if inTurn {
		turns = append(turns, current)
	}

	return turns, nil
}

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}
