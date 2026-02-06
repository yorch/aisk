package skill

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// SkillSource indicates where a skill was discovered.
type SkillSource int

const (
	SourceLocal SkillSource = iota
	SourceRemote
)

func (s SkillSource) String() string {
	switch s {
	case SourceLocal:
		return "local"
	case SourceRemote:
		return "remote"
	default:
		return "unknown"
	}
}

// Frontmatter holds the YAML metadata from SKILL.md.
type Frontmatter struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	Version      string   `yaml:"version"`
	AllowedTools []string `yaml:"allowed-tools"`
}

// Skill represents a discovered skill with its metadata and content.
type Skill struct {
	Frontmatter
	DirName        string      // directory name, e.g. "5-whys-skill"
	Path           string      // absolute path to skill directory
	Source         SkillSource // Local or Remote
	MarkdownBody   string      // SKILL.md content after frontmatter
	ReferenceFiles []string    // relative paths under reference/ or references/
	ExampleFiles   []string    // relative paths under examples/
	AssetFiles     []string    // relative paths under assets/
}

// DisplayVersion returns the version string, or "unversioned" if empty.
func (s *Skill) DisplayVersion() string {
	if s.Version == "" {
		return "unversioned"
	}
	return s.Version
}

// ParseFrontmatter splits a SKILL.md file into YAML frontmatter and markdown body.
func ParseFrontmatter(content string) (Frontmatter, string, error) {
	// Normalize line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")

	if !strings.HasPrefix(content, "---") {
		return Frontmatter{}, content, fmt.Errorf("missing frontmatter delimiter")
	}

	// Find second ---
	rest := content[3:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return Frontmatter{}, content, fmt.Errorf("missing closing frontmatter delimiter")
	}

	yamlContent := rest[:idx]
	body := strings.TrimLeft(rest[idx+4:], "\n")

	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return Frontmatter{}, body, fmt.Errorf("parsing frontmatter YAML: %w", err)
	}

	// Clean up multi-line description
	fm.Description = strings.TrimSpace(fm.Description)

	return fm, body, nil
}
