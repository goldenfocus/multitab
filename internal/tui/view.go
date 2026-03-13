package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/goldenfocus/multitab/internal/git"
)

var spinnerFrames = []string{"\u280b", "\u2819", "\u2839", "\u2838", "\u283c", "\u2834", "\u2826", "\u2827", "\u2807", "\u280f"}

// Ambient glow frames for the title
var glowFrames = []string{
	"\u2591\u2592\u2593\u2588",
	"\u2592\u2593\u2588\u2593",
	"\u2593\u2588\u2593\u2592",
	"\u2588\u2593\u2592\u2591",
}

const banner = ` _____ _____ _   ____ _____
|     |  |  | | |_   _|  _  |  ___|
| | | |  |  | |_  | | | |_| || ___ \
|_|_|_|_____|___| |_| |_____|_____|`

// View renders the spaceship dashboard.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf(" Error: %v", m.err))
	}

	if m.state == nil {
		frame := glowFrames[m.tick%len(glowFrames)]
		return bannerStyle.Render(banner) + "\n\n  " +
			pushStepActiveStyle.Render(frame+" Scanning worktrees... "+frame)
	}

	switch m.mode {
	case viewIntel:
		return m.renderIntelView()
	default:
		return m.renderDashboardView()
	}
}

func (m Model) renderDashboardView() string {
	var sections []string

	// ── Banner ───────────────────────────────────
	sections = append(sections, renderBanner(m.tick))

	// ── Agent table ──────────────────────────────
	sections = append(sections, renderAgentTable(m.state.Agents, m.cursor, m.tick))

	// ── Deploy queue ─────────────────────────────
	sections = append(sections, renderQueueBar(m.state.ReadyCount, m.state.TotalCount, m.tick))

	// ── Staged commits ───────────────────────────
	if len(m.state.StagedCommits) > 0 {
		sections = append(sections, renderCommitsPanel(m.state.StagedCommits))
	}

	// ── System status ────────────────────────────
	sections = append(sections, renderStatusBar(m))

	// ── Push progress ────────────────────────────
	if m.pushing {
		sections = append(sections, renderLivePush(m))
	}
	if m.pushErr != nil && !m.pushing {
		sections = append(sections, errorStyle.Render(fmt.Sprintf("  Push failed: %v", m.pushErr)))
	}
	if m.pushDone {
		sections = append(sections,
			statusOkStyle.Render(fmt.Sprintf("  \u2714 Deployed in %s", m.pushElapsed.Round(time.Millisecond))))
	}

	// ── Footer ───────────────────────────────────
	sections = append(sections, renderDashboardFooter(m))

	content := strings.Join(sections, "\n")

	if m.width > 60 {
		maxWidth := min(m.width-4, 72)
		return frameBorder.Width(maxWidth).Render(content)
	}
	return content
}

func (m Model) renderIntelView() string {
	if m.state == nil || m.cursor >= len(m.state.Agents) {
		m.mode = viewDashboard
		return m.renderDashboardView()
	}

	agent := m.state.Agents[m.cursor]
	var sections []string

	// ── Header ───────────────────────────────────
	sections = append(sections, renderBanner(m.tick))

	// ── Agent intel panel ────────────────────────
	sections = append(sections, renderIntelPanel(agent, m.tick))

	// ── Footer ───────────────────────────────────
	sections = append(sections, renderIntelFooter(agent))

	content := strings.Join(sections, "\n")

	if m.width > 60 {
		maxWidth := min(m.width-4, 72)
		return frameBorder.Width(maxWidth).Render(content)
	}
	return content
}

// ── Banner ───────────────────────────────────────────────────

func renderBanner(tick int) string {
	// Animated side accents
	frames := []string{"\u25c6", "\u25c7", "\u25c6", "\u25c7"}
	accent := frames[tick%len(frames)]

	title := bannerStyle.Render("  MULTITAB")
	glow := cursorStyle.Render(accent)

	return fmt.Sprintf("\n  %s %s %s\n  %s",
		glow, title, glow,
		subtitleStyle.Render("multi-agent push orchestrator"))
}

// ── Agent table ──────────────────────────────────────────────

