package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/goldenfocus/multitab/internal/git"
)

// handleKeypress routes key events to actions.
func handleKeypress(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	// Global keys (work in any mode)
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "r", "R":
		if !m.pushing {
			return m, refreshCmd(m.repoRoot)
		}
	}

	// Mode-specific keys
	switch m.mode {
	case viewDashboard:
		return handleDashboardKeys(m, msg)
	case viewIntel:
		return handleIntelKeys(m, msg)
	}

	return m, nil
}

func handleDashboardKeys(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	agentCount := 0
	if m.state != nil {
		agentCount = len(m.state.Agents)
	}

	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < agentCount-1 {
			m.cursor++
		}
	case "enter":
		if agentCount > 0 {
			m.mode = viewIntel
		}
	case "p", "P":
		if !m.pushing && m.state != nil && len(m.state.StagedCommits) > 0 {
			m.pushing = true
			m.pushDone = false
			m.pushErr = nil
			ps, cmd := initPush(m.state.RepoRoot, m.buildCmd)
			ps.steps[0].status = stepRunning
			m.push = ps
			return m, cmd
		}
	}
	return m, nil
}

func handleIntelKeys(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "backspace", "left", "h":
		m.mode = viewDashboard
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.state != nil && m.cursor < len(m.state.Agents)-1 {
			m.cursor++
		}
	case "x", "X":
		// Discard selected agent
		if m.state != nil && m.cursor < len(m.state.Agents) {
			agent := m.state.Agents[m.cursor]
			return m, discardAgent(m.repoRoot, agent)
		}
	}
	return m, nil
}

func discardAgent(repoRoot string, agent git.Agent) tea.Cmd {
	return func() tea.Msg {
		err := git.CleanupWorktree(repoRoot, agent.Path, agent.Branch)
		return discardResultMsg{err: err}
	}
}
