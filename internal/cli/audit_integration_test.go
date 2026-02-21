package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yorch/aisk/internal/manifest"
)

type auditLine struct {
	Command string         `json:"command"`
	Action  string         `json:"action"`
	Status  string         `json:"status"`
	Skill   string         `json:"skill"`
	Client  string         `json:"client_id"`
	Details map[string]any `json:"details"`
}

func TestInstallWritesAuditEvents_DryRun(t *testing.T) {
	home := t.TempDir()
	skillsRepo := t.TempDir()
	logPath := filepath.Join(t.TempDir(), "audit.log")

	t.Setenv("HOME", home)
	t.Setenv("AISK_SKILLS_PATH", skillsRepo)
	t.Setenv("AISK_AUDIT_ENABLED", "true")
	t.Setenv("AISK_AUDIT_LOG_PATH", logPath)

	if err := os.MkdirAll(filepath.Join(home, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	skillDir := filepath.Join(skillsRepo, "skill-a")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(`---
name: skill-a
description: test
version: 1.0.0
---
# Skill A
Use when: test
`), 0o644); err != nil {
		t.Fatal(err)
	}

	origClient, origScope, origRefs, origDryRun := installClient, installScope, installIncludeRefs, installDryRun
	t.Cleanup(func() {
		installClient, installScope, installIncludeRefs, installDryRun = origClient, origScope, origRefs, origDryRun
	})
	installClient = "claude"
	installScope = "global"
	installIncludeRefs = false
	installDryRun = true

	if err := runInstall(nil, []string{"skill-a"}); err != nil {
		t.Fatalf("runInstall error: %v", err)
	}

	events := readAuditLines(t, logPath)
	if !hasEvent(events, "install", "command.install", "started") {
		t.Fatalf("missing install command started event: %+v", events)
	}
	if !hasEvent(events, "install", "command.install", "success") {
		t.Fatalf("missing install command success event: %+v", events)
	}
	if !hasEvent(events, "install", "install.adapter.apply", "skipped") {
		t.Fatalf("missing install adapter dry-run event: %+v", events)
	}
}

func TestUninstallWritesAuditEvents_Success(t *testing.T) {
	home := t.TempDir()
	skillsRepo := t.TempDir()
	logPath := filepath.Join(t.TempDir(), "audit.log")

	t.Setenv("HOME", home)
	t.Setenv("AISK_SKILLS_PATH", skillsRepo)
	t.Setenv("AISK_AUDIT_ENABLED", "true")
	t.Setenv("AISK_AUDIT_LOG_PATH", logPath)

	skillDir := filepath.Join(skillsRepo, "skill-a")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(`---
name: skill-a
description: test
version: 1.0.0
---
# Skill A
Use when: test
`), 0o644); err != nil {
		t.Fatal(err)
	}

	targetRoot := t.TempDir()
	installedDir := filepath.Join(targetRoot, "skill-a")
	if err := os.MkdirAll(installedDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(installedDir, "SKILL.md"), []byte("# Skill A"), 0o644); err != nil {
		t.Fatal(err)
	}

	manifestPath := filepath.Join(home, ".aisk", "manifest.json")
	m, err := manifest.Load(manifestPath)
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
		InstallPath:  targetRoot,
	})
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}

	origUninstallClient := uninstallClient
	t.Cleanup(func() { uninstallClient = origUninstallClient })
	uninstallClient = ""

	if err := runUninstall(nil, []string{"skill-a"}); err != nil {
		t.Fatalf("runUninstall error: %v", err)
	}

	if _, err := os.Stat(installedDir); !os.IsNotExist(err) {
		t.Fatalf("expected uninstall to remove %s", installedDir)
	}

	events := readAuditLines(t, logPath)
	if !hasEvent(events, "uninstall", "command.uninstall", "started") {
		t.Fatalf("missing uninstall command started event: %+v", events)
	}
	if !hasEvent(events, "uninstall", "uninstall.adapter.apply", "success") {
		t.Fatalf("missing uninstall adapter success event: %+v", events)
	}
	if !hasEvent(events, "uninstall", "command.uninstall", "success") {
		t.Fatalf("missing uninstall command success event: %+v", events)
	}
}

