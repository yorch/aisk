package skill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateName_Valid(t *testing.T) {
	valid := []string{"my-skill", "a", "abc-def-123", "skill1", "a-b-c"}
	for _, name := range valid {
		if err := ValidateName(name); err != nil {
			t.Errorf("ValidateName(%q) = %v, want nil", name, err)
		}
	}
}

func TestValidateName_Invalid(t *testing.T) {
	invalid := []string{
		"",                    // empty
		"MySkill",             // uppercase
		"-leading",            // leading hyphen
		"trailing-",           // trailing hyphen
		"double--hyphen",      // consecutive hyphens
		"has space",           // space
		"has_underscore",      // underscore
		"ALLCAPS",             // all caps
		strings.Repeat("a", 65), // too long
	}
	for _, name := range invalid {
		if err := ValidateName(name); err == nil {
			t.Errorf("ValidateName(%q) = nil, want error", name)
		}
	}
}

func TestLintSkillMD_Valid(t *testing.T) {
	content := `---
name: my-skill
description: A test skill
version: 1.0.0
---
# My Skill

Use when: you need to test things.
`
	report := LintSkillMD(content)
	if report.HasErrors() {
		t.Errorf("expected no errors, got: %+v", report.Errors())
	}
	if len(report.Warnings()) != 0 {
		t.Errorf("expected no warnings, got: %+v", report.Warnings())
	}
}

func TestLintSkillMD_MissingFrontmatter(t *testing.T) {
	report := LintSkillMD("no frontmatter here")
	if !report.HasErrors() {
		t.Error("expected error for missing frontmatter")
	}
}

func TestLintSkillMD_MissingName(t *testing.T) {
	content := `---
description: A test skill
version: 1.0.0
---
Some body.

Use when: testing.
`
	report := LintSkillMD(content)
	if !report.HasErrors() {
		t.Error("expected error for missing name")
	}
}

func TestLintSkillMD_MissingDescription(t *testing.T) {
	content := `---
name: my-skill
version: 1.0.0
---
Some body.

Use when: testing.
`
	report := LintSkillMD(content)
	if !report.HasErrors() {
		t.Error("expected error for missing description")
	}
}

func TestLintSkillMD_InvalidVersion(t *testing.T) {
	content := `---
name: my-skill
description: A test skill
version: beta
---
Some body.

Use when: testing.
`
	report := LintSkillMD(content)
	if report.HasErrors() {
		t.Error("invalid version should be a warning, not an error")
	}
	if len(report.Warnings()) == 0 {
		t.Error("expected warning for non-semver version")
	}
}

func TestLintSkillMD_EmptyBody(t *testing.T) {
	content := `---
name: my-skill
description: A test skill
version: 1.0.0
---
`
	report := LintSkillMD(content)
	if !report.HasErrors() {
		t.Error("expected error for empty body")
	}
}

func TestLintSkillMD_NoTrigger(t *testing.T) {
	content := `---
name: my-skill
description: A test skill
version: 1.0.0
---
# My Skill

This skill does something.
`
	report := LintSkillMD(content)
	warns := report.Warnings()
	found := false
	for _, w := range warns {
		if strings.Contains(w.Message, "Use when:") {
			found = true
		}
	}
	if !found {
		t.Error("expected warning about missing 'Use when:' trigger")
	}
}

func TestLintSkillDir_Valid(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "my-skill")
	os.MkdirAll(filepath.Join(skillDir, "reference"), 0o755)
	os.MkdirAll(filepath.Join(skillDir, "examples"), 0o755)

	// Add a reference file so dirs aren't empty
	os.WriteFile(filepath.Join(skillDir, "reference", "guide.md"), []byte("ref"), 0o644)
	os.WriteFile(filepath.Join(skillDir, "examples", "ex1.md"), []byte("ex"), 0o644)

	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(`---
name: my-skill
description: A test skill
version: 1.0.0
---
# My Skill

Use when: testing.
`), 0o644)

	report, err := LintSkillDir(skillDir)
	if err != nil {
		t.Fatalf("LintSkillDir error: %v", err)
	}
	if report.HasErrors() {
		t.Errorf("expected no errors, got: %+v", report.Errors())
	}
}

func TestLintSkillDir_NoSkillMD(t *testing.T) {
	dir := t.TempDir()
	report, err := LintSkillDir(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !report.HasErrors() {
		t.Error("expected error for missing SKILL.md")
	}
}

func TestLintSkillDir_EmptyRefDir(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "my-skill")
	os.MkdirAll(filepath.Join(skillDir, "reference"), 0o755)
	os.MkdirAll(filepath.Join(skillDir, "examples"), 0o755)

	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(`---
name: my-skill
description: A test skill
version: 1.0.0
---
# My Skill

Use when: testing.
`), 0o644)

	report, err := LintSkillDir(skillDir)
	if err != nil {
		t.Fatalf("LintSkillDir error: %v", err)
	}
	warns := report.Warnings()
	if len(warns) < 2 {
		t.Errorf("expected warnings for empty reference/ and examples/, got %d warnings", len(warns))
	}
}

func TestLintSkillDir_NotADirectory(t *testing.T) {
	f := filepath.Join(t.TempDir(), "file.txt")
	os.WriteFile(f, []byte("hi"), 0o644)
	_, err := LintSkillDir(f)
	if err == nil {
		t.Error("expected error for non-directory path")
	}
}
