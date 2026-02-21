package gitignore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureEntries_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	gi := filepath.Join(dir, ".gitignore")

	added, err := EnsureEntries(gi, []string{".claude/skills/"})
	if err != nil {
		t.Fatalf("EnsureEntries error: %v", err)
	}
	if len(added) != 1 {
		t.Errorf("expected 1 added, got %d", len(added))
	}

	data, _ := os.ReadFile(gi)
	content := string(data)
	if !strings.Contains(content, sectionStart) {
		t.Error("missing section start marker")
	}
	if !strings.Contains(content, ".claude/skills/") {
		t.Error("missing entry")
	}
	if !strings.Contains(content, sectionEnd) {
		t.Error("missing section end marker")
	}
}

func TestEnsureEntries_AddsToExisting(t *testing.T) {
	dir := t.TempDir()
	gi := filepath.Join(dir, ".gitignore")

	os.WriteFile(gi, []byte("node_modules/\n"), 0o644)

	added, err := EnsureEntries(gi, []string{".cursor/rules/"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(added) != 1 {
		t.Errorf("expected 1 added, got %d", len(added))
	}

	data, _ := os.ReadFile(gi)
	content := string(data)
	if !strings.Contains(content, "node_modules/") {
		t.Error("existing content should be preserved")
	}
	if !strings.Contains(content, ".cursor/rules/") {
		t.Error("new entry should be present")
	}
}

func TestEnsureEntries_AppendsToSection(t *testing.T) {
	dir := t.TempDir()
	gi := filepath.Join(dir, ".gitignore")

	EnsureEntries(gi, []string{".claude/skills/"})
	added, err := EnsureEntries(gi, []string{".cursor/rules/"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(added) != 1 {
		t.Errorf("expected 1 added, got %d", len(added))
	}

	data, _ := os.ReadFile(gi)
	content := string(data)
	if !strings.Contains(content, ".claude/skills/") {
		t.Error("original entry should still be present")
	}
	if !strings.Contains(content, ".cursor/rules/") {
		t.Error("new entry should be present")
	}

	// Should have only one section start/end
	if strings.Count(content, sectionStart) != 1 {
		t.Error("should have exactly one section start marker")
	}
}

func TestEnsureEntries_Dedup(t *testing.T) {
	dir := t.TempDir()
	gi := filepath.Join(dir, ".gitignore")

	EnsureEntries(gi, []string{".claude/skills/"})
	added, err := EnsureEntries(gi, []string{".claude/skills/"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(added) != 0 {
		t.Errorf("expected 0 added (duplicate), got %d", len(added))
	}
}

func TestRemoveEntries_Basic(t *testing.T) {
	dir := t.TempDir()
	gi := filepath.Join(dir, ".gitignore")

	EnsureEntries(gi, []string{".claude/skills/", ".cursor/rules/"})

	removed, err := RemoveEntries(gi, []string{".claude/skills/"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(removed) != 1 {
		t.Errorf("expected 1 removed, got %d", len(removed))
	}

	data, _ := os.ReadFile(gi)
	content := string(data)
	if strings.Contains(content, ".claude/skills/") {
		t.Error("removed entry should not be present")
	}
	if !strings.Contains(content, ".cursor/rules/") {
		t.Error("remaining entry should still be present")
	}
}

func TestRemoveEntries_CleansEmptySection(t *testing.T) {
	dir := t.TempDir()
	gi := filepath.Join(dir, ".gitignore")

	os.WriteFile(gi, []byte("node_modules/\n"), 0o644)
	EnsureEntries(gi, []string{".claude/skills/"})

	removed, err := RemoveEntries(gi, []string{".claude/skills/"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(removed) != 1 {
		t.Errorf("expected 1 removed, got %d", len(removed))
	}

	data, _ := os.ReadFile(gi)
	content := string(data)
	if strings.Contains(content, sectionStart) {
		t.Error("empty section should be removed")
	}
	if !strings.Contains(content, "node_modules/") {
		t.Error("non-managed content should be preserved")
	}
}

func TestRemoveEntries_NonExistentFile(t *testing.T) {
	removed, err := RemoveEntries("/nonexistent/.gitignore", []string{"foo"})
	if err != nil {
		t.Fatalf("should not error on missing file: %v", err)
	}
	if len(removed) != 0 {
		t.Error("should return 0 removed for nonexistent file")
	}
}

func TestRemoveEntries_NoMatchingEntries(t *testing.T) {
	dir := t.TempDir()
	gi := filepath.Join(dir, ".gitignore")

	EnsureEntries(gi, []string{".claude/skills/"})

	removed, err := RemoveEntries(gi, []string{".cursor/rules/"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(removed) != 0 {
		t.Errorf("expected 0 removed, got %d", len(removed))
	}
}

func TestGitignorePatternsForClient(t *testing.T) {
	tests := []struct {
		clientID string
		want     string
	}{
		{"claude", ".claude/skills/"},
		{"cursor", ".cursor/rules/"},
		{"windsurf", ".windsurf/rules/"},
		{"copilot", ".github/copilot-instructions.md"},
		{"gemini", "GEMINI.md"},
		{"codex", "AGENTS.md"},
	}
	for _, tc := range tests {
		patterns := GitignorePatternsForClient(tc.clientID, "")
		if len(patterns) == 0 || patterns[0] != tc.want {
			t.Errorf("GitignorePatternsForClient(%q) = %v, want [%s]", tc.clientID, patterns, tc.want)
		}
	}
}
