package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yorch/aisk/internal/skill"
)

func TestCursorAdapter_Install(t *testing.T) {
	dir := t.TempDir()

	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{
			Name:        "test-skill",
			Description: "A test skill for testing",
		},
		DirName:      "test-skill",
		MarkdownBody: "# Test\n\nBody content.",
	}

	adapter := &CursorAdapter{}
	if err := adapter.Install(s, dir, InstallOpts{}); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	dest := filepath.Join(dir, "test-skill.mdc")
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	content := string(data)
	if !strings.HasPrefix(content, "---\n") {
		t.Error("should start with YAML frontmatter")
	}
	if !strings.Contains(content, "description: A test skill for testing") {
		t.Error("should contain description in frontmatter")
	}
	if !strings.Contains(content, "alwaysApply: false") {
		t.Error("should contain alwaysApply: false")
	}
	if !strings.Contains(content, "Body content.") {
		t.Error("should contain skill body")
	}
}

func TestCursorAdapter_Uninstall(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "test-skill.mdc")
	os.WriteFile(dest, []byte("content"), 0o644)

	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{Name: "test-skill"},
		DirName:     "test-skill",
	}

	adapter := &CursorAdapter{}
	if err := adapter.Uninstall(s, dir); err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Error("file should be removed")
	}
}

func TestCursorAdapter_Uninstall_NotExists(t *testing.T) {
	dir := t.TempDir()

	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{Name: "nonexistent"},
		DirName:     "nonexistent",
	}

	adapter := &CursorAdapter{}
	if err := adapter.Uninstall(s, dir); err != nil {
		t.Fatalf("Uninstall of non-existent should not fail: %v", err)
	}
}
