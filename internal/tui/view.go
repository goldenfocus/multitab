package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/goldenfocus/multitab/internal/git"
)

var spinnerFrames = []string{"\u280b", "\u2819", "\u2839", "\u2838", "\u283c", "\u2834", "\u2826", "\u2827", "\u2807", "\u280f"}

// View renders the spaceship dashboard.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf(" Error: %v", m.err))
	}

	if m.state == nil {
		spinner := spinnerFrames[m.tick%len(spinnerFrames)]
		return bannerStyle.Render("  MULTITAB") + "\n\n  " +
			pushStepActiveStyle.Render(spinner+" scanning worktrees...")
	}

	switch m.mode {
	case viewIntel:
		return m.renderIntelView()
	case viewSpawn:
		return m.renderSpawnView()
	case viewLog:
		return renderLogView(m)
	case viewPlayback:
		return m.renderPlaybackView()
	default:
		return m.renderDashboardView()
	}
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Dashboard view
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func (m Model) renderDashboardView() string {
	var sections []string

	sections = append(sections, renderBanner(m.tick))
	sections = append(sections, renderAgentTable(m.state.Agents, m.cursor, m.tick))
	sections = append(sections, renderQueueBar(m.state.ReadyCount, m.state.TotalCount, m.tick))

	if len(m.state.StagedCommits) > 0 {
		sections = append(sections, renderCommitsPanel(m.state.StagedCommits))
	}

	sections = append(sections, renderSystemStatus(m))

	if m.pushing {
		sections = append(sections, renderLivePush(m))
	}
	if m.pushErr != nil && !m.pushing {
		sections = append(sections, errorStyle.Render(fmt.Sprintf("  \u2718 Push failed: %v", m.pushErr)))
	}
	if m.pushDone {
		sections = append(sections,
			successStyle.Render(fmt.Sprintf("  \u2714 Deployed in %s", m.pushElapsed.Round(time.Millisecond))))
	}

	// Spawn feedback
	if m.spawnOk != "" {
		sections = append(sections, successStyle.Render("  \u2714 "+m.spawnOk))
	}
	if m.spawnErr != nil {
		sections = append(sections, errorStyle.Render(fmt.Sprintf("  \u2718 Spawn failed: %v", m.spawnErr)))
	}

	sections = append(sections, renderDashboardFooter(m))

	content := strings.Join(sections, "\n")
	if m.width > 60 {
		maxWidth := clampInt(m.width-4, 60, 76)
		return frameBorder.Width(maxWidth).Render(content)
	}
	return content
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Intel view (drill-down)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func (m Model) renderIntelView() string {
	if m.state == nil || m.cursor >= len(m.state.Agents) {
		return m.renderDashboardView()
	}

	agent := m.state.Agents[m.cursor]
	var sections []string

	// Header with agent name + status (outside viewport so always visible)
	icon := agentIcon(agent.Status, m.tick)
	badge := renderStatusBadge(agent.Status, m.tick)
	sections = append(sections, fmt.Sprintf("\n  %s %s  %s",
		icon, intelHeaderStyle.Render(agent.Name), badge))

	// Scrollable viewport with intel content
	sections = append(sections, "\n"+panelStyle.Render(m.viewport.View()))

	// Scroll indicator
	pct := m.viewport.ScrollPercent()
	lines := strings.Count(m.logContent, "\n") + 1
	scrollInfo := statusIndicatorStyle.Render(
		fmt.Sprintf("  %d%% \u2502 %d lines", int(pct*100), lines))
	sections = append(sections, scrollInfo)

	sections = append(sections, renderIntelFooter(agent))

	content := strings.Join(sections, "\n")
	if m.width > 60 {
		maxWidth := clampInt(m.width-4, 60, 76)
		return frameBorder.Width(maxWidth).Render(content)
	}
	return content
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Spawn view (new agent prompt)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func (m Model) renderSpawnView() string {
	var sections []string

	sections = append(sections, renderBanner(m.tick))
	sections = append(sections, "")

	// Spawn prompt panel
	var lines []string
	lines = append(lines, "  "+spawnPromptStyle.Render("\u25b6 NEW AGENT"))
	lines = append(lines, "")
	lines = append(lines, "  "+spawnHintStyle.Render("Type a task description or path to an .md file."))
	lines = append(lines, "  "+spawnHintStyle.Render("multitab will create a worktree and launch Claude."))
	lines = append(lines, "")
	lines = append(lines, "  "+m.promptInput.View())
	lines = append(lines, "")

	sections = append(sections, panelActiveStyle.Render(strings.Join(lines, "\n")))

	// Footer
	keys := []struct{ key, label string }{
		{"enter", "launch"},
		{"esc", "cancel"},
	}
	var parts []string
	for _, k := range keys {
		parts = append(parts, footerKeyStyle.Render(k.key)+" "+footerStyle.Render(k.label))
	}
	sections = append(sections, "\n  "+strings.Join(parts, "  "+separatorStyle.Render("\u2502")+"  "))

	content := strings.Join(sections, "\n")
	if m.width > 60 {
		maxWidth := clampInt(m.width-4, 60, 76)
		return frameBorder.Width(maxWidth).Render(content)
	}
	return content
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Banner
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func renderBanner(tick int) string {
	// Pulsing accent diamonds
	diamonds := []string{"\u25c6\u25c7\u25c6", "\u25c7\u25c6\u25c7", "\u25c6\u25c6\u25c7", "\u25c7\u25c7\u25c6"}
	left := bannerAccentStyle.Render(diamonds[tick%len(diamonds)])
	right := bannerAccentStyle.Render(diamonds[(tick+2)%len(diamonds)])

	title := bannerStyle.Render(" M U L T I T A B ")
	sub := subtitleStyle.Render("multi-agent command center")

	return fmt.Sprintf("\n  %s%s%s\n  %s", left, title, right, sub)
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Agent table
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func renderAgentTable(agents []git.Agent, cursor, tick int) string {
	if len(agents) == 0 {
		return "\n" + statusIdleStyle.Render("  No agents. Press ") +
			footerKeyStyle.Render("n") +
			statusIdleStyle.Render(" to launch one.\n")
	}

	// Separate active agents from dormant (IDLE with 0 commits, 0 files)
	var activeIdxs []int
	var dormantCount int
	for i, agent := range agents {
		if agent.Status == git.StatusIdle && agent.Commits == 0 && agent.Files == 0 && agent.DirtyFiles == 0 {
			dormantCount++
		} else {
			activeIdxs = append(activeIdxs, i)
		}
	}

	var b strings.Builder

	// Section header
	b.WriteString("\n")
	b.WriteString(separatorStyle.Render("  \u2501\u2501 "))
	b.WriteString(sectionTitleStyle.Render("AGENTS"))
	b.WriteString(separatorStyle.Render(" " + strings.Repeat("\u2501", 48)))
	b.WriteString("\n\n")

	// Column headers
	header := fmt.Sprintf("    %-26s %-16s %-8s %-6s", "", "STATUS", "COMMITS", "FILES")
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	for i, agent := range agents {
		// Skip dormant agents in main list (unless selected)
		isDormant := agent.Status == git.StatusIdle && agent.Commits == 0 && agent.Files == 0 && agent.DirtyFiles == 0
		if isDormant && i != cursor {
			continue
		}

		isSelected := i == cursor

		pointer := "  "
		if isSelected {
			pointer = cursorStyle.Render("\u25b8 ")
		}

		icon := agentIcon(agent.Status, tick)

		displayName := cleanAgentName(agent.Name)

		var name string
		if isSelected {
			name = agentNameSelectedStyle.Render(displayName)
		} else {
			name = agentNameStyle.Render(displayName)
		}

		status := renderStatusBadge(agent.Status, tick)
		commits := fmt.Sprintf("%-8d", agent.Commits)
		files := fmt.Sprintf("%-6d", agent.Files)

		b.WriteString(fmt.Sprintf("%s%s %s%s%s%s\n", pointer, icon, name, status, commits, files))

		// Inline stale hint (only when not selected — intel view shows full detail)
		if (agent.Status == git.StatusStale || agent.Status == git.StatusAbandoned) && !isSelected {
			hint := staleHint(agent)
			b.WriteString(statusIndicatorStyle.Render("      " + hint))
			b.WriteString("\n")
		}
	}

	// Dormant summary (collapsed)
	if dormantCount > 0 {
		label := "dormant worktree"
		if dormantCount > 1 {
			label = "dormant worktrees"
		}
		b.WriteString(statusIdleStyle.Render(
			fmt.Sprintf("    \u25cb %d %s (idle, no changes)\n", dormantCount, label)))
	}

	return b.String()
}

// cleanAgentName makes agent names display-friendly.
// Strips UUID/hex suffixes, truncates, and cleans up auto-generated garbage.
func cleanAgentName(name string) string {
	// Strip common auto-generated suffixes (UUIDs, hex hashes)
	// e.g., "worktree-agent-a3e4aa70" → "worktree-agent"
	// e.g., "agent-afae1234" → "agent"
	parts := strings.Split(name, "-")
	var cleaned []string
	for _, p := range parts {
		// Skip parts that look like hex hashes (8+ hex chars)
		if len(p) >= 8 && isHex(p) {
			continue
		}
		// Skip "worktree" prefix — redundant info
		if p == "worktree" {
			continue
		}
		cleaned = append(cleaned, p)
	}

	result := strings.Join(cleaned, "-")
	if result == "" {
		result = name // fallback to original if we stripped everything
	}

	// Truncate to fit table
	if len(result) > 22 {
		result = result[:19] + "..."
	}

	return result
}

func isHex(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

func renderStatusBadge(status git.AgentStatus, tick int) string {
	switch status {
	case git.StatusStaged:
		return statusStagedStyle.Render(fmt.Sprintf("%-16s", "\u2713 STAGED"))
	case git.StatusWorking:
		return statusWorkingStyle.Render(fmt.Sprintf("%-16s", "\u25cf ACTIVE"))
	case git.StatusStale:
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

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Deploy queue
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func renderQueueBar(ready, total, tick int) string {
	if total == 0 {
		return ""
	}

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

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Staged commits
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func renderCommitsPanel(commits []git.StagedCommit) string {
	var lines []string
	for _, c := range commits {
		hash := commitHashStyle.Render(c.Hash[:7])
		msg := commitStyle.Render(" " + c.Message)
		lines = append(lines, "  "+hash+msg)
	}
	content := strings.Join(lines, "\n")

	return panelStyle.Render(
		panelTitleStyle.Render(" \u2261 STAGED COMMITS ") + "\n" + content,
	)
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// System status
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func renderSystemStatus(m Model) string {
	var items []string

	if len(m.state.Conflicts) > 0 {
		items = append(items, statusWarnStyle.Render(
			fmt.Sprintf("  \u26a0 CONFLICTS  %d file(s) touched by multiple agents", len(m.state.Conflicts))))
	} else {
		items = append(items, statusOkStyle.Render("  \u2713 CONFLICTS  clear"))
	}

	if m.state.HasMigrations {
		items = append(items, statusWarnStyle.Render("  \u26a0 MIGRATIONS  pending"))
	}

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

	if m.state.LastPushHash != "" {
		items = append(items, statusIndicatorStyle.Render(
			fmt.Sprintf("  \u2022 LAST DEPLOY  %s \u2014 %s", m.state.LastPushTime, m.state.LastPushHash)))
	}

	return "\n" + strings.Join(items, "\n")
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Intel panel
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func renderIntelContent(agent git.Agent, tick int) string {
	return renderIntelPanel(agent, tick)
}

func renderIntelPanel(agent git.Agent, tick int) string {
	var lines []string

	verdict := agentVerdict(agent)
	lines = append(lines, "  "+verdict)
	lines = append(lines, "")

	if !agent.LastCommitTime.IsZero() {
		ago := formatStaleTime(time.Since(agent.LastCommitTime))
		lines = append(lines, renderIntelRow("Last active", ago+" ago"))
	}
	lines = append(lines, renderIntelRow("Branch", agent.Branch))
	lines = append(lines, renderIntelRow("Path", agent.Path))
	lines = append(lines, renderIntelRow("Commits", fmt.Sprintf("%d ahead of origin/main", agent.Commits)))
	lines = append(lines, renderIntelRow("Dirty files", fmt.Sprintf("%d", agent.DirtyFiles)))

	// Only show push status when there's actual work to push
	if agent.Commits > 0 || agent.DirtyFiles > 0 {
		if agent.AlreadyPushed && agent.DirtyFiles == 0 {
			lines = append(lines, renderIntelRow("Pushed?", statusOkStyle.Render("Yes \u2014 all work is on remote")))
		} else if agent.Commits > 0 {
			lines = append(lines, renderIntelRow("Pushed?", statusWarnStyle.Render("No \u2014 unpushed commits")))
		}
	}

	// Last human prompt from Claude conversation
	if agent.LastPrompt != "" {
		lines = append(lines, "")
		lines = append(lines, "  "+panelTitleStyle.Render("LAST HUMAN PROMPT"))
		prompt := agent.LastPrompt
		if len(prompt) > 200 {
			prompt = prompt[:200] + "..."
		}
		lines = append(lines, "    "+commitStyle.Render(prompt))
		if !agent.LastPromptAt.IsZero() {
			ago := formatStaleTime(time.Since(agent.LastPromptAt))
			lines = append(lines, "    "+statusIndicatorStyle.Render(ago+" ago"))
		}
		if agent.HumanMsgCount > 0 {
			lines = append(lines, "    "+statusIndicatorStyle.Render(
				fmt.Sprintf("%d human messages in session", agent.HumanMsgCount)))
		}
	}

	if len(agent.CommitMessages) > 0 {
		lines = append(lines, "")
		lines = append(lines, "  "+panelTitleStyle.Render("COMMITS"))
		for _, msg := range agent.CommitMessages {
			lines = append(lines, "    "+commitHashStyle.Render("\u2022")+" "+commitStyle.Render(msg))
		}
	}

	if len(agent.ChangedFiles) > 0 {
		lines = append(lines, "")
		lines = append(lines, "  "+panelTitleStyle.Render(fmt.Sprintf("FILES (%d)", len(agent.ChangedFiles))))
		for _, f := range agent.ChangedFiles {
			lines = append(lines, "    "+intelFileStyle.Render(f))
		}
	}

	return strings.Join(lines, "\n")
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

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Playback view (conversation replay)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func (m Model) renderPlaybackView() string {
	if len(m.chatTurns) == 0 {
		return m.renderIntelView()
	}

	var sections []string

	// Compact header with agent name + turn counter
	agentName := ""
	if m.state != nil && m.chatAgentIdx < len(m.state.Agents) {
		agentName = m.state.Agents[m.chatAgentIdx].Name
	}

	turnLabel := fmt.Sprintf("TURN %d/%d", m.chatTurnIdx+1, len(m.chatTurns))
	sections = append(sections, fmt.Sprintf("\n  %s  %s  %s",
		bannerAccentStyle.Render("\u25c6"),
		intelHeaderStyle.Render(agentName),
		statusWorkingStyle.Render(turnLabel),
	))

	// Progress dots — visual indicator of position
	dots := renderTurnDots(m.chatTurnIdx, len(m.chatTurns))
	sections = append(sections, "  "+dots)

	// The viewport shows the current turn content
	sections = append(sections, "\n"+panelStyle.Render(m.viewport.View()))

	// Scroll indicator
	pct := m.viewport.ScrollPercent()
	lines := strings.Count(m.logContent, "\n") + 1
	scrollInfo := statusIndicatorStyle.Render(
		fmt.Sprintf("  %d%% \u2502 %d lines", int(pct*100), lines))
	sections = append(sections, scrollInfo)

	// Footer
	keys := []struct{ key, label string }{
		{"esc", "back"},
		{"\u2190\u2192", "prev/next turn"},
		{"\u2191\u2193", "scroll"},
		{"g/G", "top/bottom"},
	}
	var parts []string
	for _, k := range keys {
		parts = append(parts, footerKeyStyle.Render(k.key)+" "+footerStyle.Render(k.label))
	}
	sections = append(sections, "\n  "+strings.Join(parts, "  "+separatorStyle.Render("\u2502")+"  "))

	content := strings.Join(sections, "\n")
	if m.width > 60 {
		maxWidth := clampInt(m.width-4, 60, 76)
		return frameBorder.Width(maxWidth).Render(content)
	}
	return content
}

func renderTurnDots(current, total int) string {
	if total <= 1 {
		return ""
	}
	var dots []string
	maxDots := min(total, 30) // cap for very long conversations
	for i := 0; i < maxDots; i++ {
		if i == current {
			dots = append(dots, statusWorkingStyle.Render("\u25cf"))
		} else {
			dots = append(dots, statusIndicatorStyle.Render("\u25cb"))
		}
	}
	if total > maxDots {
		dots = append(dots, statusIndicatorStyle.Render(fmt.Sprintf(" +%d", total-maxDots)))
	}
	return strings.Join(dots, " ")
}

func formatTurnContent(turn ChatTurn) string {
	var sb strings.Builder

	// Human prompt
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	header := "  HUMAN #" + fmt.Sprintf("%d", turn.TurnNumber)
	if turn.HumanTime != "" {
		header += "  [" + turn.HumanTime + "]"
	}
	sb.WriteString(header + "\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	humanText := turn.HumanText
	if len(humanText) > 3000 {
		humanText = humanText[:3000] + "\n\n  ... (truncated)"
	}
	sb.WriteString(humanText)
	sb.WriteString("\n\n")

	// Assistant response
	if turn.AssistantText != "" {
		sb.WriteString("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄\n")
		aHeader := "  CLAUDE"
		if turn.AssistantTime != "" {
			aHeader += "  [" + turn.AssistantTime + "]"
		}
		sb.WriteString(aHeader + "\n")
		sb.WriteString("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄\n\n")

		assistantText := turn.AssistantText
		if len(assistantText) > 5000 {
			assistantText = assistantText[:5000] + "\n\n  ... (truncated)"
		}
		sb.WriteString(assistantText)
	} else {
		sb.WriteString("\n  (awaiting response...)")
	}

	return sb.String()
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Push progress
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

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

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Footers
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func renderDashboardFooter(m Model) string {
	if m.pushing {
		spinner := spinnerFrames[m.spinFrame%len(spinnerFrames)]
		return "\n  " + pushStepActiveStyle.Render(spinner+" deploying...")
	}

	keys := []struct{ key, label string }{
		{"\u2191\u2193", "navigate"},
		{"\u21b5", "inspect"},
		{"c", "replay"},
		{"s", "stage"},
		{"x", "kill"},
		{"n", "new agent"},
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
		{"\u2190", "back"},
		{"\u2191\u2193", "scroll"},
		{"\u2192", "replay"},
		{"tab", "next agent"},
		{"l", "logs"},
		{"s", "stage"},
		{"x", "kill"},
	}

	var parts []string
	for _, k := range keys {
		parts = append(parts, footerKeyStyle.Render(k.key)+" "+footerStyle.Render(k.label))
	}

	return "\n  " + strings.Join(parts, "  "+separatorStyle.Render("\u2502")+"  ")
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Helpers
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

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

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
