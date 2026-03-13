package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	agentpkg "github.com/goldenfocus/multitab/internal/agent"
	"github.com/goldenfocus/multitab/internal/git"
	"github.com/goldenfocus/multitab/internal/queue"
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

	// Playback view handles turn navigation
	if m.mode == viewPlayback {
		return handlePlaybackKeys(m, msg)
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
		m.cursor = prevActiveAgent(m.state, m.cursor)
	case "down", "j":
		m.cursor = nextActiveAgent(m.state, m.cursor)
	case "enter", "right":
		if agentCount > 0 {
			m.mode = viewIntel
			m = refreshIntelViewport(m)
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
	case "c", "C":
		// Conversation playback from dashboard
		if agentCount > 0 && m.state != nil {
			m.chatAgentIdx = m.cursor
			return m, fetchTurnsCmd(m.state.Agents[m.cursor].Path)
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
	case "esc", "backspace":
		m.mode = viewDashboard
		return m, nil
	case "left":
		m.mode = viewDashboard
		return m, nil
	case "right", "enter":
		// Drill deeper — open conversation replay
		if m.state != nil && m.cursor < len(m.state.Agents) {
			m.chatAgentIdx = m.cursor
			return m, fetchTurnsCmd(m.state.Agents[m.cursor].Path)
		}
		return m, nil
	case "tab", "]":
		// Next agent (stay in intel view)
		if m.state != nil && m.cursor < len(m.state.Agents)-1 {
			m.cursor++
			m = refreshIntelViewport(m)
		}
		return m, nil
	case "shift+tab", "[":
		// Previous agent (stay in intel view)
		if m.cursor > 0 {
			m.cursor--
			m = refreshIntelViewport(m)
		}
		return m, nil
	case "l", "L", "o", "O":
		// Jump into log view (raw .log file from spawned agents)
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
	case "c", "C":
		// Conversation playback from intel view
		if m.state != nil && m.cursor < len(m.state.Agents) {
			m.chatAgentIdx = m.cursor
			return m, fetchTurnsCmd(m.state.Agents[m.cursor].Path)
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

// refreshIntelViewport renders the intel panel content into the viewport.
func refreshIntelViewport(m Model) Model {
	if m.state == nil || m.cursor >= len(m.state.Agents) {
		return m
	}
	agent := m.state.Agents[m.cursor]
	content := renderIntelContent(agent, m.tick)
	m.logContent = content
	m.viewport = initViewport(content, m.width, m.height)
	m.viewport.GotoTop()
	return m
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

// isDormant returns true if an agent is idle with zero activity.
func isDormant(a git.Agent) bool {
	return a.Status == git.StatusIdle && a.Commits == 0 && a.Files == 0 && a.DirtyFiles == 0
}

// nextActiveAgent finds the next non-dormant agent index, or the next index if all are dormant.
func nextActiveAgent(state *queue.State, current int) int {
	if state == nil || len(state.Agents) == 0 {
		return current
	}
	n := len(state.Agents)
	// Try to find next active agent
	for i := current + 1; i < n; i++ {
		if !isDormant(state.Agents[i]) {
			return i
		}
	}
	// No more active agents below — just go to next if any
	if current < n-1 {
		return current + 1
	}
	return current
}

// prevActiveAgent finds the previous non-dormant agent index.
func prevActiveAgent(state *queue.State, current int) int {
	if state == nil || len(state.Agents) == 0 {
		return current
	}
	// Try to find prev active agent
	for i := current - 1; i >= 0; i-- {
		if !isDormant(state.Agents[i]) {
			return i
		}
	}
	// No more active agents above — just go to prev if any
	if current > 0 {
		return current - 1
	}
	return current
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

func handlePlaybackKeys(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "backspace", "q":
		m.mode = viewIntel
		return m, nil
	case "left", "h":
		// Previous turn
		if m.chatTurnIdx > 0 {
			m.chatTurnIdx--
			content := formatTurnContent(m.chatTurns[m.chatTurnIdx])
			m.logContent = content
			m.viewport.SetContent(content)
			m.viewport.GotoTop()
		}
		return m, nil
	case "right", "l":
		// Next turn
		if m.chatTurnIdx < len(m.chatTurns)-1 {
			m.chatTurnIdx++
			content := formatTurnContent(m.chatTurns[m.chatTurnIdx])
			m.logContent = content
			m.viewport.SetContent(content)
			m.viewport.GotoTop()
		}
		return m, nil
	case "g":
		m.viewport.GotoTop()
		return m, nil
	case "G":
		m.viewport.GotoBottom()
		return m, nil
	case "home":
		// Jump to first turn
		m.chatTurnIdx = 0
		content := formatTurnContent(m.chatTurns[0])
		m.logContent = content
		m.viewport.SetContent(content)
		m.viewport.GotoTop()
		return m, nil
	case "end":
		// Jump to last turn
		m.chatTurnIdx = len(m.chatTurns) - 1
		content := formatTurnContent(m.chatTurns[m.chatTurnIdx])
		m.logContent = content
		m.viewport.SetContent(content)
		m.viewport.GotoTop()
		return m, nil
	}

	// Forward to viewport for scrolling (up/down/pgup/pgdn)
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func fetchTurnsCmd(agentPath string) tea.Cmd {
	return func() tea.Msg {
		turns, err := agentpkg.ParseTurns(agentPath)
		if err != nil || len(turns) == 0 {
			return chatTurnsMsg{err: err}
		}
		// Convert agent turns to TUI turns
		var chatTurns []ChatTurn
		for _, t := range turns {
			chatTurns = append(chatTurns, ChatTurn{
				HumanText:     t.HumanText,
				HumanTime:     t.HumanTime,
				AssistantText: t.AssistantText,
				AssistantTime: t.AssistantTime,
				TurnNumber:    t.Number,
			})
		}
		return chatTurnsMsg{turns: chatTurns}
	}
}

func fetchChatCmd(agentPath string) tea.Cmd {
	return func() tea.Msg {
		content, err := agentpkg.ReadFullChat(agentPath)
		if err != nil {
			return logContentMsg{content: "(no conversation found)", err: nil}
		}
		return logContentMsg{content: content}
	}
}

func spawnAgentCmd(repoRoot, prompt string) tea.Cmd {
	return func() tea.Msg {
		result := agentpkg.Spawn(repoRoot, prompt)
		if result.Err != nil {
			return spawnResultMsg{name: result.Name, err: result.Err}
		}
		return spawnResultMsg{name: result.Name}
	}
}
