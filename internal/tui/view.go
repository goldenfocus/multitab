package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/goldenfocus/multitab/internal/git"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

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

// ─────────────────────────────────────────────────
// Dashboard view
// ─────────────────────────────────────────────────

func (m Model) renderDashboardView() string {
	// ── Left panel: agent dashboard ──
	var sections []string

	sections = append(sections, renderBanner(m.tick))
	sections = append(sections, "")
	sections = append(sections, renderAgentTable(m.state.Agents, m.cursor, m.tick))

	if m.state.TotalCount > 0 {
		sections = append(sections, renderQueueBar(m.state.ReadyCount, m.state.TotalCount, m.tick))
	}

	if len(m.state.StagedCommits) > 0 {
		sections = append(sections, renderCommitsPanel(m.state.StagedCommits))
	}

	sections = append(sections, renderSystemStatus(m))

	if m.pushing {
		sections = append(sections, renderLivePush(m))
	}
	if m.pushErr != nil && !m.pushing {
		sections = append(sections, errorStyle.Render(fmt.Sprintf("  ✗ Push failed: %v", m.pushErr)))
	}
	if m.pushDone {
		sections = append(sections,
			successStyle.Render(fmt.Sprintf("  ✓ Deployed in %s", m.pushElapsed.Round(time.Millisecond))))
	}

	// Spawn feedback
	if m.spawnOk != "" {
		sections = append(sections, successStyle.Render("  ✓ "+m.spawnOk))
	}
	if m.spawnErr != nil {
		sections = append(sections, errorStyle.Render(fmt.Sprintf("  ✗ Spawn failed: %v", m.spawnErr)))
	}

	// Bottom status bar
	sections = append(sections, "")
	sections = append(sections, renderStatusBar(m))
	sections = append(sections, renderDashboardFooter(m))

	dashContent := strings.Join(sections, "\n")

	// ── Wide layout: dashboard + chat side by side ──
	if m.width >= 120 {
		leftWidth := clampInt(m.width*45/100, 50, 80)
		rightWidth := m.width - leftWidth - 4

		left := frameBorder.Width(leftWidth).Render(dashContent)
		right := m.renderChatPanel(rightWidth)

		return lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right)
	}

	// ── Narrow layout: dashboard only ──
	if m.width > 60 {
		maxWidth := clampInt(m.width-4, 60, 80)
		return frameBorder.Width(maxWidth).Render(dashContent)
	}
	return dashContent
}

// ─────────────────────────────────────────────────
// Chat panel (right side of split layout)
// ─────────────────────────────────────────────────

func (m Model) renderChatPanel(width int) string {
	innerWidth := width - 4

	var sections []string

	// Chat header
	sections = append(sections, renderChatPanelHeader(m.tick))

	// Chat messages
	chatContent := renderChatMessages(m.chatHistory, m.chatStreamBuf, m.chatStreaming, innerWidth, m.tick)
	// Use a fixed-height panel that fills available space
	chatHeight := maxInt(m.height-12, 8)
	chatPanel := lipgloss.NewStyle().
		Width(innerWidth).
		Height(chatHeight).
		Render(chatContent)
	// Only show the tail (most recent messages)
	chatLines := strings.Split(chatPanel, "\n")
	if len(chatLines) > chatHeight {
		chatLines = chatLines[len(chatLines)-chatHeight:]
	}
	sections = append(sections, panelStyle.Width(innerWidth).Render(strings.Join(chatLines, "\n")))

	// Voice indicator
	sections = append(sections, renderVoiceIndicator(m.voice, m.speaking, m.tick))

	// Input box
	var inputLines []string
	inputLines = append(inputLines, "")
	if m.chatStreaming {
		spinner := spinnerFrames[m.tick%len(spinnerFrames)]
		inputLines = append(inputLines, "  "+pushStepActiveStyle.Render(spinner+" commander is thinking..."))
	} else {
		inputLines = append(inputLines, "  "+chatPromptStyle.Render("▸")+" "+m.chatInput.View())
	}
	inputLines = append(inputLines, "")

	inputBorder := panelStyle
	if m.chatFocused {
		inputBorder = panelActiveStyle
	}
	sections = append(sections, inputBorder.Width(innerWidth).Render(strings.Join(inputLines, "\n")))

	// Chat footer
	sections = append(sections, renderChatPanelFooter(m))

	content := strings.Join(sections, "\n")

	// Use the hull border for the chat panel too
	chatBorderStyle := frameBorder.Copy().BorderForeground(lipgloss.Color("#334155"))
	if m.chatFocused {
		chatBorderStyle = chatBorderStyle.BorderForeground(cyan)
	}
	return chatBorderStyle.Width(width).Render(content)
}

