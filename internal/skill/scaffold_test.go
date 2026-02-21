package skill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffold_CreatesStructure(t *testing.T) {
	dir := t.TempDir()
	skillDir, err := Scaffold(dir, "my-new-skill")
	if err != nil {
		t.Fatalf("Scaffold error: %v", err)
	}

	// Check directory was created
	if _, err := os.Stat(skillDir); err != nil {
		t.Fatalf("skill directory not created: %v", err)
	}

	// Check expected files and dirs
	expected := []string{
		"SKILL.md",
		"README.md",
		"reference",
		"examples",
	}
	for _, name := range expected {
		p := filepath.Join(skillDir, name)
		if _, err := os.Stat(p); err != nil {
			t.Errorf("expected %s to exist: %v", name, err)
		}
	}
}

func TestScaffold_InvalidName(t *testing.T) {
	dir := t.TempDir()
	_, err := Scaffold(dir, "InvalidName")
	if err == nil {
		t.Error("expected error for invalid name")
	}
}

func TestScaffold_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "existing-skill"), 0o755)

	_, err := Scaffold(dir, "existing-skill")
	if err == nil {
		t.Error("expected error for existing directory")
	}
}

func TestScaffold_TemplateContent(t *testing.T) {
	dir := t.TempDir()
	skillDir, err := Scaffold(dir, "test-skill")
	if err != nil {
		t.Fatalf("Scaffold error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("reading SKILL.md: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "name: test-skill") {
		t.Error("SKILL.md should contain the skill name in frontmatter")
	}
	if !strings.Contains(content, "# Test Skill") {
		t.Error("SKILL.md should contain the title-cased name")
	}
	if !strings.Contains(content, "Use when:") {
		t.Error("SKILL.md should contain a 'Use when:' section")
	}
}

func TestScaffold_TemplateContent_DigitWordPrefix(t *testing.T) {
	dir := t.TempDir()
	skillDir, err := Scaffold(dir, "skill-2fa")
	if err != nil {
		t.Fatalf("Scaffold error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		t.Fatalf("reading SKILL.md: %v", err)
	}

	if !strings.Contains(string(data), "# Skill 2fa") {
		t.Fatalf("expected digit-prefixed token to remain valid in title, got:\n%s", string(data))
	}
}
