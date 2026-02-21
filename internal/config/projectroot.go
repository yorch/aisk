package config

import (
	"os"
	"path/filepath"
)

// projectMarkers are files/directories that indicate a project root.
var projectMarkers = []string{
	".git",
	"package.json",
	"go.mod",
	"Cargo.toml",
	"pyproject.toml",
}

// FindProjectRoot walks up from startDir looking for a project root marker.
// Returns the project root path, or "" if not found.
func FindProjectRoot(startDir string) string {
	dir := startDir
	for {
		for _, marker := range projectMarkers {
			p := filepath.Join(dir, marker)
			if _, err := os.Stat(p); err == nil {
				return dir
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return ""
		}
		dir = parent
	}
}
