package tui

import (
	"time"

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
		return m, nil

	case tickMsg:
		if m.pushing {
			// Don't auto-refresh during push, just re-tick
			return m, tickEvery(5 * time.Second)
		}
		return m, tea.Batch(
			refreshCmd(m.repoRoot),
			tickEvery(5*time.Second),
		)

	case pushStepCompleteMsg:
		return handlePushStep(m, msg)

	case pushTickMsg:
		if m.pushing {
			m.spinFrame++
			return m, pushTick()
		}
		return m, nil
	}

	return m, nil
}
