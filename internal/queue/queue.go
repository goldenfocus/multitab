package queue

import (
	"sync"

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

	// Inspect all agents in parallel
	var wg sync.WaitGroup
	for i := range agents {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if err := git.InspectAgent(repoRoot, &agents[idx]); err != nil {
				agents[idx].Status = git.StatusIdle
			}
			if conv, err := agent.FindConversation(agents[idx].Path); err == nil && conv != nil {
				agents[idx].LastPrompt = conv.LastPrompt
				agents[idx].LastPromptAt = conv.LastPromptAt
				agents[idx].HumanMsgCount = conv.MessageCount
				agents[idx].SessionID = conv.SessionID
			}
		}(i)
	}
	wg.Wait()
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
