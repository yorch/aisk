package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yorch/aisk/internal/skill"
)

func TestMarkdownAdapter_Install_NewFile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "GEMINI.md")

	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{
			Name:        "test-skill",
			Description: "A test skill",
		},
		DirName:      "test-skill",
		MarkdownBody: "# Test Skill\n\nContent here.",
	}

	adapter := &MarkdownAdapter{ClientName: "Gemini"}
	err := adapter.Install(s, target, InstallOpts{})
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "<!-- aisk:start:test-skill -->") {
		t.Error("missing start marker")
	}
	if !strings.Contains(content, "<!-- aisk:end:test-skill -->") {
		t.Error("missing end marker")
	}
	if !strings.Contains(content, "# test-skill") {
		t.Error("missing skill header")
	}
}

func TestMarkdownAdapter_Install_Idempotent(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "instructions.md")

	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{
			Name:        "test-skill",
			Description: "A test skill",
		},
		DirName:      "test-skill",
		MarkdownBody: "# Test Skill\n\nVersion 1.",
	}

	adapter := &MarkdownAdapter{ClientName: "Codex"}

	// Install twice
	if err := adapter.Install(s, target, InstallOpts{}); err != nil {
		t.Fatalf("first install: %v", err)
	}

	s.MarkdownBody = "# Test Skill\n\nVersion 2."
	if err := adapter.Install(s, target, InstallOpts{}); err != nil {
		t.Fatalf("second install: %v", err)
	}

	data, _ := os.ReadFile(target)
	content := string(data)

	// Should have exactly one start marker
	count := strings.Count(content, "<!-- aisk:start:test-skill -->")
	if count != 1 {
		t.Errorf("found %d start markers, want 1", count)
	}

	// Should have the updated content
	if !strings.Contains(content, "Version 2") {
		t.Error("should contain updated content")
	}
}

func TestMarkdownAdapter_Install_AppendToExisting(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "GEMINI.md")
	os.WriteFile(target, []byte("# Existing Content\n\nSome instructions.\n"), 0o644)

	s := &skill.Skill{
		Frontmatter:  skill.Frontmatter{Name: "new-skill"},
		DirName:      "new-skill",
		MarkdownBody: "# New Skill body",
	}

	adapter := &MarkdownAdapter{ClientName: "Gemini"}
	if err := adapter.Install(s, target, InstallOpts{}); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	data, _ := os.ReadFile(target)
	content := string(data)

	if !strings.Contains(content, "# Existing Content") {
		t.Error("existing content should be preserved")
	}
	if !strings.Contains(content, "<!-- aisk:start:new-skill -->") {
		t.Error("new skill section should be appended")
	}
}

func TestMarkdownAdapter_Uninstall(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "test.md")

	content := `# Existing

<!-- aisk:start:test-skill -->
# test-skill
Content
<!-- aisk:end:test-skill -->

# More Existing`

	os.WriteFile(target, []byte(content), 0o644)

	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{Name: "test-skill"},
		DirName:     "test-skill",
	}

	adapter := &MarkdownAdapter{ClientName: "Gemini"}
	if err := adapter.Uninstall(s, target); err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	data, _ := os.ReadFile(target)
	result := string(data)

	if strings.Contains(result, "aisk:start") {
		t.Error("section markers should be removed")
	}
	if !strings.Contains(result, "# Existing") {
		t.Error("surrounding content should be preserved")
	}
	if !strings.Contains(result, "# More Existing") {
		t.Error("surrounding content should be preserved")
	}
}
