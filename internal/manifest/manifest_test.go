package manifest

import (
	"path/filepath"
	"testing"
	"time"
)

func TestManifest_AddAndFind(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")

	m, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	now := time.Now()
	m.Add(Installation{
		SkillName:    "test-skill",
		SkillVersion: "1.0.0",
		ClientID:     "claude",
		Scope:        "global",
		InstalledAt:  now,
		UpdatedAt:    now,
		InstallPath:  "/path/to/skill",
	})

	found := m.Find("test-skill", "claude")
	if len(found) != 1 {
		t.Fatalf("found %d installations, want 1", len(found))
	}
	if found[0].SkillVersion != "1.0.0" {
		t.Errorf("version = %q, want %q", found[0].SkillVersion, "1.0.0")
	}
}

func TestManifest_AddReplacesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")

	m, _ := Load(path)
	now := time.Now()

	m.Add(Installation{SkillName: "s", ClientID: "claude", Scope: "global", SkillVersion: "1.0", InstalledAt: now, UpdatedAt: now})
	m.Add(Installation{SkillName: "s", ClientID: "claude", Scope: "global", SkillVersion: "2.0", InstalledAt: now, UpdatedAt: now})

	found := m.Find("s", "claude")
	if len(found) != 1 {
		t.Fatalf("found %d, want 1 (should replace)", len(found))
	}
	if found[0].SkillVersion != "2.0" {
		t.Errorf("version = %q, want %q", found[0].SkillVersion, "2.0")
	}
}

func TestManifest_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")

	m, _ := Load(path)
	now := time.Now()
	m.Add(Installation{
		SkillName:    "test",
		SkillVersion: "1.0",
		ClientID:     "claude",
		Scope:        "global",
		InstalledAt:  now,
		UpdatedAt:    now,
		InstallPath:  "/test",
	})

	if err := m.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	m2, err := Load(path)
	if err != nil {
		t.Fatalf("Load after save: %v", err)
	}

	if len(m2.Installations) != 1 {
		t.Fatalf("loaded %d installations, want 1", len(m2.Installations))
	}
	if m2.Installations[0].SkillName != "test" {
		t.Error("loaded skill name mismatch")
	}
}

func TestManifest_Remove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")

	m, _ := Load(path)
	now := time.Now()
	m.Add(Installation{SkillName: "a", ClientID: "claude", Scope: "global", InstalledAt: now, UpdatedAt: now})
	m.Add(Installation{SkillName: "a", ClientID: "gemini", Scope: "global", InstalledAt: now, UpdatedAt: now})
	m.Add(Installation{SkillName: "b", ClientID: "claude", Scope: "global", InstalledAt: now, UpdatedAt: now})

	m.Remove("a", "claude", "global")

	if len(m.Find("a", "")) != 1 {
		t.Error("should have 1 remaining installation for skill 'a'")
	}
	if len(m.Find("b", "")) != 1 {
		t.Error("skill 'b' should be unaffected")
	}
}

func TestManifest_RemoveAll(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")

	m, _ := Load(path)
	now := time.Now()
	m.Add(Installation{SkillName: "a", ClientID: "claude", Scope: "global", InstalledAt: now, UpdatedAt: now})
	m.Add(Installation{SkillName: "a", ClientID: "gemini", Scope: "global", InstalledAt: now, UpdatedAt: now})

	m.RemoveAll("a")

	if len(m.Find("a", "")) != 0 {
		t.Error("all installations of 'a' should be removed")
	}
}

func TestManifest_AllSkillNames(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")

	m, _ := Load(path)
	now := time.Now()
	m.Add(Installation{SkillName: "a", ClientID: "claude", Scope: "global", InstalledAt: now, UpdatedAt: now})
	m.Add(Installation{SkillName: "a", ClientID: "gemini", Scope: "global", InstalledAt: now, UpdatedAt: now})
	m.Add(Installation{SkillName: "b", ClientID: "claude", Scope: "global", InstalledAt: now, UpdatedAt: now})

	names := m.AllSkillNames()
	if len(names) != 2 {
		t.Errorf("got %d unique names, want 2", len(names))
	}
}

func TestManifest_LoadNonExistent(t *testing.T) {
	m, err := Load("/nonexistent/path/manifest.json")
	if err != nil {
		t.Fatalf("should not error on missing file: %v", err)
	}
	if len(m.Installations) != 0 {
		t.Error("should return empty manifest")
	}
}
