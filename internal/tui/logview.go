package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// logContent is fetched asynchronously to avoid blocking the TUI.
type logContentMsg struct {
	content string
	err     error
}

// fetchLogCmd reads the log file for the given agent.
func fetchLogCmd(agentPath string) tea.Cmd {
	return func() tea.Msg {
		logFile := agentPath + ".log"

		// Also check inside the worktree for common log locations
		candidates := []string{
			logFile,
			filepath.Join(agentPath, ".claude", "output.log"),
		}

		for _, path := range candidates {
			data, err := os.ReadFile(path)
			if err == nil {
				content := string(data)
				if content == "" {
					content = "(log file exists but is empty — agent may still be starting)"
				}
				return logContentMsg{content: content}
			}
		}

		// No log file — show git log instead as fallback
		return logContentMsg{
			content: "(no log file found — agent was not spawned by multitab)",
		}
	}
}

// autoRefreshLogCmd refreshes the log every 2 seconds.
type logRefreshTickMsg time.Time

func logRefreshTick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return logRefreshTickMsg(t)
	})
}

// initViewport creates or resets the viewport with new content.
func initViewport(content string, width, height int) viewport.Model {
	vp := viewport.New(maxInt(width-8, 40), maxInt(height-10, 10))
	vp.SetContent(content)
	vp.GotoBottom()
	return vp
}

// renderLogView renders the full log viewer.
func renderLogView(m Model) string {
	if m.state == nil || m.cursor >= len(m.state.Agents) {
		return ""
	}

	agent := m.state.Agents[m.cursor]
	var sections []string

	// Header
	sections = append(sections, renderBanner(m.tick))

	// Agent info bar
	icon := agentIcon(agent.Status, m.tick)
	badge := renderStatusBadge(agent.Status, m.tick)
	agentHeader := fmt.Sprintf("  %s %s  %s",
		icon, intelHeaderStyle.Render(agent.Name), badge)
	sections = append(sections, "\n"+agentHeader)

	// Viewport with log content
	vpContent := m.viewport.View()
	sections = append(sections, "\n"+panelStyle.Render(vpContent))

	// Scroll indicator
	pct := m.viewport.ScrollPercent()
	scrollInfo := statusIndicatorStyle.Render(
		fmt.Sprintf("  %d%% │ %d lines", int(pct*100), strings.Count(m.logContent, "\n")+1))
	sections = append(sections, scrollInfo)

	// Footer
	keys := []struct{ key, label string }{
		{"esc", "back"},
		{"↑↓", "scroll"},
		{"g/G", "top/bottom"},
		{"q", "quit"},
	}
	var parts []string
	for _, k := range keys {
		parts = append(parts, footerKeyStyle.Render(k.key)+" "+footerStyle.Render(k.label))
	}
	sections = append(sections, "\n  "+strings.Join(parts, "  "))

	content := strings.Join(sections, "\n")
	if m.width > 60 {
		maxWidth := clampInt(m.width-4, 60, 80)
		return frameBorder.Width(maxWidth).Render(content)
	}
	return content
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
