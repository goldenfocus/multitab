package tui

import (
	"fmt"
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
		// Stop any streaming/speaking and return to dashboard
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
		m.mode = viewDashboard
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
// Chat view rendering
// ─────────────────────────────────────────────────

func (m Model) renderChatView() string {
	var sections []string

	maxWidth := clampInt(m.width-4, 60, 80)
	innerWidth := maxWidth - 6

	// Header
	sections = append(sections, renderChatHeader(m.tick))

	// Chat history panel (scrollable viewport)
	chatContent := renderChatMessages(m.chatHistory, m.chatStreamBuf, m.chatStreaming, innerWidth, m.tick)
	m.viewport.SetContent(chatContent)
	sections = append(sections, "\n"+panelStyle.Width(innerWidth).Render(m.viewport.View()))

	// Voice status indicator
	voiceIndicator := renderVoiceIndicator(m.voice, m.speaking, m.tick)
	sections = append(sections, voiceIndicator)

	// Input panel
	var inputLines []string
	inputLines = append(inputLines, "")
	if m.chatStreaming {
		spinner := spinnerFrames[m.tick%len(spinnerFrames)]
		inputLines = append(inputLines, "  "+pushStepActiveStyle.Render(spinner+" commander is thinking..."))
	} else {
		inputLines = append(inputLines, "  "+chatPromptStyle.Render("▸")+" "+m.chatInput.View())
	}
	inputLines = append(inputLines, "")
	sections = append(sections, panelActiveStyle.Width(innerWidth).Render(strings.Join(inputLines, "\n")))

	// Footer
	sections = append(sections, renderChatFooter(m))

	content := strings.Join(sections, "\n")
	if m.width > 60 {
		return frameBorder.Width(maxWidth).Render(content)
	}
	return content
}

func renderChatHeader(tick int) string {
	frames := []string{"◆◇◆", "◇◆◇", "◆◆◇", "◇◇◆"}
	accent := bannerAccentStyle.Render(frames[tick%len(frames)])

	title := bannerStyle.Render(" COMMANDER ")
	sub := subtitleStyle.Render("  mission control AI")
	scan := dimSeparatorStyle.Render("  " + strings.Repeat("━", 58))

	return fmt.Sprintf("  %s%s%s\n%s\n%s", accent, title, accent, sub, scan)
}

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

func renderChatFooter(m Model) string {
	keys := []struct{ key, label string }{
		{"esc", "back"},
		{"enter", "send"},
		{"ctrl+v", "voice: " + m.voice.String()},
	}
	if m.voice == voiceManual {
		keys = append(keys, struct{ key, label string }{"ctrl+p", "play"})
	}
	keys = append(keys, struct{ key, label string }{"ctrl+l", "clear"})

	var parts []string
	for _, k := range keys {
		parts = append(parts, footerKeyStyle.Render(k.key)+" "+footerStyle.Render(k.label))
	}
	return "\n  " + strings.Join(parts, "  ")
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
