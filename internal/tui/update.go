package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vibeyang/multitab/internal/git"
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
			return m, tickEvery(5 * time.Second)
		}
		return m, tea.Batch(
			refreshCmd(m.repoRoot),
			tickEvery(5*time.Second),
		)

	case pushStepMsg:
		if msg.err != nil {
			m.pushErr = msg.err
			m.pushing = false
			return m, nil
		}
		m.pushStep = msg.step
		if msg.step == git.StepDone {
			m.pushing = false
			m.pushDone = true
			// Refresh after push completes
			return m, refreshCmd(m.repoRoot)
		}
		return m, nil
	}

	return m, nil
}

// startPush kicks off the push sequence as a series of commands.
func startPush(repoRoot, buildCmd string) tea.Cmd {
	return func() tea.Msg {
		// Step 1: Fetch
		if err := git.Fetch(repoRoot); err != nil {
			return pushStepMsg{step: git.StepFetch, err: err}
		}

		// Step 2: Rebase
		if err := git.Rebase(repoRoot); err != nil {
			return pushStepMsg{step: git.StepRebase, err: err}
		}

		// Step 3: Build
		if buildCmd != "" {
			if err := git.RunBuild(repoRoot, buildCmd); err != nil {
				return pushStepMsg{step: git.StepBuild, err: err}
			}
		}

		// Step 4: Push
		if err := git.Push(repoRoot); err != nil {
			return pushStepMsg{step: git.StepPush, err: err}
		}

		return pushStepMsg{step: git.StepDone}
	}
}
