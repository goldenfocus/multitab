package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	pink      = lipgloss.Color("#e740a9")
	cyan      = lipgloss.Color("#00d4ff")
	green     = lipgloss.Color("#00ff88")
	yellow    = lipgloss.Color("#ffcc00")
	red       = lipgloss.Color("#ff4444")
	dimGray   = lipgloss.Color("#555555")
	lightGray = lipgloss.Color("#aaaaaa")
	white     = lipgloss.Color("#ffffff")
	bgDark    = lipgloss.Color("#111111")

	// Border style for the main frame
	frameBorder = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(cyan).
			Padding(1, 2)

	// Title
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(cyan).
			MarginBottom(1)

	// Agent status styles
	agentNameStyle = lipgloss.NewStyle().
			Width(28).
			Foreground(white)

	statusStagedStyle = lipgloss.NewStyle().
				Foreground(green).
				Bold(true)

	statusWorkingStyle = lipgloss.NewStyle().
				Foreground(yellow).
				Bold(true)

	statusIdleStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	// Column header
	headerStyle = lipgloss.NewStyle().
			Foreground(lightGray).
			Bold(true).
			MarginBottom(0)

	// Separator line
	separatorStyle = lipgloss.NewStyle().
			Foreground(dimGray)

	// Panel styles
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dimGray).
			Padding(0, 1)

	panelTitleStyle = lipgloss.NewStyle().
			Foreground(cyan).
			Bold(true)

	commitStyle = lipgloss.NewStyle().
			Foreground(lightGray)

	// Status bar items
	statusOkStyle = lipgloss.NewStyle().
			Foreground(green)

	statusWarnStyle = lipgloss.NewStyle().
			Foreground(yellow)

	statusErrorStyle = lipgloss.NewStyle().
			Foreground(red)

	// Progress bar
	progressFilledStyle = lipgloss.NewStyle().
				Foreground(green).
				Background(green)

	progressEmptyStyle = lipgloss.NewStyle().
				Foreground(dimGray).
				Background(dimGray)

	// Footer
	footerStyle = lipgloss.NewStyle().
			Foreground(dimGray).
			MarginTop(1)

	footerKeyStyle = lipgloss.NewStyle().
			Foreground(pink).
			Bold(true)

	// Push progress
	pushStepActiveStyle = lipgloss.NewStyle().
				Foreground(cyan).
				Bold(true)

	pushStepDoneStyle = lipgloss.NewStyle().
				Foreground(green)

	pushStepPendingStyle = lipgloss.NewStyle().
				Foreground(dimGray)

	// Error display
	errorStyle = lipgloss.NewStyle().
			Foreground(red).
			Bold(true)
)
