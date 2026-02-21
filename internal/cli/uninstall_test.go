package cli

import (
	"path/filepath"
	"testing"

	"github.com/yorch/aisk/internal/manifest"
)

func TestIsInstallationInProject_AbsolutePathInProject(t *testing.T) {
	root := t.TempDir()
	inst := manifest.Installation{
		Scope:       "project",
		InstallPath: filepath.Join(root, ".claude", "skills"),
	}

	if !isInstallationInProject(inst, root) {
		t.Fatal("expected installation to be treated as in-project")
	}
}

func TestIsInstallationInProject_AbsolutePathOutsideProject(t *testing.T) {
	root := t.TempDir()
	other := t.TempDir()
	inst := manifest.Installation{
		Scope:       "project",
		InstallPath: filepath.Join(other, ".claude", "skills"),
	}

	if isInstallationInProject(inst, root) {
		t.Fatal("expected installation outside project root to be excluded")
	}
}

func TestIsInstallationInProject_LegacyRelativePath(t *testing.T) {
	root := t.TempDir()
	inst := manifest.Installation{
		Scope:       "project",
		InstallPath: ".claude/skills",
	}

	if !isInstallationInProject(inst, root) {
		t.Fatal("expected legacy relative install path to be treated as in-project")
	}
}