func renderAgentTable(agents []git.Agent, cursor, tick int) string {
	if len(agents) == 0 {
		return "\n" + statusIdleStyle.Render("  No agents detected. Create worktrees to get started.\n")
	}

	var b strings.Builder

	// Header with decorative line
	b.WriteString("\n")
	b.WriteString(separatorStyle.Render("  \u2500\u2500 "))
	b.WriteString(headerStyle.Render("AGENTS"))
	b.WriteString(separatorStyle.Render(" " + strings.Repeat("\u2500", 50)))
	b.WriteString("\n\n")

	// Column headers
	header := fmt.Sprintf("    %-26s %-16s %-8s %-6s", "", "STATUS", "COMMITS", "FILES")
	b.WriteString(separatorStyle.Render(header))
	b.WriteString("\n")

	// Rows
	for i, agent := range agents {
		isSelected := i == cursor

		// Cursor indicator
		pointer := "  "
		if isSelected {
			pointer = cursorStyle.Render("\u25b8 ")
		}

		// Agent icon
		icon := agentIcon(agent.Status, tick)

		// Agent name
		var name string
		if isSelected {
			name = agentNameSelectedStyle.Render(agent.Name)
		} else {
			name = agentNameStyle.Render(agent.Name)
		}

		// Status badge
		status := renderStatusBadge(agent.Status, tick)

		// Counts
		commits := fmt.Sprintf("%-8d", agent.Commits)
		files := fmt.Sprintf("%-6d", agent.Files)

		b.WriteString(fmt.Sprintf("%s%s %s%s%s%s\n", pointer, icon, name, status, commits, files))

		// Show stale hint inline
		if (agent.Status == git.StatusStale || agent.Status == git.StatusAbandoned) && !isSelected {
			hint := staleHint(agent)
			b.WriteString(statusIndicatorStyle.Render("      " + hint))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func renderStatusBadge(status git.AgentStatus, tick int) string {
	switch status {
	case git.StatusStaged:
		return statusStagedStyle.Render(fmt.Sprintf("%-16s", "\u2713 STAGED"))
	case git.StatusWorking:
		return statusWorkingStyle.Render(fmt.Sprintf("%-16s", "\u25cf ACTIVE"))
	case git.StatusStale:
		// Blinking effect for attention
		if tick%4 < 3 {
			return statusStaleStyle.Render(fmt.Sprintf("%-16s", "\u25cc STALE"))
		}
		return statusStaleStyle.Render(fmt.Sprintf("%-16s", "  STALE"))
	case git.StatusAbandoned:
		return statusAbandonedStyle.Render(fmt.Sprintf("%-16s", "\u2205 ABANDONED"))
	default:
		return statusIdleStyle.Render(fmt.Sprintf("%-16s", "\u25cb IDLE"))
	}
}

func agentIcon(status git.AgentStatus, tick int) string {
	switch status {
	case git.StatusStaged:
		return statusStagedStyle.Render("\u25c8")
	case git.StatusWorking:
		// Pulsing dot
		frames := []string{"\u25cf", "\u25cb", "\u25cf", "\u25cf"}
		return statusWorkingStyle.Render(frames[tick%len(frames)])
	case git.StatusStale:
		return statusStaleStyle.Render("\u25cc")
	case git.StatusAbandoned:
		return statusAbandonedStyle.Render("\u25cb")
	default:
		return statusIdleStyle.Render("\u25cb")
	}
}

func staleHint(agent git.Agent) string {
	if agent.Status == git.StatusAbandoned {
		return "\u2514 already pushed, safe to discard"
	}
	if agent.StaleFor > 0 {
		return fmt.Sprintf("\u2514 inactive %s, %d unpushed commit(s)",
			formatStaleTime(agent.StaleFor), agent.Commits)
	}
	return ""
}

// ── Deploy queue ─────────────────────────────────────────────

func renderQueueBar(ready, total, tick int) string {
	if total == 0 {
		return ""
	}

	// Animated label
	var label string
	if ready == total && total > 0 {
		label = statusOkStyle.Render(fmt.Sprintf("  DEPLOY QUEUE  %d/%d  ALL READY", ready, total))
	} else {
		label = headerStyle.Render(fmt.Sprintf("  DEPLOY QUEUE  %d/%d", ready, total))
	}

	barWidth := 40
	filled := 0
	if total > 0 {
		filled = (ready * barWidth) / total
	}
	pct := 0
	if total > 0 {
		pct = (ready * 100) / total
	}

	// Animated fill with scanning effect
	var barParts []string
	for i := 0; i < barWidth; i++ {
		if i < filled {
			barParts = append(barParts, "\u2588")
		} else if i == filled && tick%2 == 0 {
			barParts = append(barParts, "\u2592")
		} else {
			barParts = append(barParts, "\u2591")
		}
	}

	filledStr := strings.Join(barParts[:filled], "")
	restStr := strings.Join(barParts[filled:], "")
	barStyled := statusOkStyle.Render(filledStr) + statusIdleStyle.Render(restStr)

	return fmt.Sprintf("\n%s\n  %s  %s\n",
		label,
		barStyled,
		statusIndicatorStyle.Render(fmt.Sprintf("%d%%", pct)),
	)
}

// ── Staged commits ───────────────────────────────────────────

func renderCommitsPanel(commits []git.StagedCommit) string {
	var lines []string
	for _, c := range commits {
		hash := commitHashStyle.Render(c.Hash)
		msg := commitStyle.Render(" " + c.Message)
		lines = append(lines, "  "+hash+msg)
	}
	content := strings.Join(lines, "\n")

	return panelStyle.Render(
		panelTitleStyle.Render(" \u2261 STAGED COMMITS ") + "\n" + content,
	)
}

// ── System status ────────────────────────────────────────────

func renderStatusBar(m Model) string {
	var items []string

	// Conflicts
	if len(m.state.Conflicts) > 0 {
		items = append(items, statusWarnStyle.Render(
			fmt.Sprintf("  \u26a0 CONFLICTS  %d file(s) touched by multiple agents", len(m.state.Conflicts))))
	} else {
		items = append(items, statusOkStyle.Render("  \u2713 CONFLICTS  None"))
	}

	// Migrations
	if m.state.HasMigrations {
		items = append(items, statusWarnStyle.Render("  \u26a0 MIGRATIONS  Pending"))
	}

	// Stale agents
	staleCount := 0
	for _, a := range m.state.Agents {
		if a.Status == git.StatusStale || a.Status == git.StatusAbandoned {
			staleCount++
		}
	}
	if staleCount > 0 {
		items = append(items, statusStaleStyle.Render(
			fmt.Sprintf("  \u25cc STALE  %d worktree(s) need attention", staleCount)))
	}

	// Last deploy
	if m.state.LastPushHash != "" {
		items = append(items, statusIndicatorStyle.Render(
			fmt.Sprintf("  \u2022 LAST DEPLOY  %s \u2014 %s", m.state.LastPushTime, m.state.LastPushHash)))
	}

	return "\n" + strings.Join(items, "\n")
}

// ── Intel panel (drill-down) ─────────────────────────────────

func renderIntelPanel(agent git.Agent, tick int) string {
	var lines []string

	// Agent name + status header
	icon := agentIcon(agent.Status, tick)
	badge := renderStatusBadge(agent.Status, tick)
	lines = append(lines, fmt.Sprintf("  %s %s  %s",
		icon, intelHeaderStyle.Render(agent.Name), badge))
	lines = append(lines, "")

	// Verdict
	verdict := agentVerdict(agent)
	lines = append(lines, "  "+verdict)
	lines = append(lines, "")

	// Key info
	if !agent.LastCommitTime.IsZero() {
		ago := formatStaleTime(time.Since(agent.LastCommitTime))
		lines = append(lines, renderIntelRow("Last active", ago+" ago"))
	}
	lines = append(lines, renderIntelRow("Branch", agent.Branch))
	lines = append(lines, renderIntelRow("Path", agent.Path))
	lines = append(lines, renderIntelRow("Commits", fmt.Sprintf("%d ahead of origin/main", agent.Commits)))
	lines = append(lines, renderIntelRow("Dirty files", fmt.Sprintf("%d", agent.DirtyFiles)))

	if agent.AlreadyPushed {
		lines = append(lines, renderIntelRow("Pushed?", statusOkStyle.Render("Yes \u2014 all work is on remote")))
	} else if agent.Commits > 0 {
		lines = append(lines, renderIntelRow("Pushed?", statusWarnStyle.Render("No \u2014 unpushed commits")))
	}

	// Commit messages
	if len(agent.CommitMessages) > 0 {
		lines = append(lines, "")
		lines = append(lines, "  "+panelTitleStyle.Render("COMMITS"))
		for _, msg := range agent.CommitMessages {
			lines = append(lines, "    "+commitHashStyle.Render("\u2022")+" "+commitStyle.Render(msg))
		}
	}

	// Changed files (show first 15)
	if len(agent.ChangedFiles) > 0 {
		lines = append(lines, "")
		lines = append(lines, "  "+panelTitleStyle.Render(fmt.Sprintf("FILES (%d)", len(agent.ChangedFiles))))
		limit := 15
		if len(agent.ChangedFiles) < limit {
			limit = len(agent.ChangedFiles)
		}
		for _, f := range agent.ChangedFiles[:limit] {
			lines = append(lines, "    "+intelFileStyle.Render(f))
		}
		if len(agent.ChangedFiles) > 15 {
			lines = append(lines, statusIndicatorStyle.Render(
				fmt.Sprintf("    ... and %d more", len(agent.ChangedFiles)-15)))
		}
	}

	content := strings.Join(lines, "\n")
	return "\n" + panelActiveStyle.Render(content)
}

func agentVerdict(agent git.Agent) string {
	switch agent.Status {
	case git.StatusAbandoned:
		return statusOkStyle.Render("\u2714 All work already on remote. Safe to discard.")
	case git.StatusStale:
		if agent.DirtyFiles > 0 {
			return statusWarnStyle.Render(fmt.Sprintf(
				"\u26a0 Inactive with %d uncommitted file(s). Review before discarding.", agent.DirtyFiles))
		}
		return statusStaleStyle.Render(fmt.Sprintf(
			"\u25cc Inactive with %d unpushed commit(s). Stage or discard?", agent.Commits))
	case git.StatusWorking:
		return statusWorkingStyle.Render("\u25cf Agent is actively working.")
	case git.StatusStaged:
		return statusStagedStyle.Render("\u2713 Merged to local main, ready for push.")
	default:
		return statusIdleStyle.Render("\u25cb No activity.")
	}
}

func renderIntelRow(label, value string) string {
	return "  " + intelLabelStyle.Render(label) + " " + intelValueStyle.Render(value)
}

// ── Push progress ────────────────────────────────────────────

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

	header := pushStepActiveStyle.Render(fmt.Sprintf(" \u25b6 DEPLOYING  %s ", elapsed))

	return "\n" + panelActiveStyle.Render(
		header + "\n" + strings.Join(lines, "\n"),
	)
}

// ── Footers ──────────────────────────────────────────────────

func renderDashboardFooter(m Model) string {
	if m.pushing {
		spinner := spinnerFrames[m.spinFrame%len(spinnerFrames)]
		return "\n  " + pushStepActiveStyle.Render(spinner+" deploying...")
	}

	keys := []struct{ key, label string }{
		{"\u2191\u2193", "navigate"},
		{"\u21b5", "inspect"},
		{"p", "push"},
		{"r", "refresh"},
		{"q", "quit"},
	}

	var parts []string
	for _, k := range keys {
		parts = append(parts, footerKeyStyle.Render(k.key)+" "+footerStyle.Render(k.label))
	}

	return "\n  " + strings.Join(parts, "  "+separatorStyle.Render("\u2502")+"  ")
}

func renderIntelFooter(agent git.Agent) string {
	keys := []struct{ key, label string }{
		{"esc", "back"},
		{"\u2191\u2193", "prev/next agent"},
	}

	if agent.Status == git.StatusStale || agent.Status == git.StatusAbandoned || agent.Status == git.StatusIdle {
		keys = append(keys, struct{ key, label string }{"x", "discard"})
	}

	keys = append(keys, struct{ key, label string }{"q", "quit"})

	var parts []string
	for _, k := range keys {
		parts = append(parts, footerKeyStyle.Render(k.key)+" "+footerStyle.Render(k.label))
	}

	return "\n  " + strings.Join(parts, "  "+separatorStyle.Render("\u2502")+"  ")
}

// ── Helpers ──────────────────────────────────────────────────

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

func formatStaleTime(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		m := int(d.Minutes())
		if m == 1 {
			return "1 min"
		}
		return fmt.Sprintf("%d min", m)
	}
	if d < 24*time.Hour {
		h := int(d.Hours())
		if h == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", h)
	}
	days := int(d.Hours() / 24)
	if days == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", days)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
