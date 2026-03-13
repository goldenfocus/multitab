package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// AgentStatus represents the current state of a worktree agent.
type AgentStatus int

const (
	StatusIdle      AgentStatus = iota // no activity
	StatusWorking                      // uncommitted changes or unpushed commits
	StatusStaged                       // branch merged into local main, ready to push
	StatusStale                        // inactive for 30+ min with unpushed work
	StatusAbandoned                    // all work already on origin/main, just needs cleanup
)

func (s AgentStatus) String() string {
	switch s {
	case StatusIdle:
		return "IDLE"
	case StatusWorking:
		return "WORKING"
	case StatusStaged:
		return "STAGED"
	case StatusStale:
		return "STALE"
	case StatusAbandoned:
		return "ABANDONED"
	default:
		return "UNKNOWN"
	}
}

// Agent represents a single worktree (one Claude Code session).
type Agent struct {
	Name           string
	Path           string
	Branch         string
	Status         AgentStatus
	Commits        int      // number of commits ahead of origin/main
	Files          int      // number of changed files
	DirtyFiles     int      // uncommitted changes only
	LastCommit     string   // short hash of last commit
	LastCommitTime time.Time
	LastCommitMsg  string   // subject line of last commit
	CommitMessages []string // all commit messages (for intel)
	ChangedFiles   []string // list of changed file paths (for intel)
	AlreadyPushed  bool     // true if all commits exist on origin/main
	StaleFor       time.Duration

	// Conversation intel (from Claude JSONL)
	LastPrompt    string    // last human message text
	LastPromptAt  time.Time // when it was sent
	HumanMsgCount int       // total human messages in session
	SessionID     string    // Claude session ID
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
	if current.Path != "" && !(current.Branch == "main" && current.Path == repoRoot) {
		current.Name = deriveAgentName(current.Path, current.Branch)
		agents = append(agents, current)
	}

	return agents, nil
}

// InspectAgent populates status, commit count, file count, and intel for an agent.
func InspectAgent(repoRoot string, agent *Agent) error {
	// Count uncommitted changes
	statusCmd := exec.Command("git", "status", "--short")
	statusCmd.Dir = agent.Path
	statusOut, err := statusCmd.Output()
	if err != nil {
		return fmt.Errorf("git status in %s: %w", agent.Path, err)
	}
	agent.DirtyFiles = countLines(string(statusOut))

	// Count commits ahead of origin/main
	logCmd := exec.Command("git", "log", "origin/main..HEAD", "--oneline")
	logCmd.Dir = agent.Path
	logOut, err := logCmd.Output()
	if err != nil {
		logOut = []byte{}
	}
	agent.Commits = countLines(string(logOut))

	// Gather commit messages for intel
	if agent.Commits > 0 {
		msgCmd := exec.Command("git", "log", "origin/main..HEAD", "--format=%s")
		msgCmd.Dir = agent.Path
		msgOut, err := msgCmd.Output()
		if err == nil {
			for _, line := range strings.Split(strings.TrimSpace(string(msgOut)), "\n") {
				if line != "" {
					agent.CommitMessages = append(agent.CommitMessages, line)
				}
			}
		}
	}

	// Get last commit hash, time, and message
	hashCmd := exec.Command("git", "log", "-1", "--format=%h\t%aI\t%s")
	hashCmd.Dir = agent.Path
	hashOut, err := hashCmd.Output()
	if err == nil {
		parts := strings.SplitN(strings.TrimSpace(string(hashOut)), "\t", 3)
		if len(parts) >= 1 {
			agent.LastCommit = parts[0]
		}
		if len(parts) >= 2 {
			if t, err := time.Parse(time.RFC3339, parts[1]); err == nil {
				agent.LastCommitTime = t
			}
		}
		if len(parts) >= 3 {
			agent.LastCommitMsg = parts[2]
		}
	}

	// Get changed files (committed + uncommitted)
	diffCmd := exec.Command("git", "diff", "--name-only", "origin/main..HEAD")
	diffCmd.Dir = agent.Path
	diffOut, err := diffCmd.Output()
	if err != nil {
		diffOut = []byte{}
	}
	committedFiles := countLines(string(diffOut))

	// Collect file list for intel
	for _, f := range strings.Split(strings.TrimSpace(string(diffOut)), "\n") {
		if f != "" {
			agent.ChangedFiles = append(agent.ChangedFiles, f)
		}
	}
	// Add uncommitted files
	for _, line := range strings.Split(strings.TrimSpace(string(statusOut)), "\n") {
		if len(line) > 3 {
			f := strings.TrimSpace(line[2:])
			agent.ChangedFiles = appendUnique(agent.ChangedFiles, f)
		}
	}

	agent.Files = committedFiles + agent.DirtyFiles

	// Check if all work is already on origin/main
	agent.AlreadyPushed = checkAlreadyPushed(agent)

	// Determine status — order matters!
	hasWork := agent.Commits > 0 || agent.DirtyFiles > 0

	if !hasWork {
		// No commits, no dirty files = nothing to do here
		agent.Status = StatusIdle
	} else if isStaged(repoRoot, agent.Branch) {
		agent.Status = StatusStaged
	} else if agent.AlreadyPushed && agent.DirtyFiles == 0 {
		agent.Status = StatusAbandoned
	} else if isStaleAgent(agent) {
		agent.Status = StatusStale
		if !agent.LastCommitTime.IsZero() {
			agent.StaleFor = time.Since(agent.LastCommitTime)
		}
	} else {
		agent.Status = StatusWorking
	}

	return nil
}

// checkAlreadyPushed returns true if every commit on this branch exists on origin/main.
func checkAlreadyPushed(agent *Agent) bool {
	if agent.Commits == 0 && agent.DirtyFiles == 0 {
		return true
	}
	if agent.DirtyFiles > 0 {
		return false // uncommitted work is never "pushed"
	}
	// Check if HEAD is an ancestor of origin/main (all commits already on remote)
	cmd := exec.Command("git", "merge-base", "--is-ancestor", "HEAD", "origin/main")
	cmd.Dir = agent.Path
	return cmd.Run() == nil
}

// isStaleAgent checks if the agent hasn't had activity in 30+ minutes.
func isStaleAgent(agent *Agent) bool {
	if agent.LastCommitTime.IsZero() {
		return false
	}
	// If there are uncommitted changes, check last commit time
	// If there are only committed changes, check last commit time
	if agent.Commits > 0 || agent.DirtyFiles > 0 {
		return time.Since(agent.LastCommitTime) > 30*time.Minute
	}
	return false
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
	if branch != "" && branch != "main" {
		return branch
	}
	return filepath.Base(path)
}

func countLines(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
}

func appendUnique(slice []string, s string) []string {
	for _, v := range slice {
		if v == s {
			return slice
		}
	}
	return append(slice, s)
}
