package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/goldenfocus/multitab/internal/agent"
	"github.com/goldenfocus/multitab/internal/git"
)

func handleKeypress(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	// Spawn mode handles its own input
	if m.mode == viewSpawn {
		return handleSpawnKeys(m, msg)
	}

	// Log view handles its own scrolling
	if m.mode == viewLog {
		return handleLogKeys(m, msg)
	}

	// Global keys
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "r", "R":
		if !m.pushing {
			return m, refreshCmd(m.repoRoot)
		}
	case "n", "N":
		m.mode = viewSpawn
		m.promptInput.Focus()
		m.promptInput.SetValue("")
		m.spawnErr = nil
		m.spawnOk = ""
		return m, textinput.Blink
	}

	// Mode-specific
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
	case "s", "S":
		// Stage: merge agent's branch into local main
		if agentCount > 0 && m.state != nil {
			a := m.state.Agents[m.cursor]
			if a.Status != git.StatusStaged && a.Commits > 0 {
				return m, stageAgentCmd(m.repoRoot, a)
			}
		}
	case "x", "X":
		// Kill: force-remove worktree + branch
		if agentCount > 0 && m.state != nil {
			a := m.state.Agents[m.cursor]
			return m, killAgentCmd(m.repoRoot, a)
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
	case "l", "L", "o", "O":
		// Jump into log view
		if m.state != nil && m.cursor < len(m.state.Agents) {
			a := m.state.Agents[m.cursor]
			m.mode = viewLog
			m.logContent = ""
			m.viewport = initViewport("loading...", m.width, m.height)
			return m, tea.Batch(
				fetchLogCmd(a.Path),
				logRefreshTick(),
			)
		}
	case "s", "S":
		// Stage: merge agent's branch into local main
		if m.state != nil && m.cursor < len(m.state.Agents) {
			a := m.state.Agents[m.cursor]
			if a.Status != git.StatusStaged && a.Commits > 0 {
				return m, stageAgentCmd(m.repoRoot, a)
			}
		}
	case "x", "X":
		// Kill: force-remove worktree + branch (any status)
		if m.state != nil && m.cursor < len(m.state.Agents) {
			a := m.state.Agents[m.cursor]
			return m, killAgentCmd(m.repoRoot, a)
		}
	}
	return m, nil
}

func handleLogKeys(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "backspace", "q":
		m.mode = viewIntel
		return m, nil
	case "g":
		m.viewport.GotoTop()
		return m, nil
	case "G":
		m.viewport.GotoBottom()
		return m, nil
	}

	// Forward to viewport for scrolling (up/down/pgup/pgdn)
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func handleSpawnKeys(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = viewDashboard
		m.promptInput.Blur()
		return m, nil
	case "enter":
		prompt := m.promptInput.Value()
		if prompt == "" {
			return m, nil
		}
		m.promptInput.Blur()
		m.mode = viewDashboard
		return m, spawnAgentCmd(m.repoRoot, prompt)
	}

	var cmd tea.Cmd
	m.promptInput, cmd = m.promptInput.Update(msg)
	return m, cmd
}

func discardAgentCmd(repoRoot string, a git.Agent) tea.Cmd {
	return func() tea.Msg {
		err := git.CleanupWorktree(repoRoot, a.Path, a.Branch)
		return discardResultMsg{err: err}
	}
}

func stageAgentCmd(repoRoot string, a git.Agent) tea.Cmd {
	return func() tea.Msg {
		err := git.MergeBranch(repoRoot, a.Branch)
		return stageResultMsg{name: a.Name, err: err}
	}
}

func killAgentCmd(repoRoot string, a git.Agent) tea.Cmd {
	return func() tea.Msg {
		err := git.ForceCleanupWorktree(repoRoot, a.Path, a.Branch)
		return killResultMsg{name: a.Name, err: err}
	}
}

func spawnAgentCmd(repoRoot, prompt string) tea.Cmd {
	return func() tea.Msg {
		result := agent.Spawn(repoRoot, prompt)
		if result.Err != nil {
			return spawnResultMsg{name: result.Name, err: result.Err}
		}
		return spawnResultMsg{name: result.Name}
	}
}
