package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/goldenfocus/multitab/internal/git"
)

// pushState tracks the full push sequence.
type pushState struct {
	repoRoot string
	buildCmd string
	steps    []pushStepState
	current  int
	startAt  time.Time
}

type pushStepState struct {
	step    git.PushStep
	status  stepStatus
	elapsed time.Duration
	err     error
}

type stepStatus int

const (
	stepPending stepStatus = iota
	stepRunning
	stepDone
	stepFailed
)

// Messages for the push chain
type pushStartMsg struct{}

type pushStepStartMsg struct {
	step git.PushStep
}

type pushStepCompleteMsg struct {
	step    git.PushStep
	elapsed time.Duration
	err     error
}

type pushTickMsg time.Time

// initPush creates the push state and kicks off the first step.
func initPush(repoRoot, buildCmd string) (pushState, tea.Cmd) {
	steps := []pushStepState{
		{step: git.StepFetch, status: stepPending},
		{step: git.StepRebase, status: stepPending},
		{step: git.StepPush, status: stepPending},
	}

	// Insert build step if configured
	if buildCmd != "" {
		steps = []pushStepState{
			{step: git.StepFetch, status: stepPending},
			{step: git.StepRebase, status: stepPending},
			{step: git.StepBuild, status: stepPending},
			{step: git.StepPush, status: stepPending},
		}
	}

	ps := pushState{
		repoRoot: repoRoot,
		buildCmd: buildCmd,
		steps:    steps,
		current:  0,
		startAt:  time.Now(),
	}

	return ps, tea.Batch(
		runPushStep(ps.repoRoot, ps.buildCmd, ps.steps[0].step),
		pushTick(),
	)
}

// runPushStep executes a single step and returns the result as a message.
func runPushStep(repoRoot, buildCmd string, step git.PushStep) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()

		var err error
		switch step {
		case git.StepFetch:
			err = git.Fetch(repoRoot)
		case git.StepRebase:
			err = git.Rebase(repoRoot)
		case git.StepBuild:
			err = git.RunBuild(repoRoot, buildCmd)
		case git.StepPush:
			err = git.Push(repoRoot)
		}

		return pushStepCompleteMsg{
			step:    step,
			elapsed: time.Since(start),
			err:     err,
		}
	}
}

// pushTick sends periodic ticks to animate the spinner during push.
func pushTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return pushTickMsg(t)
	})
}

// handlePushStep processes a completed step and chains the next one.
func handlePushStep(m Model, msg pushStepCompleteMsg) (Model, tea.Cmd) {
	ps := m.push

	// Update the current step
	ps.steps[ps.current].elapsed = msg.elapsed

	if msg.err != nil {
		ps.steps[ps.current].status = stepFailed
		ps.steps[ps.current].err = msg.err
		m.push = ps
		m.pushing = false
		m.pushErr = msg.err
		return m, nil
	}

	ps.steps[ps.current].status = stepDone

	// Move to next step
	ps.current++
	if ps.current >= len(ps.steps) {
		// All done
		m.push = ps
		m.pushing = false
		m.pushDone = true
		m.pushElapsed = time.Since(ps.startAt)
		return m, refreshCmd(m.repoRoot)
	}

	// Start next step
	ps.steps[ps.current].status = stepRunning
	m.push = ps
	return m, tea.Batch(
		runPushStep(ps.repoRoot, ps.buildCmd, ps.steps[ps.current].step),
		pushTick(),
	)
}
