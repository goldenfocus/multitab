package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// AgentStatus represents the current state of a worktree agent.
type AgentStatus int

const (
	StatusIdle    AgentStatus = iota // no activity
	StatusWorking                    // uncommitted changes or unpushed commits
	StatusStaged                     // branch merged into local main, ready to push
)

func (s AgentStatus) String() string {
	switch s {
	case StatusIdle:
		return "IDLE"
	case StatusWorking:
		return "WORKING"
	case StatusStaged:
		return "STAGED"
	default:
		return "UNKNOWN"
	}
}

// Agent represents a single worktree (one Claude Code session).
type Agent struct {
	Name       string
	Path       string
	Branch     string
	Status     AgentStatus
	Commits    int    // number of commits ahead of origin/main
	Files      int    // number of changed files
	LastCommit string // short hash of last commit
}

// DiscoverWorktrees finds all git worktrees in the repo.
func DiscoverWorktrees(repoRoot string) ([]Agent, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git worktree list: %w", err)
	}

	var agents []Agent
	scanner := bufio.NewScanner(strings.NewReader(string(out)))

	var current Agent
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "worktree "):
			current.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "branch "):
			ref := strings.TrimPrefix(line, "branch ")
			current.Branch = strings.TrimPrefix(ref, "refs/heads/")
		case line == "":
			if current.Path != "" {
				// Skip the main worktree itself
				if current.Branch == "main" && current.Path == repoRoot {
					current = Agent{}
					continue
				}
				current.Name = deriveAgentName(current.Path, current.Branch)
				agents = append(agents, current)
			}
			current = Agent{}
		}
	}
	// Handle last entry (no trailing blank line)
	if current.Path != "" && !(current.Branch == "main" && current.Path == repoRoot) {
		current.Name = deriveAgentName(current.Path, current.Branch)
		agents = append(agents, current)
	}

	return agents, nil
}

// InspectAgent populates status, commit count, and file count for an agent.
func InspectAgent(repoRoot string, agent *Agent) error {
	// Count uncommitted changes
	statusCmd := exec.Command("git", "status", "--short")
	statusCmd.Dir = agent.Path
	statusOut, err := statusCmd.Output()
	if err != nil {
		return fmt.Errorf("git status in %s: %w", agent.Path, err)
	}
	dirtyFiles := countLines(string(statusOut))

	// Count commits ahead of origin/main
	logCmd := exec.Command("git", "log", "origin/main..HEAD", "--oneline")
	logCmd.Dir = agent.Path
	logOut, err := logCmd.Output()
	if err != nil {
		// If origin/main doesn't exist, treat as 0
		logOut = []byte{}
	}
	commitCount := countLines(string(logOut))

	// Get last commit hash
	hashCmd := exec.Command("git", "log", "-1", "--format=%h")
	hashCmd.Dir = agent.Path
	hashOut, err := hashCmd.Output()
	if err == nil {
		agent.LastCommit = strings.TrimSpace(string(hashOut))
	}

	// Count unique files changed in commits ahead of origin/main
	diffCmd := exec.Command("git", "diff", "--name-only", "origin/main..HEAD")
	diffCmd.Dir = agent.Path
	diffOut, err := diffCmd.Output()
	if err != nil {
		diffOut = []byte{}
	}
	committedFiles := countLines(string(diffOut))

	agent.Commits = commitCount
	agent.Files = committedFiles + dirtyFiles

	// Determine status
	if isStaged(repoRoot, agent.Branch) {
		agent.Status = StatusStaged
	} else if dirtyFiles > 0 || commitCount > 0 {
		agent.Status = StatusWorking
	} else {
		agent.Status = StatusIdle
	}

	return nil
}

// isStaged checks if a branch has been merged into local main.
func isStaged(repoRoot, branch string) bool {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", branch, "main")
	cmd.Dir = repoRoot
	err := cmd.Run()
	return err == nil
}

// deriveAgentName extracts a short name from the worktree path or branch.
func deriveAgentName(path, branch string) string {
	// Use branch name if available and not "main"
	if branch != "" && branch != "main" {
		return branch
	}
	// Fall back to last directory segment
	return filepath.Base(path)
}

func countLines(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}
