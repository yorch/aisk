package adapter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yorch/aisk/internal/skill"
)

// CursorAdapter writes skills as .mdc files with Cursor's YAML frontmatter.
type CursorAdapter struct{}

func (a *CursorAdapter) Install(s *skill.Skill, targetPath string, opts InstallOpts) error {
	content, err := a.buildContent(s, opts.IncludeRefs)
	if err != nil {
		return err
	}

	// Ensure rules directory exists
	if err := os.MkdirAll(targetPath, 0o755); err != nil {
		return fmt.Errorf("creating rules dir: %w", err)
	}

	dest := filepath.Join(targetPath, s.DirName+".mdc")
	return os.WriteFile(dest, []byte(content), 0o644)
}

func (a *CursorAdapter) Uninstall(s *skill.Skill, targetPath string) error {
	dest := filepath.Join(targetPath, s.DirName+".mdc")
	err := os.Remove(dest)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (a *CursorAdapter) Describe(s *skill.Skill, targetPath string, opts InstallOpts) string {
	dest := filepath.Join(targetPath, s.DirName+".mdc")
	return fmt.Sprintf("write %s", dest)
}

func (a *CursorAdapter) buildContent(s *skill.Skill, includeRefs bool) (string, error) {
	// Truncate description for frontmatter (first line only, max 200 chars)
	desc := strings.Split(s.Frontmatter.Description, "\n")[0]
	if len(desc) > 200 {
		desc = desc[:197] + "..."
	}

	var b strings.Builder

	// Cursor .mdc frontmatter
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("description: %s\n", desc))
	b.WriteString("globs:\n")
	b.WriteString("alwaysApply: false\n")
	b.WriteString("---\n\n")

	// Body
	body := s.MarkdownBody
	if includeRefs {
		fullContent, err := skill.ReadFullContent(s, true)
		if err != nil {
			return "", err
		}
		body = fullContent
	}

	b.WriteString(body)

	return b.String(), nil
}
