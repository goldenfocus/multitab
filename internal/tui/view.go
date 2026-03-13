package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/vibeyang/multitab/internal/git"
)

// View renders the spaceship dashboard.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	if m.state == nil {
		return titleStyle.Render("  MULTITAB") + "\n\n  Scanning worktrees..."
	}

	var sections []string

	// Title
	sections = append(sections, titleStyle.Render("  MULTITAB"))

	// Agent table
	sections = append(sections, renderAgentTable(m.state.Agents))

	// Deploy queue progress
	sections = append(sections, renderQueueBar(m.state.ReadyCount, m.state.TotalCount))

	// Staged commits panel
	if len(m.state.StagedCommits) > 0 {
		sections = append(sections, renderCommitsPanel(m.state.StagedCommits))
	}

	// Status bar
	sections = append(sections, renderStatusBar(m))

	// Push progress (if pushing)
	if m.pushing {
		sections = append(sections, renderPushProgress(m.pushStep))
	}

	// Push error
	if m.pushErr != nil {
		sections = append(sections, errorStyle.Render(fmt.Sprintf("  Push failed: %v", m.pushErr)))
	}

	// Push success
	if m.pushDone {
		sections = append(sections, statusOkStyle.Render("  Push complete!"))
	}

	// Footer
	sections = append(sections, renderFooter())

	content := strings.Join(sections, "\n")

	// Apply frame border if terminal is wide enough
	if m.width > 60 {
		maxWidth := min(m.width-4, 70)
		return frameBorder.Width(maxWidth).Render(content)
	}
	return content
}

func renderAgentTable(agents []git.Agent) string {
	if len(agents) == 0 {
		return statusIdleStyle.Render("  No worktrees found. Run `multitab init` or create worktrees.")
	}

	var b strings.Builder

	// Header
	header := fmt.Sprintf("  %-28s %-14s %-10s %-6s", "AGENTS", "STATUS", "COMMITS", "FILES")
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(separatorStyle.Render("  " + strings.Repeat("\u2500", 60)))
	b.WriteString("\n")

	// Rows
	for _, agent := range agents {
		icon := agentIcon(agent.Status)
		name := agentNameStyle.Render(icon + " " + agent.Name)

		var status string
		switch agent.Status {
		case git.StatusStaged:
			status = statusStagedStyle.Render(fmt.Sprintf("%-14s", "STAGED"))
		case git.StatusWorking:
			status = statusWorkingStyle.Render(fmt.Sprintf("%-14s", "WORKING"))
		default:
			status = statusIdleStyle.Render(fmt.Sprintf("%-14s", "IDLE"))
		}

		commits := fmt.Sprintf("%-10d", agent.Commits)
		files := fmt.Sprintf("%-6d", agent.Files)

		b.WriteString("  " + name + status + commits + files + "\n")
	}

	return b.String()
}

func renderQueueBar(ready, total int) string {
	if total == 0 {
		return ""
	}

	label := fmt.Sprintf("  DEPLOY QUEUE %d/%d", ready, total)

	barWidth := 40
	filled := 0
	if total > 0 {
		filled = (ready * barWidth) / total
	}
	pct := 0
	if total > 0 {
		pct = (ready * 100) / total
	}

	bar := strings.Repeat("\u2588", filled) + strings.Repeat("\u2591", barWidth-filled)
	barStyled := statusOkStyle.Render(bar[:filled]) + statusIdleStyle.Render(bar[filled:])

	return fmt.Sprintf("\n%s\n  %s  %d%% ready\n",
		headerStyle.Render(label),
		barStyled,
		pct,
	)
}

func renderCommitsPanel(commits []git.StagedCommit) string {
	var lines []string
	for _, c := range commits {
		lines = append(lines, commitStyle.Render(c.Message))
	}
	content := strings.Join(lines, "\n")

	return panelStyle.Render(
		panelTitleStyle.Render(" STAGED COMMITS ") + "\n" + content,
	)
}

func renderStatusBar(m Model) string {
	var items []string

	// Conflicts
	if len(m.state.Conflicts) > 0 {
		items = append(items,
			statusWarnStyle.Render(fmt.Sprintf("  CONFLICTS: %d detected", len(m.state.Conflicts))))
	} else {
		items = append(items,
			statusOkStyle.Render("  CONFLICTS: None detected"))
	}

	// Migrations
	if m.state.HasMigrations {
		items = append(items,
			statusWarnStyle.Render("  MIGRATIONS: Pending"))
	}

	// Last deploy
	if m.state.LastPushHash != "" {
		items = append(items, statusIdleStyle.Render(
			fmt.Sprintf("  LAST DEPLOY: %s \u2014 %s", m.state.LastPushTime, m.state.LastPushHash)))
	}

	return strings.Join(items, "\n")
}

func renderPushProgress(step git.PushStep) string {
	steps := []git.PushStep{git.StepFetch, git.StepRebase, git.StepBuild, git.StepPush, git.StepVerify}
	var lines []string

	for _, s := range steps {
		var style lipgloss.Style
		prefix := "  "
		switch {
		case s < step:
			style = pushStepDoneStyle
			prefix = "  \u2714 "
		case s == step:
			style = pushStepActiveStyle
			prefix = "  \u25b6 "
		default:
			style = pushStepPendingStyle
			prefix = "    "
		}
		lines = append(lines, style.Render(prefix+s.String()))
	}

	return "\n" + panelStyle.Render(
		pushStepActiveStyle.Render(" PUSHING ") + "\n" + strings.Join(lines, "\n"),
	)
}

func renderFooter() string {
	keys := []struct{ key, label string }{
		{"P", "Push"},
		{"D", "Diff"},
		{"R", "Refresh"},
		{"L", "Log"},
		{"Q", "Quit"},
	}

	var parts []string
	for _, k := range keys {
		parts = append(parts, "["+footerKeyStyle.Render(k.key)+"]"+k.label)
	}

	return "\n" + footerStyle.Render("  "+strings.Join(parts, "  "))
}

func agentIcon(status git.AgentStatus) string {
	switch status {
	case git.StatusStaged:
		return "\u25cf" // filled circle
	case git.StatusWorking:
		return "\u25cf" // filled circle
	case git.StatusIdle:
		return "\u25cb" // empty circle
	default:
		return "\u25cb"
	}
}
