package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles all events and returns the new model + commands.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return handleKeypress(m, msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.mode == viewLog || m.mode == viewPlayback || m.mode == viewIntel {
			m.viewport.Width = maxInt(m.width-8, 40)
			m.viewport.Height = maxInt(m.height-10, 10)
		}
		return m, nil

	case refreshMsg:
		m.state = msg.state
		m.err = msg.err
		if m.state != nil && m.cursor >= len(m.state.Agents) {
			m.cursor = max(0, len(m.state.Agents)-1)
		}
		// Refresh intel viewport content if we're in intel view
		if m.mode == viewIntel {
			m = refreshIntelViewport(m)
		}
		return m, nil

	case tickMsg:
		m.tick++
		var cmds []tea.Cmd
		cmds = append(cmds, ambientTick())
		if !m.pushing && m.tick%10 == 0 {
			cmds = append(cmds, refreshCmd(m.repoRoot))
		}
		// Auto-clear feedback messages after ~6 seconds (12 ticks)
		if (m.spawnOk != "" || m.spawnErr != nil) && m.feedbackAt > 0 && m.tick-m.feedbackAt > 12 {
			m.spawnOk = ""
			m.spawnErr = nil
		}
		return m, tea.Batch(cmds...)

	case pushStepCompleteMsg:
		return handlePushStep(m, msg)

	case pushTickMsg:
		if m.pushing {
			m.spinFrame++
			return m, pushTick()
		}
		return m, nil

	case discardResultMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		m.mode = viewDashboard
		return m, refreshCmd(m.repoRoot)

	case stageResultMsg:
		if msg.err != nil {
			m.spawnOk = ""
			m.spawnErr = fmt.Errorf("stage %s: %v", msg.name, msg.err)
		} else {
			m.spawnOk = fmt.Sprintf("Staged %q \u2192 local main", msg.name)
			m.spawnErr = nil
		}
		m.feedbackAt = m.tick
		m.mode = viewDashboard
		return m, refreshCmd(m.repoRoot)

	case killResultMsg:
		if msg.err != nil {
			m.spawnOk = ""
			m.spawnErr = fmt.Errorf("kill %s: %v", msg.name, msg.err)
		} else {
			m.spawnOk = fmt.Sprintf("Killed %q", msg.name)
			m.spawnErr = nil
		}
		m.feedbackAt = m.tick
		m.mode = viewDashboard
		return m, refreshCmd(m.repoRoot)

	case spawnResultMsg:
		if msg.err != nil {
			m.spawnErr = msg.err
			m.spawnOk = ""
		} else {
			m.spawnOk = fmt.Sprintf("Agent %q launched", msg.name)
			m.spawnErr = nil
		}
		m.feedbackAt = m.tick
		return m, refreshCmd(m.repoRoot)

	case logContentMsg:
		if msg.err != nil {
			m.logContent = fmt.Sprintf("Error reading log: %v", msg.err)
		} else {
			m.logContent = msg.content
		}
		m.viewport.SetContent(m.logContent)
		m.viewport.GotoBottom()
		return m, nil

	case chatTurnsMsg:
		if msg.err != nil || len(msg.turns) == 0 {
			// No conversation JSONL found — show in log view with helpful message
			m.mode = viewLog
			m.logContent = "No conversation history found for this agent.\n\n" +
				"This agent may have been created manually (not via Claude Code),\n" +
				"or the session JSONL has been cleaned up.\n\n" +
				"Press [esc] to go back."
			m.viewport = initViewport(m.logContent, m.width, m.height)
			return m, nil
		}
		m.chatTurns = msg.turns
		m.chatTurnIdx = len(msg.turns) - 1 // start at the LAST turn (most recent)
		m.mode = viewPlayback
		content := formatTurnContent(m.chatTurns[m.chatTurnIdx])
		m.logContent = content
		m.viewport = initViewport(content, m.width, m.height)
		m.viewport.GotoTop()
		return m, nil

	case logRefreshTickMsg:
		// Auto-refresh log every 2s while in log view
		if m.mode == viewLog && m.state != nil && m.cursor < len(m.state.Agents) {
			a := m.state.Agents[m.cursor]
			return m, tea.Batch(
				fetchLogCmd(a.Path),
				logRefreshTick(),
			)
		}
		return m, nil
	}

	// Forward to textinput in spawn mode
	if m.mode == viewSpawn {
		var cmd tea.Cmd
		m.promptInput, cmd = m.promptInput.Update(msg)
		return m, cmd
	}

	// Forward to viewport in log mode
	if m.mode == viewLog {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
