package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// StagedCommit represents a commit on local main that hasn't been pushed.
type StagedCommit struct {
	Hash    string
	Message string
}

// GetStagedCommits returns commits on local main ahead of origin/main.
func GetStagedCommits(repoRoot string) ([]StagedCommit, error) {
	cmd := exec.Command("git", "log", "origin/main..main", "--oneline")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}

	var commits []StagedCommit
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		msg := ""
		if len(parts) > 1 {
			msg = parts[1]
		}
		commits = append(commits, StagedCommit{
			Hash:    parts[0],
			Message: msg,
		})
	}
	return commits, nil
}

// MergeBranch merges a branch into the current branch (used for staging).
func MergeBranch(repoRoot, branch string) error {
	cmd := exec.Command("git", "merge", "--no-edit", branch)
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git merge %s: %s", branch, string(out))
	}
	return nil
}

// GetLastPushInfo returns the short hash and relative time of origin/main.
func GetLastPushInfo(repoRoot string) (hash string, timeAgo string, err error) {
	hashCmd := exec.Command("git", "log", "-1", "--format=%h", "origin/main")
	hashCmd.Dir = repoRoot
	hashOut, err := hashCmd.Output()
	if err != nil {
		return "", "", err
	}
	hash = strings.TrimSpace(string(hashOut))

	timeCmd := exec.Command("git", "log", "-1", "--format=%cr", "origin/main")
	timeCmd.Dir = repoRoot
	timeOut, err := timeCmd.Output()
	if err != nil {
		return hash, "", err
	}
	timeAgo = strings.TrimSpace(string(timeOut))

	return hash, timeAgo, nil
}
