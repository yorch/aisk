package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yorch/aisk/internal/manifest"
)

func TestInferSectionInstallOperation_Create(t *testing.T) {
	target := filepath.Join(t.TempDir(), "instructions.md")
	if got := inferSectionInstallOperation(target, "skill-a"); got != "create" {
		t.Fatalf("expected create, got %q", got)
	}
}

func TestInferSectionInstallOperation_Replace(t *testing.T) {
	target := filepath.Join(t.TempDir(), "instructions.md")
	content := "<!-- aisk:start:skill-a -->\nbody\n<!-- aisk:end:skill-a -->\n"
	if err := os.WriteFile(target, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := inferSectionInstallOperation(target, "skill-a"); got != "replace" {
		t.Fatalf("expected replace, got %q", got)
	}
}

func TestInferSectionInstallOperation_Append(t *testing.T) {
	target := filepath.Join(t.TempDir(), "instructions.md")
	if err := os.WriteFile(target, []byte("# Existing content\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := inferSectionInstallOperation(target, "skill-a"); got != "append" {
		t.Fatalf("expected append, got %q", got)
	}
}

func TestRunPlanInstall_WithYesRequiresSkillArg(t *testing.T) {
	home := t.TempDir()
	skillsRepo := t.TempDir()
	createTestSkill(t, skillsRepo, "skill-a", "1.0.0")
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", home)
	t.Setenv("AISK_SKILLS_PATH", skillsRepo)

	origAssumeYes, origClient := assumeYes, planInstallClient
	t.Cleanup(func() {
		assumeYes, planInstallClient = origAssumeYes, origClient
	})
	assumeYes = true
	planInstallClient = "claude"

	err := runPlanInstall(nil, nil)
	if err == nil || !strings.Contains(err.Error(), "skill argument is required") {
		t.Fatalf("expected missing skill validation error, got: %v", err)
	}
}

func TestRunPlanInstall_WithYesRequiresClient(t *testing.T) {
	home := t.TempDir()
	skillsRepo := t.TempDir()
	createTestSkill(t, skillsRepo, "skill-a", "1.0.0")

	t.Setenv("HOME", home)
	t.Setenv("AISK_SKILLS_PATH", skillsRepo)

	origAssumeYes, origClient := assumeYes, planInstallClient
	t.Cleanup(func() {
		assumeYes, planInstallClient = origAssumeYes, origClient
	})
	assumeYes = true
	planInstallClient = ""

	err := runPlanInstall(nil, []string{"skill-a"})
	if err == nil || !strings.Contains(err.Error(), "--client is required") {
		t.Fatalf("expected missing client validation error, got: %v", err)
	}
}

func TestRunPlanInstall_SuccessOutput(t *testing.T) {
	home := t.TempDir()
	skillsRepo := t.TempDir()
	createTestSkill(t, skillsRepo, "skill-a", "1.0.0")
	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", home)
	t.Setenv("AISK_SKILLS_PATH", skillsRepo)

	origAssumeYes, origClient, origScope := assumeYes, planInstallClient, planInstallScope
	t.Cleanup(func() {
		assumeYes, planInstallClient, planInstallScope = origAssumeYes, origClient, origScope
	})
	assumeYes = true
	planInstallClient = "claude"
	planInstallScope = "global"

	out := captureStdout(t, func() {
		if err := runPlanInstall(nil, []string{"skill-a"}); err != nil {
			t.Fatalf("runPlanInstall error: %v", err)
		}
	})
	if !strings.Contains(out, `Plan (install): "skill-a"`) {
		t.Fatalf("expected plan install header, got: %s", out)
	}
	if !strings.Contains(out, "adapter:") {
		t.Fatalf("expected adapter description, got: %s", out)
	}
}

func TestRunPlanUpdate_SuccessOutput(t *testing.T) {
	home := t.TempDir()
	skillsRepo := t.TempDir()
	createTestSkill(t, skillsRepo, "skill-a", "2.0.0")

	t.Setenv("HOME", home)
	t.Setenv("AISK_SKILLS_PATH", skillsRepo)

	m, err := manifest.Load(filepath.Join(home, ".aisk", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	m.Add(manifest.Installation{
		SkillName:    "skill-a",
		SkillVersion: "1.0.0",
		ClientID:     "claude",
		Scope:        "global",
		InstalledAt:  time.Now(),
		UpdatedAt:    time.Now(),
		InstallPath:  filepath.Join(home, ".claude", "skills"),
	})
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}

	origClient := planUpdateClient
	t.Cleanup(func() { planUpdateClient = origClient })
	planUpdateClient = ""

	out := captureStdout(t, func() {
		if err := runPlanUpdate(nil, []string{"skill-a"}); err != nil {
			t.Fatalf("runPlanUpdate error: %v", err)
		}
	})
	if !strings.Contains(out, "Plan (update):") || !strings.Contains(out, "1.0.0 -> 2.0.0") {
		t.Fatalf("unexpected plan update output: %s", out)
	}
}

func TestRunPlanUninstall_SuccessOutput(t *testing.T) {
	home := t.TempDir()
	skillsRepo := t.TempDir()
	createTestSkill(t, skillsRepo, "skill-a", "1.0.0")

	t.Setenv("HOME", home)
	t.Setenv("AISK_SKILLS_PATH", skillsRepo)

	m, err := manifest.Load(filepath.Join(home, ".aisk", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	m.Add(manifest.Installation{
		SkillName:    "skill-a",
		SkillVersion: "1.0.0",
		ClientID:     "codex",
		Scope:        "project",
		InstalledAt:  time.Now(),
		UpdatedAt:    time.Now(),
		InstallPath:  "AGENTS.md",
	})
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}

	origClient := planUninstallClient
	t.Cleanup(func() { planUninstallClient = origClient })
	planUninstallClient = ""

	out := captureStdout(t, func() {
		if err := runPlanUninstall(nil, []string{"skill-a"}); err != nil {
			t.Fatalf("runPlanUninstall error: %v", err)
		}
	})
	if !strings.Contains(out, `Plan (uninstall): "skill-a"`) || !strings.Contains(out, "remove managed section from AGENTS.md") {
		t.Fatalf("unexpected plan uninstall output: %s", out)
	}
}

func createTestSkill(t *testing.T, repo, name, version string) {
	t.Helper()
	skillDir := filepath.Join(repo, name)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `---
name: ` + name + `
description: test
version: ` + version + `
---
# Skill
Use when: test
`
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
