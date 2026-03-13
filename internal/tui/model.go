package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/vibeyang/multitab/internal/queue"
)

// Model holds all TUI state.
type Model struct {
	repoRoot string
	buildCmd string
	state    *queue.State
	err      error
	width    int
	height   int
	quitting bool

	// Push state
	pushing     bool
	push        pushState
	pushErr     error
	pushDone    bool
	pushElapsed time.Duration
	spinFrame   int // animation frame counter

	// View toggles
	showDiff bool
	showLog  bool
}

// Messages
type refreshMsg struct {
	state *queue.State
	err   error
}

type tickMsg time.Time

// NewModel creates the initial model.
func NewModel(repoRoot, buildCmd string) Model {
	return Model{
		repoRoot: repoRoot,
		buildCmd: buildCmd,
	}
}

// Init starts the first refresh and the tick timer.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		refreshCmd(m.repoRoot),
		tickEvery(5*time.Second),
	)
}

func refreshCmd(repoRoot string) tea.Cmd {
	return func() tea.Msg {
		state, err := queue.Refresh(repoRoot)
		return refreshMsg{state: state, err: err}
	}
}

func tickEvery(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