func renderChatPanelHeader(tick int) string {
	title := chatCmdLabelStyle.Render(" COMMANDER ")
	sub := subtitleStyle.Render("  mission control AI")
	scan := dimSeparatorStyle.Render("  " + strings.Repeat("━", 40))
	return fmt.Sprintf("  %s\n%s\n%s", title, sub, scan)
}

func renderChatPanelFooter(m Model) string {
	var keys []struct{ key, label string }
	if m.chatFocused {
		keys = []struct{ key, label string }{
			{"esc", "dashboard"},
			{"enter", "send"},
			{"ctrl+v", m.voice.String()},
		}
		if m.voice == voiceManual {
			keys = append(keys, struct{ key, label string }{"ctrl+p", "play"})
		}
	} else {
		keys = []struct{ key, label string }{
			{"/", "chat"},
		}
	}

	var parts []string
	for _, k := range keys {
		parts = append(parts, footerKeyStyle.Render(k.key)+" "+footerStyle.Render(k.label))
	}
	return "\n  " + strings.Join(parts, "  ")
}

// ─────────────────────────────────────────────────
// Intel view (drill-down with scrollable viewport)
// ─────────────────────────────────────────────────

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
		fmt.Sprintf("  %d%% │ %d lines", int(pct*100), lines))
	sections = append(sections, scrollInfo)

	sections = append(sections, renderIntelFooter(agent))

	content := strings.Join(sections, "\n")
	if m.width > 60 {
		maxWidth := clampInt(m.width-4, 60, 80)
		return frameBorder.Width(maxWidth).Render(content)
	}
	return content
}

// ─────────────────────────────────────────────────
// Spawn view (new agent prompt)
// ─────────────────────────────────────────────────

func (m Model) renderSpawnView() string {
	var sections []string

	sections = append(sections, renderBanner(m.tick))
	sections = append(sections, "")

	// Spawn prompt panel
	var lines []string
	lines = append(lines, "")
	lines = append(lines, "  "+spawnPromptStyle.Render("▸ NEW AGENT"))
	lines = append(lines, "")
	lines = append(lines, "  "+spawnHintStyle.Render("Type a task description or path to .md file."))
	lines = append(lines, "  "+spawnHintStyle.Render("A new worktree + Claude session will launch."))
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
	sections = append(sections, "\n  "+strings.Join(parts, "  "))

	content := strings.Join(sections, "\n")
	if m.width > 60 {
		maxWidth := clampInt(m.width-4, 60, 80)
		return frameBorder.Width(maxWidth).Render(content)
	}
	return content
}

// ─────────────────────────────────────────────────
// Banner
// ─────────────────────────────────────────────────

func renderBanner(tick int) string {
	// Pulsing accent diamonds
	frames := []string{"◆◇◆", "◇◆◇", "◆◆◇", "◇◇◆"}
	left := bannerAccentStyle.Render(frames[tick%len(frames)])
	right := bannerAccentStyle.Render(frames[(tick+2)%len(frames)])

	title := bannerStyle.Render(" M U L T I T A B ")
	sub := subtitleStyle.Render("  multi-agent command center")

	// Structural scan line
	scan := dimSeparatorStyle.Render("  " + strings.Repeat("━", 58))

	return fmt.Sprintf("  %s%s%s\n%s\n%s", left, title, right, sub, scan)
}

