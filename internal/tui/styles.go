package tui

import "github.com/charmbracelet/lipgloss"

var (
	// ── Palette ──────────────────────────────────────────────
	pink       = lipgloss.Color("#e740a9")
	cyan       = lipgloss.Color("#00d4ff")
	green      = lipgloss.Color("#00ff88")
	yellow     = lipgloss.Color("#ffcc00")
	red        = lipgloss.Color("#ff4444")
	orange     = lipgloss.Color("#ff8844")
	dimGray    = lipgloss.Color("#444444")
	midGray    = lipgloss.Color("#666666")
	lightGray  = lipgloss.Color("#999999")
	brightGray = lipgloss.Color("#cccccc")
	white      = lipgloss.Color("#ffffff")
	deepBlue   = lipgloss.Color("#0a0e27")
	accentBlue = lipgloss.Color("#1a3a5c")

	// ── Outer frame ──────────────────────────────────────────
	frameBorder = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(cyan).
			Padding(0, 2).
			PaddingTop(1).
			PaddingBottom(1)

	// ── Title / banner ───────────────────────────────────────
	bannerStyle = lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(midGray).
			Italic(true)

	// ── Agent table ──────────────────────────────────────────
	agentNameStyle = lipgloss.NewStyle().
			Width(24).
			Foreground(white)

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
			Foreground(orange).
			Bold(true)

	statusAbandonedStyle = lipgloss.NewStyle().
				Foreground(red)

	statusIdleStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	headerStyle = lipgloss.NewStyle().
			Foreground(midGray).
			Bold(true)

	separatorStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	selectedRowStyle = lipgloss.NewStyle().
			Foreground(cyan)

	cursorStyle = lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true)

	// ── Panels ───────────────────────────────────────────────
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dimGray).
			Padding(0, 1)

	panelActiveStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cyan).
			Padding(0, 1)

	panelTitleStyle = lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true)

	commitStyle = lipgloss.NewStyle().
			Foreground(lightGray)

	commitHashStyle = lipgloss.NewStyle().
			Foreground(yellow)

	// ── Intel panel ──────────────────────────────────────────
	intelLabelStyle = lipgloss.NewStyle().
			Foreground(midGray).
			Width(12)

	intelValueStyle = lipgloss.NewStyle().
			Foreground(brightGray)

	intelFileStyle = lipgloss.NewStyle().
			Foreground(lightGray)

	intelHeaderStyle = lipgloss.NewStyle().
			Foreground(pink).
			Bold(true)

	// ── Status bar ───────────────────────────────────────────
	statusOkStyle = lipgloss.NewStyle().
			Foreground(green)

	statusWarnStyle = lipgloss.NewStyle().
			Foreground(yellow)

	statusIndicatorStyle = lipgloss.NewStyle().
				Foreground(dimGray)

	// ── Footer ───────────────────────────────────────────────
	footerStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	footerKeyStyle = lipgloss.NewStyle().
			Foreground(pink).
			Bold(true)

	// ── Push progress ────────────────────────────────────────
	pushStepActiveStyle = lipgloss.NewStyle().
				Foreground(cyan).
				Bold(true)

	pushStepDoneStyle = lipgloss.NewStyle().
				Foreground(green)

	pushStepPendingStyle = lipgloss.NewStyle().
				Foreground(dimGray)

	// ── Errors ───────────────────────────────────────────────
	errorStyle = lipgloss.NewStyle().
			Foreground(red).
			Bold(true)
)
