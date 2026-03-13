package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles all events and returns the new model + commands.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		newM, cmd := handleKeypress(m, msg)
		return newM, cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case refreshMsg:
		m.state = msg.state
		m.err = msg.err
		// Clamp cursor
		if m.state != nil && m.cursor >= len(m.state.Agents) {
			m.cursor = max(0, len(m.state.Agents)-1)
		}
		return m, nil

	case tickMsg:
		m.tick++
		var cmds []tea.Cmd
		cmds = append(cmds, ambientTick())
		// Refresh every 10 ticks (5 seconds at 500ms interval)
		if !m.pushing && m.tick%10 == 0 {
			cmds = append(cmds, refreshCmd(m.repoRoot))
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

	case spawnResultMsg:
		if msg.err != nil {
			m.spawnErr = msg.err
			m.spawnOk = ""
		} else {
			m.spawnOk = fmt.Sprintf("Agent %q launched", msg.name)
			m.spawnErr = nil
		}
		return m, refreshCmd(m.repoRoot)
	}

	// Forward all unhandled messages to textinput when in spawn mode
	if m.mode == viewSpawn {
		var cmd tea.Cmd
		m.promptInput, cmd = m.promptInput.Update(msg)
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
