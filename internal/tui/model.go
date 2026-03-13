package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/goldenfocus/multitab/internal/queue"
)

// viewMode controls what the main panel shows.
type viewMode int

const (
	viewDashboard viewMode = iota
	viewIntel
	viewSpawn
	viewLog      // scrollable log viewer
	viewPlayback // conversation replay — step through turns
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
	cursor int
	mode   viewMode

	// Spawn input
	promptInput textinput.Model
	spawnErr    error
	spawnOk     string

	// Log viewer
	viewport   viewport.Model
	logContent string

	// Playback (conversation replay)
	chatTurns    []ChatTurn // parsed conversation turns
	chatTurnIdx  int        // current turn being viewed
	chatAgentIdx int        // which agent's chat we're viewing

	// Push state
	pushing     bool
	push        pushState
	pushErr     error
	pushDone    bool
	pushElapsed time.Duration
	spinFrame   int

	// Ambient animation
	tick int
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

type stageResultMsg struct {
	name string
	err  error
}

type killResultMsg struct {
	name string
	err  error
}

type spawnResultMsg struct {
	name string
	err  error
}

// ChatTurn represents one human↔assistant exchange.
type ChatTurn struct {
	HumanText    string
	HumanTime    string
	AssistantText string
	AssistantTime string
	TurnNumber   int
}

type chatTurnsMsg struct {
	turns []ChatTurn
	err   error
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

func ambientTick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
