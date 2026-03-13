package queue

import (
	"github.com/vibeyang/multitab/internal/git"
)

// Conflict represents two agents modifying the same file.
type Conflict struct {
	File   string
	Agent1 string
	Agent2 string
}

// DetectConflicts checks if any working/staged agents have overlapping file changes.
func DetectConflicts(agents []git.Agent) []Conflict {
	// Build a map of file → agent names
	fileOwners := make(map[string][]string)
	for _, agent := range agents {
		if agent.Status == git.StatusIdle {
			continue
		}
		files, err := git.ChangedFiles(agent.Path)
		if err != nil {
			continue
		}
		for _, f := range files {
			fileOwners[f] = append(fileOwners[f], agent.Name)
		}
	}

	// Find files owned by 2+ agents
	var conflicts []Conflict
	for file, owners := range fileOwners {
		if len(owners) < 2 {
			continue
		}
		for i := 0; i < len(owners)-1; i++ {
			for j := i + 1; j < len(owners); j++ {
				conflicts = append(conflicts, Conflict{
					File:   file,
					Agent1: owners[i],
					Agent2: owners[j],
				})
			}
		}
	}
	return conflicts
}
