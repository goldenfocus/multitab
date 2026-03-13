package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/goldenfocus/multitab/internal/git"
	"github.com/goldenfocus/multitab/internal/queue"
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

	fmt.Println("MULTITAB STATUS")
	fmt.Println("\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550\u2550")

	if len(state.Agents) == 0 {
		fmt.Println("No worktrees found.")
		return nil
	}

	fmt.Printf("\n%-28s %-16s %-8s %-6s\n", "AGENT", "STATUS", "COMMITS", "FILES")
	fmt.Println("\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500")

	for _, agent := range state.Agents {
		icon := statusIcon(agent.Status)
		fmt.Printf("%s %-26s %-16s %-8d %-6d\n",
			icon, agent.Name, agent.Status, agent.Commits, agent.Files)

		// Show intel for stale/abandoned
		if agent.Status == git.StatusStale {
			ago := formatStaleTimeText(agent.StaleFor)
			fmt.Printf("  \u2514 inactive %s", ago)
			if agent.Commits > 0 {
				fmt.Printf(", %d unpushed commit(s)", agent.Commits)
			}
			fmt.Println()
		} else if agent.Status == git.StatusAbandoned {
			fmt.Println("  \u2514 all work already pushed, safe to discard")
		}
	}

	if len(state.StagedCommits) > 0 {
		fmt.Printf("\nSTAGED COMMITS (%d):\n", len(state.StagedCommits))
		for _, c := range state.StagedCommits {
			fmt.Printf("  %s %s\n", c.Hash, c.Message)
		}
	}

	fmt.Printf("\nDEPLOY QUEUE: %d/%d ready\n", state.ReadyCount, state.TotalCount)

	if len(state.Conflicts) > 0 {
		fmt.Printf("\nCONFLICTS (%d):\n", len(state.Conflicts))
		for _, c := range state.Conflicts {
			fmt.Printf("  %s: %s <-> %s\n", c.File, c.Agent1, c.Agent2)
		}
	} else {
		fmt.Println("\nCONFLICTS: None")
	}

	if state.HasMigrations {
		fmt.Println("MIGRATIONS: Pending changes detected")
	}

	if state.LastPushHash != "" {
		fmt.Printf("LAST DEPLOY: %s \u2014 %s\n", state.LastPushTime, state.LastPushHash)
	}

	return nil
}

func statusIcon(s git.AgentStatus) string {
	switch s {
	case git.StatusStaged:
		return "\u2713"
	case git.StatusWorking:
		return "\u25cf"
	case git.StatusStale:
		return "\u25cc"
	case git.StatusAbandoned:
		return "\u2205"
	default:
		return "\u25cb"
	}
}

func formatStaleTimeText(d time.Duration) string {
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%d min", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%d hours", int(d.Hours()))
	}
	return fmt.Sprintf("%d days", int(d.Hours()/24))
}