func TestCreateWritesAuditEvents_Success(t *testing.T) {
	home := t.TempDir()
	skillsRepo := t.TempDir()
	logPath := filepath.Join(t.TempDir(), "audit.log")

	t.Setenv("HOME", home)
	t.Setenv("AISK_SKILLS_PATH", skillsRepo)
	t.Setenv("AISK_AUDIT_ENABLED", "true")
	t.Setenv("AISK_AUDIT_LOG_PATH", logPath)

	origCreatePath := createPath
	t.Cleanup(func() { createPath = origCreatePath })
	createPath = ""

	if err := runCreate(nil, []string{"my-new-skill"}); err != nil {
		t.Fatalf("runCreate error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(skillsRepo, "my-new-skill", "SKILL.md")); err != nil {
		t.Fatalf("expected scaffold files to exist: %v", err)
	}

	events := readAuditLines(t, logPath)
	if !hasEvent(events, "create", "command.create", "started") {
		t.Fatalf("missing create command start event: %+v", events)
	}
	if !hasEvent(events, "create", "create.scaffold", "success") {
		t.Fatalf("missing create scaffold success event: %+v", events)
	}
	if !hasEvent(events, "create", "command.create", "success") {
		t.Fatalf("missing create command success event: %+v", events)
	}
}

func TestLintWritesAuditEvents_Success(t *testing.T) {
	home := t.TempDir()
	logPath := filepath.Join(t.TempDir(), "audit.log")
	target := filepath.Join(t.TempDir(), "SKILL.md")

	t.Setenv("HOME", home)
	t.Setenv("AISK_AUDIT_ENABLED", "true")
	t.Setenv("AISK_AUDIT_LOG_PATH", logPath)

	if err := os.WriteFile(target, []byte(`---
name: skill-a
description: test
version: 1.0.0
---
# Skill A
Use when: test
`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := runLint(nil, []string{target}); err != nil {
		t.Fatalf("runLint error: %v", err)
	}

	events := readAuditLines(t, logPath)
	if !hasEvent(events, "lint", "command.lint", "started") {
		t.Fatalf("missing lint command start event: %+v", events)
	}
	if !hasEvent(events, "lint", "command.lint", "success") {
		t.Fatalf("missing lint command success event: %+v", events)
	}
}

func TestUpdateWritesAuditEvents_Success(t *testing.T) {
	home := t.TempDir()
	skillsRepo := t.TempDir()
	logPath := filepath.Join(t.TempDir(), "audit.log")

	t.Setenv("HOME", home)
	t.Setenv("AISK_SKILLS_PATH", skillsRepo)
	t.Setenv("AISK_AUDIT_ENABLED", "true")
	t.Setenv("AISK_AUDIT_LOG_PATH", logPath)

	skillDir := filepath.Join(skillsRepo, "skill-a")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(`---
name: skill-a
description: test
version: 2.0.0
---
# Skill A
Use when: test
`), 0o644); err != nil {
		t.Fatal(err)
	}

	targetRoot := t.TempDir()
	manifestPath := filepath.Join(home, ".aisk", "manifest.json")
	m, err := manifest.Load(manifestPath)
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
		InstallPath:  targetRoot,
	})
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}

	origUpdateClient := updateClient
	t.Cleanup(func() { updateClient = origUpdateClient })
	updateClient = ""

	if err := runUpdate(nil, []string{"skill-a"}); err != nil {
		t.Fatalf("runUpdate error: %v", err)
	}

	events := readAuditLines(t, logPath)
	if !hasEvent(events, "update", "command.update", "started") {
		t.Fatalf("missing update command start event: %+v", events)
	}
	if !hasEvent(events, "update", "update.adapter.apply", "success") {
		t.Fatalf("missing update adapter success event: %+v", events)
	}
	if !hasEvent(events, "update", "command.update", "success") {
		t.Fatalf("missing update command success event: %+v", events)
	}
}

