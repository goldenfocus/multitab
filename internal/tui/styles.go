package tui

import "github.com/charmbracelet/lipgloss"

// ─────────────────────────────────────────────────
// Rounded border — visible but not heavy
// ─────────────────────────────────────────────────

var roundBorder = lipgloss.Border{
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
	dimGray   = lipgloss.Color("#2a2a3e")
	midGray   = lipgloss.Color("#64748b")
	lightGray = lipgloss.Color("#94a3b8")
	softWhite = lipgloss.Color("#cbd5e1")
	white     = lipgloss.Color("#e2e8f0")

	// ── Outer frame (the ship hull — VISIBLE) ───
	frameBorder = lipgloss.NewStyle().
			Border(roundBorder).
			BorderForeground(cyan).
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
			Foreground(lipgloss.Color("#334155"))

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
				Foreground(yellow).
				Bold(true)

	statusStaleStyle = lipgloss.NewStyle().
				Foreground(orange)

	statusAbandonedStyle = lipgloss.NewStyle().
				Foreground(red)

	statusIdleStyle = lipgloss.NewStyle().
			Foreground(midGray)

	// ── Panels ──────────────────────────────────
	panelStyle = lipgloss.NewStyle().
			Border(roundBorder).
			BorderForeground(lipgloss.Color("#334155")).
			Padding(0, 1)

	panelActiveStyle = lipgloss.NewStyle().
				Border(roundBorder).
				BorderForeground(cyan).
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
			Foreground(pink).
			Bold(true)

	// ── Push progress ───────────────────────────
	pushStepActiveStyle = lipgloss.NewStyle().
				Foreground(cyan).
				Bold(true)

	pushStepDoneStyle = lipgloss.NewStyle().
				Foreground(green)

	pushStepPendingStyle = lipgloss.NewStyle().
				Foreground(dimGray)

	// ── Errors ──────────────────────────────────
	errorStyle = lipgloss.NewStyle().
			Foreground(red).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(green).
			Bold(true)

	// ── Status bar (bottom instrument panel) ────
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lightGray)

	statusBarActiveStyle = lipgloss.NewStyle().
				Foreground(green).
				Bold(true)

	statusBarDimStyle = lipgloss.NewStyle().
				Foreground(dimGray)

	// ── Queue bar styles ────────────────────────
	queueFilledStyle = lipgloss.NewStyle().
				Foreground(cyan)

	queueEmptyStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	queueShimmerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#3a4a5a"))
)