// ─────────────────────────────────────────────────
// Agent table
// ─────────────────────────────────────────────────

func renderAgentTable(agents []git.Agent, cursor, tick int) string {
	if len(agents) == 0 {
		return "  " + statusIdleStyle.Render("No agents running. Press ") +
			footerKeyStyle.Render("n") +
			statusIdleStyle.Render(" to launch one.")
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

	// Section header — cockpit panel divider
	b.WriteString(separatorStyle.Render("  ━━ "))
	b.WriteString(sectionTitleStyle.Render("AGENTS"))
	b.WriteString(separatorStyle.Render(" " + strings.Repeat("━", 50)))
	b.WriteString("\n\n")

	// Column headers
	header := fmt.Sprintf("    %-26s %-14s %8s %6s", "", "STATUS", "COMMITS", "FILES")
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
			pointer = cursorStyle.Render("▸ ")
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
		commits := fmt.Sprintf("%8d", agent.Commits)
		files := fmt.Sprintf("%6d", agent.Files)

		b.WriteString(fmt.Sprintf("%s%s %s%s%s%s\n", pointer, icon, name, status, commits, files))

		// Inline stale hint
		if (agent.Status == git.StatusStale || agent.Status == git.StatusAbandoned) && !isSelected {
			hint := staleHint(agent)
			b.WriteString(statusIndicatorStyle.Render("     " + hint))
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
		return statusStagedStyle.Render(fmt.Sprintf("%-14s", "✓ STAGED"))
	case git.StatusWorking:
		return statusWorkingStyle.Render(fmt.Sprintf("%-14s", "● ACTIVE"))
	case git.StatusStale:
		if tick%4 < 3 {
			return statusStaleStyle.Render(fmt.Sprintf("%-14s", "◌ STALE"))
		}
		return statusStaleStyle.Render(fmt.Sprintf("%-14s", "  STALE"))
	case git.StatusAbandoned:
		return statusAbandonedStyle.Render(fmt.Sprintf("%-14s", "∅ ABANDONED"))
	default:
		return statusIdleStyle.Render(fmt.Sprintf("%-14s", "○ IDLE"))
	}
}

func agentIcon(status git.AgentStatus, tick int) string {
	switch status {
	case git.StatusStaged:
		return statusStagedStyle.Render("◈")
	case git.StatusWorking:
		frames := []string{"●", "○", "●", "●"}
		return statusWorkingStyle.Render(frames[tick%len(frames)])
	case git.StatusStale:
		return statusStaleStyle.Render("◌")
	case git.StatusAbandoned:
		return statusAbandonedStyle.Render("○")
	default:
		return statusIdleStyle.Render("○")
	}
}

func staleHint(agent git.Agent) string {
	if agent.Status == git.StatusAbandoned {
		return "└ already pushed, safe to discard"
	}
	if agent.StaleFor > 0 {
		return fmt.Sprintf("└ inactive %s, %d unpushed commit(s)",
			formatStaleTime(agent.StaleFor), agent.Commits)
	}
	return ""
}

// ─────────────────────────────────────────────────
// Deploy queue
// ─────────────────────────────────────────────────

func renderQueueBar(ready, total, tick int) string {
	if total == 0 {
		return ""
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
			barParts = append(barParts, "█")
		} else if i == filled && tick%2 == 0 {
			barParts = append(barParts, "░")
		} else {
			barParts = append(barParts, "▁")
		}
	}

	filledStr := strings.Join(barParts[:filled], "")
	restStr := strings.Join(barParts[filled:], "")
	barStyled := queueFilledStyle.Render(filledStr) + queueEmptyStyle.Render(restStr)

	// Panel divider
	divider := separatorStyle.Render("  ━━ ") +
		sectionTitleStyle.Render(fmt.Sprintf("DEPLOY QUEUE %d/%d", ready, total)) +
		separatorStyle.Render(" " + strings.Repeat("━", 36))

	return fmt.Sprintf("\n%s\n  %s  %s\n",
		divider,
		barStyled,
		statusIndicatorStyle.Render(fmt.Sprintf("%d%%", pct)),
	)
}

