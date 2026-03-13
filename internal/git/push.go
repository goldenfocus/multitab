package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// PushStep represents a step in the push sequence.
type PushStep int

const (
	StepFetch PushStep = iota
	StepRebase
	StepBuild
	StepPush
	StepVerify
	StepDone
)

func (s PushStep) String() string {
	switch s {
	case StepFetch:
		return "Fetching origin/main..."
	case StepRebase:
		return "Rebasing onto latest..."
	case StepBuild:
		return "Running build..."
	case StepPush:
		return "Pushing to origin/main..."
	case StepVerify:
		return "Verifying push..."
	case StepDone:
		return "Done!"
	default:
		return "Unknown"
	}
}

// Fetch runs git fetch origin main.
func Fetch(repoRoot string) error {
	cmd := exec.Command("git", "fetch", "origin", "main")
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("fetch: %s", string(out))
	}
	return nil
}

// Rebase runs git rebase origin/main on the current branch.
func Rebase(repoRoot string) error {
	cmd := exec.Command("git", "rebase", "origin/main")
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rebase: %s\nResolve conflicts and try again", string(out))
	}
	return nil
}

// Push pushes the current branch to origin/main.
func Push(repoRoot string) error {
	cmd := exec.Command("git", "push", "origin", "main")
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("push: %s", string(out))
	}
	return nil
}

// RunBuild executes the build command and returns any error.
func RunBuild(repoRoot, buildCmd string) error {
	parts := strings.Fields(buildCmd)
	if len(parts) == 0 {
		return nil // no build command configured
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build failed: %s", string(out))
	}
	return nil
}

// CleanupWorktree removes a worktree and its branch.
func CleanupWorktree(repoRoot, worktreePath, branch string) error {
	removeCmd := exec.Command("git", "worktree", "remove", worktreePath)
	removeCmd.Dir = repoRoot
	if out, err := removeCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("worktree remove: %s", string(out))
	}

	if branch != "" && branch != "main" {
		branchCmd := exec.Command("git", "branch", "-D", branch)
		branchCmd.Dir = repoRoot
		branchCmd.CombinedOutput() // best effort
	}
	return nil
}
