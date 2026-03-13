package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vibeyang/multitab/internal/detect"
	gitops "github.com/vibeyang/multitab/internal/git"
	"github.com/vibeyang/multitab/internal/queue"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Non-interactive batch push (for scripts/CI)",
	RunE:  runPush,
}

func runPush(cmd *cobra.Command, args []string) error {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return fmt.Errorf("not inside a git repository")
	}

	state, err := queue.Refresh(repoRoot)
	if err != nil {
		return err
	}

	if len(state.StagedCommits) == 0 {
		fmt.Println("Nothing to push — no staged commits ahead of origin/main.")
		return nil
	}

	// Warn about conflicts
	if len(state.Conflicts) > 0 {
		fmt.Fprintf(os.Stderr, "WARNING: %d file conflicts detected between agents.\n", len(state.Conflicts))
		for _, c := range state.Conflicts {
			fmt.Fprintf(os.Stderr, "  %s: %s <-> %s\n", c.File, c.Agent1, c.Agent2)
		}
	}

	fmt.Printf("Pushing %d commit(s) to origin/main...\n\n", len(state.StagedCommits))

	// Step 1: Fetch
	fmt.Print("  Fetching origin/main... ")
	if err := gitops.Fetch(repoRoot); err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("OK")

	// Step 2: Rebase
	fmt.Print("  Rebasing onto latest... ")
	if err := gitops.Rebase(repoRoot); err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("OK")

	// Step 3: Build
	buildCmd := detect.DetectBuildCommand(repoRoot)
	if buildCmd != "" {
		fmt.Printf("  Running build (%s)... ", buildCmd)
		if err := gitops.RunBuild(repoRoot, buildCmd); err != nil {
			fmt.Println("FAILED")
			return err
		}
		fmt.Println("OK")
	}

	// Step 4: Push
	fmt.Print("  Pushing to origin/main... ")
	if err := gitops.Push(repoRoot); err != nil {
		fmt.Println("FAILED")
		return err
	}
	fmt.Println("OK")

	fmt.Printf("\nPushed %d commit(s) successfully.\n", len(state.StagedCommits))
	for _, c := range state.StagedCommits {
		fmt.Printf("  %s %s\n", c.Hash, c.Message)
	}

	return nil
}
