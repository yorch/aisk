package adapter

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/yorch/aisk/internal/skill"
)

// ClaudeAdapter installs skills for Claude Code by symlinking or copying the directory.
type ClaudeAdapter struct{}

func (a *ClaudeAdapter) Install(s *skill.Skill, targetPath string, opts InstallOpts) error {
	dest := filepath.Join(targetPath, s.DirName)

	// Ensure parent directory exists
	if err := os.MkdirAll(targetPath, 0o755); err != nil {
		return fmt.Errorf("creating target dir: %w", err)
	}

	// Remove existing installation if present
	if err := os.RemoveAll(dest); err != nil {
		return fmt.Errorf("removing existing: %w", err)
	}

	if s.Source == skill.SourceLocal {
		// Symlink for local skills
		if err := os.Symlink(s.Path, dest); err != nil {
			return fmt.Errorf("creating symlink: %w", err)
		}
	} else {
		// Copy for remote skills
		if err := copyDir(s.Path, dest); err != nil {
			return fmt.Errorf("copying skill: %w", err)
		}
	}

	return nil
}

func (a *ClaudeAdapter) Uninstall(s *skill.Skill, targetPath string) error {
	dest := filepath.Join(targetPath, s.DirName)
	return os.RemoveAll(dest)
}

func (a *ClaudeAdapter) Describe(s *skill.Skill, targetPath string, opts InstallOpts) string {
	dest := filepath.Join(targetPath, s.DirName)
	if s.Source == skill.SourceLocal {
		return fmt.Sprintf("symlink %s -> %s", dest, s.Path)
	}
	return fmt.Sprintf("copy %s -> %s", s.Path, dest)
}

// copyDir recursively copies a directory tree.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}
