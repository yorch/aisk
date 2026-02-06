package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yorch/aisk/internal/skill"
)

func TestWindsurfAdapter_Install_Project(t *testing.T) {
	dir := t.TempDir()

	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{
			Name:        "test-skill",
			Description: "A test",
		},
		DirName:      "test-skill",
		MarkdownBody: "# Test\n\nBody.",
	}

	adapter := &WindsurfAdapter{}
	if err := adapter.Install(s, dir, InstallOpts{Scope: "project"}); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	dest := filepath.Join(dir, "test-skill.md")
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "# test-skill") {
		t.Error("should contain skill name as header")
	}
	if !strings.Contains(content, "Body.") {
		t.Error("should contain skill body")
	}
}

func TestWindsurfAdapter_Install_Global(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "global_rules.md")

	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{
			Name: "test-skill",
		},
		DirName:      "test-skill",
		MarkdownBody: "# Test\n\nGlobal body.",
	}

	adapter := &WindsurfAdapter{}
	if err := adapter.Install(s, target, InstallOpts{Scope: "global"}); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	data, _ := os.ReadFile(target)
	content := string(data)

	if !strings.Contains(content, "<!-- aisk:start:test-skill -->") {
		t.Error("should use section markers for global install")
	}
}
