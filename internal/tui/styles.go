package tui

import "github.com/charmbracelet/lipgloss"

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Heavy border set — the spaceship signature
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

var heavyBorder = lipgloss.Border{
	Top:         "\u2501",
	Bottom:      "\u2501",
	Left:        "\u2503",
	Right:       "\u2503",
	TopLeft:     "\u250f",
	TopRight:    "\u2513",
	BottomLeft:  "\u2517",
	BottomRight: "\u251b",
}

var heavyRoundBorder = lipgloss.Border{
	Top:         "\u2501",
	Bottom:      "\u2501",
	Left:        "\u2503",
	Right:       "\u2503",
	TopLeft:     "\u256d",
	TopRight:    "\u256e",
	BottomLeft:  "\u2570",
	BottomRight: "\u256f",
}

var (
	// ── Palette ──────────────────────────────────────
	pink       = lipgloss.Color("#e740a9")
	hotPink    = lipgloss.Color("#ff1493")
	cyan       = lipgloss.Color("#00d4ff")
	green      = lipgloss.Color("#00ff88")
	yellow     = lipgloss.Color("#ffcc00")
	red        = lipgloss.Color("#ff4444")
	orange     = lipgloss.Color("#ff8844")
	dimGray    = lipgloss.Color("#333333")
	midGray    = lipgloss.Color("#555555")
	lightGray  = lipgloss.Color("#888888")
	brightGray = lipgloss.Color("#cccccc")
	white      = lipgloss.Color("#ffffff")

	// ── Outer frame (the ship hull) ──────────────────
	frameBorder = lipgloss.NewStyle().
			Border(heavyBorder).
			BorderForeground(cyan).
			Padding(1, 2)

	// ── Banner ───────────────────────────────────────
	bannerStyle = lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true)

	bannerAccentStyle = lipgloss.NewStyle().
				Foreground(hotPink).
				Bold(true)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(midGray)

	// ── Agent table ──────────────────────────────────
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
			Foreground(lightGray).
			Bold(true)

	sectionTitleStyle = lipgloss.NewStyle().
				Foreground(cyan).
				Bold(true)

	separatorStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	cursorStyle = lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true)

	// ── Panels ───────────────────────────────────────
	panelStyle = lipgloss.NewStyle().
			Border(heavyRoundBorder).
			BorderForeground(dimGray).
			Padding(0, 1)

	panelActiveStyle = lipgloss.NewStyle().
			Border(heavyRoundBorder).
			BorderForeground(cyan).
			Padding(0, 1)

	panelTitleStyle = lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true)

	commitStyle = lipgloss.NewStyle().
			Foreground(lightGray)

	commitHashStyle = lipgloss.NewStyle().
			Foreground(yellow)

	// ── Intel ────────────────────────────────────────
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

	// ── Status bar ───────────────────────────────────
	statusOkStyle = lipgloss.NewStyle().
			Foreground(green)

	statusWarnStyle = lipgloss.NewStyle().
			Foreground(yellow)

	statusIndicatorStyle = lipgloss.NewStyle().
				Foreground(dimGray)

	// ── Spawn prompt ─────────────────────────────────
	spawnPromptStyle = lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true)

	spawnHintStyle = lipgloss.NewStyle().
			Foreground(midGray)

	// ── Footer ───────────────────────────────────────
	footerStyle = lipgloss.NewStyle().
			Foreground(midGray)

	footerKeyStyle = lipgloss.NewStyle().
			Foreground(pink).
			Bold(true)

	// ── Push progress ────────────────────────────────
	pushStepActiveStyle = lipgloss.NewStyle().
				Foreground(cyan).
				Bold(true)

	pushStepDoneStyle = lipgloss.NewStyle().
				Foreground(green)

	pushStepPendingStyle = lipgloss.NewStyle().
				Foreground(dimGray)

	// ── Errors ───────────────────────────────────────
	errorStyle = lipgloss.NewStyle().
			Foreground(red).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(green).
			Bold(true)
)
