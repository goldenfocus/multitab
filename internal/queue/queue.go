package queue

import (
	"github.com/goldenfocus/multitab/internal/agent"
	"github.com/goldenfocus/multitab/internal/detect"
	"github.com/goldenfocus/multitab/internal/git"
)

// State holds the complete dashboard state.
type State struct {
	RepoRoot      string
	Agents        []git.Agent
	StagedCommits []git.StagedCommit
	Conflicts     []Conflict
	HasMigrations bool
	LastPushHash  string
	LastPushTime  string
	ReadyCount    int
	TotalCount    int
}

// Refresh re-scans the repository and rebuilds the queue state.
func Refresh(repoRoot string) (*State, error) {
	state := &State{RepoRoot: repoRoot}

	// Discover worktrees
	agents, err := git.DiscoverWorktrees(repoRoot)
	if err != nil {
		return nil, err
	}

	// Inspect each agent
	for i := range agents {
		if err := git.InspectAgent(repoRoot, &agents[i]); err != nil {
			// Non-fatal: mark as idle with error
			agents[i].Status = git.StatusIdle
		}
		// Fetch conversation info from Claude JSONL
		if conv, err := agent.FindConversation(agents[i].Path); err == nil && conv != nil {
			agents[i].LastPrompt = conv.LastPrompt
			agents[i].LastPromptAt = conv.LastPromptAt
			agents[i].HumanMsgCount = conv.MessageCount
			agents[i].SessionID = conv.SessionID
		}
	}
	state.Agents = agents

	// Count ready vs total
	for _, a := range agents {
		state.TotalCount++
		if a.Status == git.StatusStaged {
			state.ReadyCount++
		}
	}

	// Get staged commits
	commits, err := git.GetStagedCommits(repoRoot)
	if err == nil {
		state.StagedCommits = commits
	}

	// Detect conflicts
	state.Conflicts = DetectConflicts(agents)

	// Detect migrations
	state.HasMigrations = detect.DetectMigrations(agents)

	// Last push info
	hash, timeAgo, err := git.GetLastPushInfo(repoRoot)
	if err == nil {
		state.LastPushHash = hash
		state.LastPushTime = timeAgo
	}

	return state, nil
}
