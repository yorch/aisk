package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const skillMDTemplate = `---
name: {{.Name}}
description: TODO — describe what this skill does
version: 0.1.0
---
# {{.Title}}

Use when: TODO — describe when this skill should be activated.

## Instructions

TODO — write the skill instructions here.
`

const readmeTemplate = `# {{.Title}}

TODO — describe this skill for humans.
`

// Scaffold creates a new skill directory with template files.
func Scaffold(parentDir, name string) (string, error) {
	if err := ValidateName(name); err != nil {
		return "", fmt.Errorf("invalid skill name: %w", err)
	}

	skillDir := filepath.Join(parentDir, name)

	// Check if directory already exists
	if _, err := os.Stat(skillDir); err == nil {
		return "", fmt.Errorf("directory already exists: %s", skillDir)
	}

	// Create the directory structure
	dirs := []string{
		skillDir,
		filepath.Join(skillDir, "reference"),
		filepath.Join(skillDir, "examples"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			cleanup(skillDir)
			return "", fmt.Errorf("creating directory %s: %w", d, err)
		}
	}

	// Template data
	data := struct {
		Name  string
		Title string
	}{
		Name:  name,
		Title: kebabToTitle(name),
	}

	// Write SKILL.md
	if err := writeTemplate(filepath.Join(skillDir, "SKILL.md"), skillMDTemplate, data); err != nil {
		cleanup(skillDir)
		return "", fmt.Errorf("writing SKILL.md: %w", err)
	}

	// Write README.md
	if err := writeTemplate(filepath.Join(skillDir, "README.md"), readmeTemplate, data); err != nil {
		cleanup(skillDir)
		return "", fmt.Errorf("writing README.md: %w", err)
	}

	return skillDir, nil
}

func writeTemplate(path, tmplText string, data any) error {
	tmpl, err := template.New("").Parse(tmplText)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return tmpl.Execute(f, data)
}

func cleanup(dir string) {
	os.RemoveAll(dir)
}

// kebabToTitle converts "my-cool-skill" to "My Cool Skill".
func kebabToTitle(s string) string {
	words := splitKebab(s)
	for i, w := range words {
		if len(w) > 0 && w[0] >= 'a' && w[0] <= 'z' {
			words[i] = string(w[0]-('a'-'A')) + w[1:]
		}
	}
	return joinWords(words)
}

func splitKebab(s string) []string {
	var words []string
	current := ""
	for _, c := range s {
		if c == '-' {
			if current != "" {
				words = append(words, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		words = append(words, current)
	}
	return words
}

func joinWords(words []string) string {
	result := ""
	for i, w := range words {
		if i > 0 {
			result += " "
		}
		result += w
	}
	return result
}