// ─────────────────────────────────────────────────
// Staged commits
// ─────────────────────────────────────────────────

func renderCommitsPanel(commits []git.StagedCommit) string {
	var lines []string
	for _, c := range commits {
		hash := commitHashStyle.Render(c.Hash[:7])
		msg := commitStyle.Render(" " + c.Message)
		lines = append(lines, "  "+hash+msg)
	}
	content := strings.Join(lines, "\n")

	return panelStyle.Render(
		panelTitleStyle.Render("  STAGED COMMITS") + "\n" + content,
	)
}

// ─────────────────────────────────────────────────
// System status
// ─────────────────────────────────────────────────

func renderSystemStatus(m Model) string {
	var items []string

	if len(m.state.Conflicts) > 0 {
		items = append(items, statusWarnStyle.Render(
			fmt.Sprintf("  ⚠ CONFLICTS  %d file(s) touched by multiple agents", len(m.state.Conflicts))))
	}

	if m.state.HasMigrations {
		items = append(items, statusWarnStyle.Render("  ⚠ MIGRATIONS  pending"))
	}

	staleCount := 0
	for _, a := range m.state.Agents {
		if a.Status == git.StatusStale || a.Status == git.StatusAbandoned {
			staleCount++
		}
	}
	if staleCount > 0 {
		items = append(items, statusStaleStyle.Render(
			fmt.Sprintf("  ◌ STALE  %d worktree(s) need attention", staleCount)))
	}

	if len(items) == 0 {
		return ""
	}

	return "\n" + strings.Join(items, "\n")
}

// ─────────────────────────────────────────────────
// Status bar (bottom — like the website terminal)
// ─────────────────────────────────────────────────

func renderStatusBar(m Model) string {
	var parts []string

	// Agent count
	activeCount := 0
	for _, a := range m.state.Agents {
		if a.Status == git.StatusWorking {
			activeCount++
		}
	}
	if activeCount > 0 {
		parts = append(parts, statusBarActiveStyle.Render(fmt.Sprintf("● %d active", activeCount)))
	}
	parts = append(parts, statusBarStyle.Render(fmt.Sprintf("%d agents", len(m.state.Agents))))

	// Version
	parts = append(parts, statusBarStyle.Render("multitab v0.1.0"))

	// Last deploy
	if m.state.LastPushHash != "" {
		parts = append(parts, statusBarStyle.Render(
			fmt.Sprintf("deployed %s — %s", m.state.LastPushTime, m.state.LastPushHash)))
	}

	sep := statusBarDimStyle.Render(" ┃ ")
	line := dimSeparatorStyle.Render("  " + strings.Repeat("━", 58))
	return line + "\n  " + strings.Join(parts, sep)
}

// ─────────────────────────────────────────────────
// Intel panel
// ─────────────────────────────────────────────────

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
			lines = append(lines, renderIntelRow("Pushed", statusOkStyle.Render("yes — all work on remote")))
		} else if agent.Commits > 0 {
			lines = append(lines, renderIntelRow("Pushed", statusWarnStyle.Render("no — unpushed commits")))
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
			lines = append(lines, "    "+commitHashStyle.Render("•")+" "+commitStyle.Render(msg))
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
		return statusOkStyle.Render("✓ All work already on remote. Safe to discard.")
	case git.StatusStale:
		if agent.DirtyFiles > 0 {
			return statusWarnStyle.Render(fmt.Sprintf(
				"⚠ Inactive with %d uncommitted file(s). Review before discarding.", agent.DirtyFiles))
		}
		return statusStaleStyle.Render(fmt.Sprintf(
			"◌ Inactive with %d unpushed commit(s). Stage or discard?", agent.Commits))
	case git.StatusWorking:
		return statusWorkingStyle.Render("● Agent is actively working.")
	case git.StatusStaged:
		return statusStagedStyle.Render("✓ Merged to local main, ready for push.")
	default:
		return statusIdleStyle.Render("○ No activity.")
	}
}

