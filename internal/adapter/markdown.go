package adapter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yorch/aisk/internal/skill"
)

// MarkdownAdapter consolidates a skill into a markdown section appended to a file.
// Used by Gemini CLI, Codex CLI, and VS Code Copilot.
type MarkdownAdapter struct {
	ClientName string
}

func (a *MarkdownAdapter) Install(s *skill.Skill, targetPath string, opts InstallOpts) error {
	content, err := a.buildContent(s, opts.IncludeRefs)
	if err != nil {
		return err
	}

	// Ensure parent directory
	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating parent dir: %w", err)
	}

	return appendOrReplaceSection(targetPath, s.Frontmatter.Name, content)
}

func (a *MarkdownAdapter) Uninstall(s *skill.Skill, targetPath string) error {
	return removeSection(targetPath, s.Frontmatter.Name)
}

func (a *MarkdownAdapter) Describe(s *skill.Skill, targetPath string, opts InstallOpts) string {
	return fmt.Sprintf("append skill section to %s", targetPath)
}

func (a *MarkdownAdapter) buildContent(s *skill.Skill, includeRefs bool) (string, error) {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# %s\n\n", s.Frontmatter.Name))

	if s.Frontmatter.Description != "" {
		// Render description as blockquote
		for _, line := range strings.Split(s.Frontmatter.Description, "\n") {
			b.WriteString("> " + strings.TrimSpace(line) + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(s.MarkdownBody)

	if includeRefs {
		fullContent, err := skill.ReadFullContent(s, true)
		if err != nil {
			return "", err
		}
		// ReadFullContent already includes the body, so use it directly
		b.Reset()
		b.WriteString(fmt.Sprintf("# %s\n\n", s.Frontmatter.Name))
		if s.Frontmatter.Description != "" {
			for _, line := range strings.Split(s.Frontmatter.Description, "\n") {
				b.WriteString("> " + strings.TrimSpace(line) + "\n")
			}
			b.WriteString("\n")
		}
		b.WriteString(fullContent)
	}

	return b.String(), nil
}

// Section markers for idempotent appends.
func sectionStart(name string) string { return fmt.Sprintf("<!-- aisk:start:%s -->", name) }
func sectionEnd(name string) string   { return fmt.Sprintf("<!-- aisk:end:%s -->", name) }

// appendOrReplaceSection adds or replaces a skill section in a markdown file.
func appendOrReplaceSection(filePath, skillName, content string) error {
	startMarker := sectionStart(skillName)
	endMarker := sectionEnd(skillName)
	wrapped := fmt.Sprintf("%s\n%s\n%s", startMarker, content, endMarker)

	existing, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create new file
			return os.WriteFile(filePath, []byte(wrapped+"\n"), 0o644)
		}
		return err
	}

	fileContent := string(existing)

	// Check if section already exists
	startIdx := strings.Index(fileContent, startMarker)
	endIdx := strings.Index(fileContent, endMarker)

	if startIdx >= 0 && endIdx >= 0 {
		// Replace existing section
		newContent := fileContent[:startIdx] + wrapped + fileContent[endIdx+len(endMarker):]
		return os.WriteFile(filePath, []byte(newContent), 0o644)
	}

	// Append new section
	if len(fileContent) > 0 && !strings.HasSuffix(fileContent, "\n") {
		fileContent += "\n"
	}
	fileContent += "\n" + wrapped + "\n"
	return os.WriteFile(filePath, []byte(fileContent), 0o644)
}

// removeSection removes a skill section from a markdown file.
func removeSection(filePath, skillName string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // nothing to remove
		}
		return err
	}

	fileContent := string(data)
	startMarker := sectionStart(skillName)
	endMarker := sectionEnd(skillName)

	startIdx := strings.Index(fileContent, startMarker)
	endIdx := strings.Index(fileContent, endMarker)

	if startIdx < 0 || endIdx < 0 {
		return nil // section not found
	}

	// Remove section including surrounding whitespace
	before := strings.TrimRight(fileContent[:startIdx], "\n")
	after := strings.TrimLeft(fileContent[endIdx+len(endMarker):], "\n")

	var newContent string
	if before == "" && after == "" {
		newContent = ""
	} else if before == "" {
		newContent = after + "\n"
	} else if after == "" {
		newContent = before + "\n"
	} else {
		newContent = before + "\n\n" + after + "\n"
	}

	return os.WriteFile(filePath, []byte(newContent), 0o644)
}
