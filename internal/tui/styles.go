package tui

import "github.com/charmbracelet/lipgloss"

// ─────────────────────────────────────────────────
// Thin rounded border — modern, lightweight
// ─────────────────────────────────────────────────

var thinRoundBorder = lipgloss.Border{
	Top:         "─",
	Bottom:      "─",
	Left:        "│",
	Right:       "│",
	TopLeft:     "╭",
	TopRight:    "╮",
	BottomLeft:  "╰",
	BottomRight: "╯",
}

var (
	// ── Palette (matches multitab.io website) ────
	cyan   = lipgloss.Color("#00f0ff")
	violet = lipgloss.Color("#8b5cf6")
	pink   = lipgloss.Color("#f472b6")
	green  = lipgloss.Color("#4ade80")
	yellow = lipgloss.Color("#fbbf24")
	red    = lipgloss.Color("#f87171")
	orange = lipgloss.Color("#fb923c")

	// Grays — graduated for depth
	bgDark    = lipgloss.Color("#0a0a14")
	dimGray   = lipgloss.Color("#1e1e2e")
	midGray   = lipgloss.Color("#475569")
	lightGray = lipgloss.Color("#94a3b8")
	softWhite = lipgloss.Color("#c8d6e5")
	white     = lipgloss.Color("#e2e8f0")

	// ── Outer frame ─────────────────────────────
	frameBorder = lipgloss.NewStyle().
			Border(thinRoundBorder).
			BorderForeground(lipgloss.Color("#1a2a3a")).
			Padding(1, 2)

	// ── Banner ──────────────────────────────────
	bannerStyle = lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true)

	bannerAccentStyle = lipgloss.NewStyle().
				Foreground(violet).
				Bold(true)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(midGray)

	// ── Section titles ──────────────────────────
	sectionTitleStyle = lipgloss.NewStyle().
				Foreground(cyan).
				Bold(true)

	headerStyle = lipgloss.NewStyle().
			Foreground(midGray).
			Bold(true)

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1e293b"))

	cursorStyle = lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true)

	// ── Agent table ─────────────────────────────
	agentNameStyle = lipgloss.NewStyle().
			Width(24).
			Foreground(softWhite)

	agentNameSelectedStyle = lipgloss.NewStyle().
				Width(24).
				Foreground(cyan).
				Bold(true)

	statusStagedStyle = lipgloss.NewStyle().
				Foreground(green).
				Bold(true)

	statusWorkingStyle = lipgloss.NewStyle().
				Foreground(violet).
				Bold(true)

	statusStaleStyle = lipgloss.NewStyle().
				Foreground(orange)

	statusAbandonedStyle = lipgloss.NewStyle().
				Foreground(red)

	statusIdleStyle = lipgloss.NewStyle().
			Foreground(midGray)

	// ── Panels ──────────────────────────────────
	panelStyle = lipgloss.NewStyle().
			Border(thinRoundBorder).
			BorderForeground(lipgloss.Color("#1e293b")).
			Padding(0, 1)

	panelActiveStyle = lipgloss.NewStyle().
				Border(thinRoundBorder).
				BorderForeground(lipgloss.Color("#1e3a4a")).
				Padding(0, 1)

	panelTitleStyle = lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true)

	commitStyle = lipgloss.NewStyle().
			Foreground(lightGray)

	commitHashStyle = lipgloss.NewStyle().
			Foreground(violet)

	// ── Intel ───────────────────────────────────
	intelLabelStyle = lipgloss.NewStyle().
			Foreground(midGray).
			Width(14)

	intelValueStyle = lipgloss.NewStyle().
			Foreground(softWhite)

	intelFileStyle = lipgloss.NewStyle().
			Foreground(lightGray)

	intelHeaderStyle = lipgloss.NewStyle().
				Foreground(pink).
				Bold(true)

	// ── Status bar ──────────────────────────────
	statusOkStyle = lipgloss.NewStyle().
			Foreground(green)

	statusWarnStyle = lipgloss.NewStyle().
			Foreground(yellow)

	statusIndicatorStyle = lipgloss.NewStyle().
				Foreground(midGray)

	// ── Spawn prompt ────────────────────────────
	spawnPromptStyle = lipgloss.NewStyle().
				Foreground(cyan).
				Bold(true)

	spawnHintStyle = lipgloss.NewStyle().
			Foreground(midGray)

	// ── Footer ──────────────────────────────────
	footerStyle = lipgloss.NewStyle().
			Foreground(midGray)

	footerKeyStyle = lipgloss.NewStyle().
			Foreground(violet).
			Bold(true)

	// ── Push progress ───────────────────────────
	pushStepActiveStyle = lipgloss.NewStyle().
				Foreground(cyan).
				Bold(true)

	pushStepDoneStyle = lipgloss.NewStyle().
				Foreground(green)

	pushStepPendingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#1e293b"))

	// ── Errors ──────────────────────────────────
	errorStyle = lipgloss.NewStyle().
			Foreground(red).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(green).
			Bold(true)

	// ── Status bar (bottom) ─────────────────────
	statusBarStyle = lipgloss.NewStyle().
			Foreground(midGray).
			Background(lipgloss.Color("#0e0e1a")).
			Padding(0, 1)

	statusBarActiveStyle = lipgloss.NewStyle().
				Foreground(green).
				Background(lipgloss.Color("#0e0e1a"))

	statusBarDimStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#2a2a3a")).
				Background(lipgloss.Color("#0e0e1a"))

	// ── Queue bar styles ────────────────────────
	queueFilledStyle = lipgloss.NewStyle().
				Foreground(cyan)

	queueEmptyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1a1a2e"))

	queueShimmerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#2a3a4a"))
)
