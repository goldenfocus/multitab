package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/vibeyang/multitab/internal/git"
)

var spinnerFrames = []string{"\u280b", "\u2819", "\u2839", "\u2838", "\u283c", "\u2834", "\u2826", "\u2827", "\u2807", "\u280f"}

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
		sections = append(sections, renderLivePush(m))
	}

	// Push error
	if m.pushErr != nil && !m.pushing {
		sections = append(sections, errorStyle.Render(fmt.Sprintf("  Push failed: %v", m.pushErr)))
	}

	// Push success
	if m.pushDone {
		sections = append(sections,
			statusOkStyle.Render(fmt.Sprintf("  Pushed in %s", m.pushElapsed.Round(time.Millisecond))))
	}

	// Footer
	sections = append(sections, renderFooter(m.pushing))

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
			status = statusStagedStyle.Render(fmt.Sprintf("%-14s", "\u2713 STAGED"))
		case git.StatusWorking:
			status = statusWorkingStyle.Render(fmt.Sprintf("%-14s", "\u2022 WORKING"))
		default:
			status = statusIdleStyle.Render(fmt.Sprintf("%-14s", "\u25cb IDLE"))
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

	filledBar := strings.Repeat("\u2588", filled)
	emptyBar := strings.Repeat("\u2591", barWidth-filled)
	barStyled := statusOkStyle.Render(filledBar) + statusIdleStyle.Render(emptyBar)

	return fmt.Sprintf("\n%s\n  %s  %d%% ready\n",
		headerStyle.Render(label),
		barStyled,
		pct,
	)
}

func renderCommitsPanel(commits []git.StagedCommit) string {
	var lines []string
	for _, c := range commits {
		lines = append(lines, commitStyle.Render("  "+c.Message))
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
			statusWarnStyle.Render(fmt.Sprintf("  CONFLICTS: %d file(s) touched by multiple agents", len(m.state.Conflicts))))
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

func renderLivePush(m Model) string {
	var lines []string

	elapsed := time.Since(m.push.startAt).Round(time.Millisecond)
	spinner := spinnerFrames[m.spinFrame%len(spinnerFrames)]

	for _, s := range m.push.steps {
		var line string
		switch s.status {
		case stepDone:
			line = pushStepDoneStyle.Render(
				fmt.Sprintf("  \u2714 %s  %s", s.step, formatDuration(s.elapsed)))
		case stepRunning:
			line = pushStepActiveStyle.Render(
				fmt.Sprintf("  %s %s", spinner, s.step))
		case stepFailed:
			line = errorStyle.Render(
				fmt.Sprintf("  \u2718 %s  FAILED", s.step))
		default:
			line = pushStepPendingStyle.Render(
				fmt.Sprintf("    %s", s.step))
		}
		lines = append(lines, line)
	}

	header := pushStepActiveStyle.Render(fmt.Sprintf(" PUSHING  %s ", elapsed))

	return "\n" + panelStyle.Render(
		header + "\n" + strings.Join(lines, "\n"),
	)
}

func renderFooter(pushing bool) string {
	type key struct{ key, label string }

	var keys []key
	if pushing {
		keys = []key{
			{"", "pushing..."},
		}
	} else {
		keys = []key{
			{"P", "Push"},
			{"D", "Diff"},
			{"R", "Refresh"},
			{"L", "Log"},
			{"Q", "Quit"},
		}
	}

	var parts []string
	for _, k := range keys {
		if k.key == "" {
			parts = append(parts, pushStepActiveStyle.Render(k.label))
		} else {
			parts = append(parts, "["+footerKeyStyle.Render(k.key)+"]"+k.label)
		}
	}

	return "\n" + footerStyle.Render("  "+strings.Join(parts, "  "))
}

func agentIcon(status git.AgentStatus) string {
	switch status {
	case git.StatusStaged:
		return "\u25cf"
	case git.StatusWorking:
		return "\u25cf"
	default:
		return "\u25cb"
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

// min returns the smaller of two ints.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

