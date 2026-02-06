package adapter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yorch/aisk/internal/skill"
)

func TestClaudeAdapter_Install_Symlink(t *testing.T) {
	// Create a source skill
	srcDir := t.TempDir()
	os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("# Test"), 0o644)

	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{Name: "test-skill"},
		DirName:     "test-skill",
		Path:        srcDir,
		Source:       skill.SourceLocal,
	}

	targetDir := t.TempDir()
	adapter := &ClaudeAdapter{}

	err := adapter.Install(s, targetDir, InstallOpts{})
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	link := filepath.Join(targetDir, "test-skill")
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("Readlink failed: %v", err)
	}
	if target != srcDir {
		t.Errorf("symlink target = %q, want %q", target, srcDir)
	}
}

func TestClaudeAdapter_Install_Copy(t *testing.T) {
	srcDir := t.TempDir()
	os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("# Remote"), 0o644)
	os.MkdirAll(filepath.Join(srcDir, "reference"), 0o755)
	os.WriteFile(filepath.Join(srcDir, "reference", "guide.md"), []byte("# Guide"), 0o644)

	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{Name: "remote-skill"},
		DirName:     "remote-skill",
		Path:        srcDir,
		Source:       skill.SourceRemote,
	}

	targetDir := t.TempDir()
	adapter := &ClaudeAdapter{}

	err := adapter.Install(s, targetDir, InstallOpts{})
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Verify files were copied
	dest := filepath.Join(targetDir, "remote-skill")
	if _, err := os.Stat(filepath.Join(dest, "SKILL.md")); err != nil {
		t.Error("SKILL.md not copied")
	}
	if _, err := os.Stat(filepath.Join(dest, "reference", "guide.md")); err != nil {
		t.Error("reference/guide.md not copied")
	}
}

func TestClaudeAdapter_Uninstall(t *testing.T) {
	targetDir := t.TempDir()
	dest := filepath.Join(targetDir, "test-skill")
	os.MkdirAll(dest, 0o755)
	os.WriteFile(filepath.Join(dest, "SKILL.md"), []byte("# Test"), 0o644)

	s := &skill.Skill{
		Frontmatter: skill.Frontmatter{Name: "test-skill"},
		DirName:     "test-skill",
	}

	adapter := &ClaudeAdapter{}
	err := adapter.Uninstall(s, targetDir)
	if err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Error("skill directory should be removed after uninstall")
	}
}
