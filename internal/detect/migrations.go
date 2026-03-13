package detect

import (
	"strings"

	"github.com/goldenfocus/multitab/internal/git"
)

// DetectMigrations checks if any agents have changes in supabase/migrations/.
func DetectMigrations(agents []git.Agent) bool {
	for _, agent := range agents {
		if agent.Status == git.StatusIdle {
			continue
		}
		files, err := git.ChangedFiles(agent.Path)
		if err != nil {
			continue
		}
		for _, f := range files {
			if strings.HasPrefix(f, "supabase/migrations/") {
				return true
			}
		}
	}
	return false
}
