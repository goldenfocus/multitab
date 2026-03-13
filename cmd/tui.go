package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/goldenfocus/multitab/internal/detect"
	"github.com/goldenfocus/multitab/internal/tui"
)

func runTUI(cmd *cobra.Command, args []string) error {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return fmt.Errorf("not inside a git repository")
	}

	buildCmd := detect.DetectBuildCommand(repoRoot)

	model := tui.NewModel(repoRoot, buildCmd)
	p := tea.NewProgram(model, tea.WithAltScreen())

	_, err = p.Run()
	return err
}
