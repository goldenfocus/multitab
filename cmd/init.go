package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/vibeyang/multitab/internal/detect"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize multitab config for this repo",
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return fmt.Errorf("not inside a git repository")
	}

	configPath := filepath.Join(repoRoot, ".multitab.toml")

	// Check if already exists
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Config already exists at %s\n", configPath)
		return nil
	}

	buildCmd := detect.DetectBuildCommand(repoRoot)

	config := fmt.Sprintf(`# multitab configuration
# Auto-generated — edit as needed

[repo]
main_branch = "main"
worktree_dir = ".claude/worktrees"

[build]
command = %q
timeout = "10m"

[push]
pre_push_hooks = true
auto_cleanup = true

[tui]
refresh_interval = "5s"
`, buildCmd)

	if err := os.WriteFile(configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	fmt.Printf("Created %s\n", configPath)
	fmt.Printf("Build command: %s\n", buildCmd)
	fmt.Println("\nRun `multitab` to launch the dashboard.")

	return nil
}
