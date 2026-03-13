package tui

import "github.com/charmbracelet/bubbletea"

// handleKeypress routes key events to actions.
func handleKeypress(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "p", "P":
		if !m.pushing && m.state != nil && len(m.state.StagedCommits) > 0 {
			return m, startPush(m.state.RepoRoot, m.buildCmd)
		}
	case "r", "R":
		return m, refreshCmd(m.repoRoot)
	case "d", "D":
		m.showDiff = !m.showDiff
	case "l", "L":
		m.showLog = !m.showLog
	}
	return m, nil
}
