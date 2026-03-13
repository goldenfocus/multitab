package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/goldenfocus/multitab/internal/queue"
)

// viewMode controls what the main panel shows.
type viewMode int

const (
	viewDashboard viewMode = iota // default: agent list + queue
	viewIntel                     // expanded intel for selected agent
	viewSpawn                     // prompt input for new agent
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

	// Navigation
	cursor int      // selected agent index
	mode   viewMode // current view

	// Spawn input
	promptInput textinput.Model
	spawnErr    error
	spawnOk     string // success message after spawn

	// Push state
	pushing     bool
	push        pushState
	pushErr     error
	pushDone    bool
	pushElapsed time.Duration
	spinFrame   int // animation frame counter

	// Ambient animation
	tick int // global tick counter for ambient effects
}

// Messages
type refreshMsg struct {
	state *queue.State
	err   error
}

type tickMsg time.Time

type discardResultMsg struct {
	err error
}

type spawnResultMsg struct {
	name string
	err  error
}

// NewModel creates the initial model.
func NewModel(repoRoot, buildCmd string) Model {
	ti := textinput.New()
	ti.Placeholder = "describe the task, or path to .md file..."
	ti.CharLimit = 500
	ti.Width = 56

	return Model{
		repoRoot:    repoRoot,
		buildCmd:    buildCmd,
		mode:        viewDashboard,
		promptInput: ti,
	}
}

// Init starts the first refresh and the ambient tick.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		refreshCmd(m.repoRoot),
		ambientTick(),
	)
}

func refreshCmd(repoRoot string) tea.Cmd {
	return func() tea.Msg {
		state, err := queue.Refresh(repoRoot)
		return refreshMsg{state: state, err: err}
	}
}

// ambientTick fires every 500ms for animations + periodic refresh.
func ambientTick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
