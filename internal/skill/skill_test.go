package skill

import (
	"testing"
)

func TestParseFrontmatter_Valid(t *testing.T) {
	content := `---
name: test-skill
description: A test skill
version: 1.0.0
allowed-tools:
  - Read
  - Grep
---

# Test Skill

This is the body.
`
	fm, body, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Name != "test-skill" {
		t.Errorf("name = %q, want %q", fm.Name, "test-skill")
	}
	if fm.Description != "A test skill" {
		t.Errorf("description = %q, want %q", fm.Description, "A test skill")
	}
	if fm.Version != "1.0.0" {
		t.Errorf("version = %q, want %q", fm.Version, "1.0.0")
	}
	if len(fm.AllowedTools) != 2 {
		t.Errorf("allowed-tools length = %d, want 2", len(fm.AllowedTools))
	}
	if body == "" {
		t.Error("body should not be empty")
	}
	if body[:12] != "# Test Skill" {
		t.Errorf("body starts with %q, want %q", body[:12], "# Test Skill")
	}
}

func TestParseFrontmatter_MissingVersion(t *testing.T) {
	content := `---
name: test-skill
description: No version here
---

# Body
`
	fm, _, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Version != "" {
		t.Errorf("version = %q, want empty", fm.Version)
	}
}

func TestParseFrontmatter_MultiLineDescription(t *testing.T) {
	content := `---
name: test-skill
description: |
  Line one.
  Line two.
  Use when: doing things.
---

# Body
`
	fm, _, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fm.Description == "" {
		t.Error("description should not be empty")
	}
}

func TestParseFrontmatter_MissingDelimiter(t *testing.T) {
	content := `# No frontmatter here`
	_, _, err := ParseFrontmatter(content)
	if err == nil {
		t.Error("expected error for missing frontmatter")
	}
}

func TestParseFrontmatter_MissingClosingDelimiter(t *testing.T) {
	content := `---
name: broken
`
	_, _, err := ParseFrontmatter(content)
	if err == nil {
		t.Error("expected error for missing closing delimiter")
	}
}

func TestDisplayVersion(t *testing.T) {
	s := &Skill{}
	if s.DisplayVersion() != "unversioned" {
		t.Errorf("DisplayVersion() = %q, want %q", s.DisplayVersion(), "unversioned")
	}

	s.Version = "1.0.0"
	if s.DisplayVersion() != "1.0.0" {
		t.Errorf("DisplayVersion() = %q, want %q", s.DisplayVersion(), "1.0.0")
	}
}
