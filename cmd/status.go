package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vibeyang/multitab/internal/git"
	"github.com/vibeyang/multitab/internal/queue"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show agent status (text-only, no TUI)",
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return fmt.Errorf("not inside a git repository")
	}

	state, err := queue.Refresh(repoRoot)
	if err != nil {
		return err
	}

	// Header
	fmt.Println("MULTITAB STATUS")
	fmt.Println("═══════════════════════════════════════════════════")

	if len(state.Agents) == 0 {
		fmt.Println("No worktrees found.")
		fmt.Println("Create worktrees with: git worktree add .claude/worktrees/<name> -b <name> origin/main")
		return nil
	}

	// Agents
	fmt.Printf("\n%-28s %-14s %-10s %-6s\n", "AGENT", "STATUS", "COMMITS", "FILES")
	fmt.Println("────────────────────────────────────────────────────")
	for _, agent := range state.Agents {
		icon := statusIcon(agent.Status)
		fmt.Printf("%s %-26s %-14s %-10d %-6d\n",
			icon, agent.Name, agent.Status, agent.Commits, agent.Files)
	}

	// Staged commits
	if len(state.StagedCommits) > 0 {
		fmt.Printf("\nSTAGED COMMITS (%d):\n", len(state.StagedCommits))
		for _, c := range state.StagedCommits {
			fmt.Printf("  %s %s\n", c.Hash, c.Message)
		}
	}

	// Queue
	fmt.Printf("\nDEPLOY QUEUE: %d/%d ready\n", state.ReadyCount, state.TotalCount)

	// Conflicts
	if len(state.Conflicts) > 0 {
		fmt.Printf("\nCONFLICTS (%d):\n", len(state.Conflicts))
		for _, c := range state.Conflicts {
			fmt.Printf("  %s: %s <-> %s\n", c.File, c.Agent1, c.Agent2)
		}
	} else {
		fmt.Println("\nCONFLICTS: None")
	}

	// Migrations
	if state.HasMigrations {
		fmt.Println("MIGRATIONS: Pending changes detected")
	}

	// Last push
	if state.LastPushHash != "" {
		fmt.Printf("LAST DEPLOY: %s — %s\n", state.LastPushTime, state.LastPushHash)
	}

	return nil
}

func statusIcon(s git.AgentStatus) string {
	switch s {
	case git.StatusStaged:
		return "\u2713" // checkmark
	case git.StatusWorking:
		return "\u2022" // bullet
	default:
		return "\u25cb" // empty circle
	}
}