func renderIntelRow(label, value string) string {
	return "  " + intelLabelStyle.Render(label) + " " + intelValueStyle.Render(value)
}

// ─────────────────────────────────────────────────
// Playback view (conversation replay)
// ─────────────────────────────────────────────────

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
		bannerAccentStyle.Render("◆"),
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
		fmt.Sprintf("  %d%% │ %d lines", int(pct*100), lines))
	sections = append(sections, scrollInfo)

	// Footer
	keys := []struct{ key, label string }{
		{"esc", "back"},
		{"←→", "prev/next turn"},
		{"↑↓", "scroll"},
		{"g/G", "top/bottom"},
	}
	var parts []string
	for _, k := range keys {
		parts = append(parts, footerKeyStyle.Render(k.key)+" "+footerStyle.Render(k.label))
	}
	sections = append(sections, "\n  "+strings.Join(parts, "  "))

	content := strings.Join(sections, "\n")
	if m.width > 60 {
		maxWidth := clampInt(m.width-4, 60, 80)
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
			dots = append(dots, statusWorkingStyle.Render("●"))
		} else {
			dots = append(dots, statusIndicatorStyle.Render("○"))
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
	sb.WriteString("──────────────────────────────────────────────────\n")
	header := "  HUMAN #" + fmt.Sprintf("%d", turn.TurnNumber)
	if turn.HumanTime != "" {
		header += "  [" + turn.HumanTime + "]"
	}
	sb.WriteString(header + "\n")
	sb.WriteString("──────────────────────────────────────────────────\n\n")

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

// ─────────────────────────────────────────────────
// Push progress
// ─────────────────────────────────────────────────

func renderLivePush(m Model) string {
	var lines []string

	elapsed := time.Since(m.push.startAt).Round(time.Millisecond)
	spinner := spinnerFrames[m.spinFrame%len(spinnerFrames)]

	for _, s := range m.push.steps {
		var line string
		switch s.status {
		case stepDone:
			line = pushStepDoneStyle.Render(
				fmt.Sprintf("  ✓ %s  %s", s.step, formatDuration(s.elapsed)))
		case stepRunning:
			line = pushStepActiveStyle.Render(
				fmt.Sprintf("  %s %s", spinner, s.step))
		case stepFailed:
			line = errorStyle.Render(
				fmt.Sprintf("  ✗ %s  FAILED", s.step))
		default:
			line = pushStepPendingStyle.Render(
				fmt.Sprintf("    %s", s.step))
		}
		lines = append(lines, line)
	}

	header := pushStepActiveStyle.Render(fmt.Sprintf("  ▸ DEPLOYING  %s", elapsed))

	return "\n" + panelActiveStyle.Render(
		header + "\n" + strings.Join(lines, "\n"),
	)
}

// ─────────────────────────────────────────────────
// Footers
// ─────────────────────────────────────────────────

func renderDashboardFooter(m Model) string {
	if m.pushing {
		spinner := spinnerFrames[m.spinFrame%len(spinnerFrames)]
		return "\n  " + pushStepActiveStyle.Render(spinner+" deploying...")
	}

	keys := []struct{ key, label string }{
		{"↑↓", "navigate"},
		{"⏎", "inspect"},
		{"/", "commander"},
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

	return "\n  " + strings.Join(parts, "  ")
}

func renderIntelFooter(agent git.Agent) string {
	keys := []struct{ key, label string }{
		{"←", "back"},
		{"↑↓", "scroll"},
		{"→", "replay"},
		{"tab", "next agent"},
		{"l", "logs"},
		{"s", "stage"},
		{"x", "kill"},
	}

	var parts []string
	for _, k := range keys {
		parts = append(parts, footerKeyStyle.Render(k.key)+" "+footerStyle.Render(k.label))
	}

	return "  " + strings.Join(parts, "  ")
}

// ─────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────

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
