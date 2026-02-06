package adapter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yorch/aisk/internal/skill"
)

// WindsurfAdapter handles both project-level (individual files) and
// global-level (appended sections) installations for Windsurf.
type WindsurfAdapter struct{}

func (a *WindsurfAdapter) Install(s *skill.Skill, targetPath string, opts InstallOpts) error {
	body := s.MarkdownBody
	if opts.IncludeRefs {
		fullContent, err := skill.ReadFullContent(s, true)
		if err != nil {
			return err
		}
		body = fullContent
	}

	if opts.Scope == "global" {
		// Append to global rules file using section markers
		content := fmt.Sprintf("# %s\n\n%s", s.Frontmatter.Name, body)
		dir := filepath.Dir(targetPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating dir: %w", err)
		}
		return appendOrReplaceSection(targetPath, s.Frontmatter.Name, content)
	}

	// Project-level: write individual file
	if err := os.MkdirAll(targetPath, 0o755); err != nil {
		return fmt.Errorf("creating rules dir: %w", err)
	}

	dest := filepath.Join(targetPath, s.DirName+".md")
	content := fmt.Sprintf("# %s\n\n%s", s.Frontmatter.Name, body)
	return os.WriteFile(dest, []byte(content), 0o644)
}

func (a *WindsurfAdapter) Uninstall(s *skill.Skill, targetPath string) error {
	// Try project-level file first
	projectFile := filepath.Join(targetPath, s.DirName+".md")
	if _, err := os.Stat(projectFile); err == nil {
		return os.Remove(projectFile)
	}

	// Try global-level section removal
	if strings.HasSuffix(targetPath, ".md") {
		return removeSection(targetPath, s.Frontmatter.Name)
	}

	return nil
}

func (a *WindsurfAdapter) Describe(s *skill.Skill, targetPath string, opts InstallOpts) string {
	if opts.Scope == "global" {
		return fmt.Sprintf("append skill section to %s", targetPath)
	}
	dest := filepath.Join(targetPath, s.DirName+".md")
	return fmt.Sprintf("write %s", dest)
}
