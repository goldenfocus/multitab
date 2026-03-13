package git

import (
	"os/exec"
	"strings"
)

// ChangedFiles returns the list of files changed in a branch relative to origin/main.
func ChangedFiles(worktreePath string) ([]string, error) {
	// Committed changes
	diffCmd := exec.Command("git", "diff", "--name-only", "origin/main..HEAD")
	diffCmd.Dir = worktreePath
	diffOut, err := diffCmd.Output()
	if err != nil {
		diffOut = []byte{}
	}

	// Uncommitted changes
	statusCmd := exec.Command("git", "diff", "--name-only")
	statusCmd.Dir = worktreePath
	statusOut, err := statusCmd.Output()
	if err != nil {
		statusOut = []byte{}
	}

	// Staged but not committed
	stagedCmd := exec.Command("git", "diff", "--name-only", "--cached")
	stagedCmd.Dir = worktreePath
	stagedOut, err := stagedCmd.Output()
	if err != nil {
		stagedOut = []byte{}
	}

	seen := make(map[string]bool)
	var files []string
	for _, out := range [][]byte{diffOut, statusOut, stagedOut} {
		for _, f := range strings.Split(strings.TrimSpace(string(out)), "\n") {
			if f != "" && !seen[f] {
				seen[f] = true
				files = append(files, f)
			}
		}
	}
	return files, nil
}
