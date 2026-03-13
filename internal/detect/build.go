package detect

import (
	"os"
	"path/filepath"
)

// DetectBuildCommand tries to find the right build command for the repo.
func DetectBuildCommand(repoRoot string) string {
	// Check for our custom safe-build script
	if fileExists(filepath.Join(repoRoot, "scripts", "safe-build.sh")) {
		return "bash scripts/safe-build.sh"
	}
	// Check for Makefile
	if fileExists(filepath.Join(repoRoot, "Makefile")) {
		return "make build"
	}
	// Check for package.json (Node project)
	if fileExists(filepath.Join(repoRoot, "package.json")) {
		return "npm run build"
	}
	// Check for go.mod (Go project)
	if fileExists(filepath.Join(repoRoot, "go.mod")) {
		return "go build ./..."
	}
	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