func TestStatusWritesAuditEvents_UpdateCheck(t *testing.T) {
	home := t.TempDir()
	skillsRepo := t.TempDir()
	logPath := filepath.Join(t.TempDir(), "audit.log")

	t.Setenv("HOME", home)
	t.Setenv("AISK_SKILLS_PATH", skillsRepo)
	t.Setenv("AISK_AUDIT_ENABLED", "true")
	t.Setenv("AISK_AUDIT_LOG_PATH", logPath)

	skillDir := filepath.Join(skillsRepo, "skill-a")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(`---
name: skill-a
description: test
version: 1.0.0
---
# Skill A
Use when: test
`), 0o644); err != nil {
		t.Fatal(err)
	}

	manifestPath := filepath.Join(home, ".aisk", "manifest.json")
	m, err := manifest.Load(manifestPath)
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
		InstallPath:  t.TempDir(),
	})
	if err := m.Save(); err != nil {
		t.Fatal(err)
	}

	origStatusJSON, origStatusCheck := statusJSON, statusCheckUpdates
	t.Cleanup(func() {
		statusJSON, statusCheckUpdates = origStatusJSON, origStatusCheck
	})
	statusJSON = false
	statusCheckUpdates = true

	if err := runStatus(nil, nil); err != nil {
		t.Fatalf("runStatus error: %v", err)
	}

	events := readAuditLines(t, logPath)
	if !hasEvent(events, "status", "status.updates.check", "success") {
		t.Fatalf("missing status updates check success event: %+v", events)
	}
	if !hasEvent(events, "status", "command.status", "success") {
		t.Fatalf("missing status command success event: %+v", events)
	}
}

func TestListWritesAuditEvents_RemoteSkipped(t *testing.T) {
	home := t.TempDir()
	skillsRepo := t.TempDir()
	logPath := filepath.Join(t.TempDir(), "audit.log")

	t.Setenv("HOME", home)
	t.Setenv("AISK_SKILLS_PATH", skillsRepo)
	t.Setenv("AISK_AUDIT_ENABLED", "true")
	t.Setenv("AISK_AUDIT_LOG_PATH", logPath)
	t.Setenv("AISK_REMOTE_REPO", "")

	origListRemote, origListRepo, origListJSON := listRemote, listRepo, listJSON
	t.Cleanup(func() {
		listRemote, listRepo, listJSON = origListRemote, origListRepo, origListJSON
	})
	listRemote = true
	listRepo = ""
	listJSON = false

	if err := runList(nil, nil); err != nil {
		t.Fatalf("runList error: %v", err)
	}

	events := readAuditLines(t, logPath)
	if !hasEvent(events, "list", "list.remote.fetch", "skipped") {
		t.Fatalf("missing list remote skipped event: %+v", events)
	}
	if !hasEvent(events, "list", "command.list", "success") {
		t.Fatalf("missing list command success event: %+v", events)
	}
}

func readAuditLines(t *testing.T, path string) []auditLine {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read audit log: %v", err)
	}
	rawLines := splitLines(string(data))
	lines := make([]auditLine, 0, len(rawLines))
	for _, raw := range rawLines {
		var e auditLine
		if err := json.Unmarshal([]byte(raw), &e); err != nil {
			t.Fatalf("invalid audit json line: %v\nline=%s", err, raw)
		}
		lines = append(lines, e)
	}
	return lines
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == '\n' {
			if i > start {
				out = append(out, s[start:i])
			}
			start = i + 1
		}
	}
	return out
}

func hasEvent(events []auditLine, command, action, status string) bool {
	for _, e := range events {
		if e.Command == command && e.Action == action && e.Status == status {
			return true
		}
	}
	return false
}
