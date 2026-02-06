package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanLocal_DiscoversSkills(t *testing.T) {
	dir := t.TempDir()

	// Create a valid skill
	skillDir := filepath.Join(dir, "my-skill")
	os.MkdirAll(skillDir, 0o755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(`---
name: my-skill
description: A test skill
version: 1.0.0
---

# My Skill
`), 0o644)

	// Create reference files (singular)
	refDir := filepath.Join(skillDir, "reference")
	os.MkdirAll(refDir, 0o755)
	os.WriteFile(filepath.Join(refDir, "guide.md"), []byte("# Guide"), 0o644)

	// Create a non-skill directory (no SKILL.md)
	os.MkdirAll(filepath.Join(dir, "not-a-skill"), 0o755)
	os.WriteFile(filepath.Join(dir, "not-a-skill", "README.md"), []byte("# Readme"), 0o644)

	skills, err := ScanLocal(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(skills) != 1 {
		t.Fatalf("found %d skills, want 1", len(skills))
	}

	s := skills[0]
	if s.DirName != "my-skill" {
		t.Errorf("DirName = %q, want %q", s.DirName, "my-skill")
	}
	if s.Frontmatter.Name != "my-skill" {
		t.Errorf("Name = %q, want %q", s.Frontmatter.Name, "my-skill")
	}
	if s.Source != SourceLocal {
		t.Errorf("Source = %v, want SourceLocal", s.Source)
	}
	if len(s.ReferenceFiles) != 1 {
		t.Errorf("found %d reference files, want 1", len(s.ReferenceFiles))
	}
}

func TestScanLocal_ReferencesPlural(t *testing.T) {
	dir := t.TempDir()

	skillDir := filepath.Join(dir, "test-skill")
	os.MkdirAll(filepath.Join(skillDir, "references"), 0o755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(`---
name: test-skill
description: test
---

# Test
`), 0o644)
	os.WriteFile(filepath.Join(skillDir, "references", "ref.md"), []byte("# Ref"), 0o644)

	skills, err := ScanLocal(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(skills) != 1 {
		t.Fatalf("found %d skills, want 1", len(skills))
	}

	if len(skills[0].ReferenceFiles) != 1 {
		t.Errorf("found %d reference files, want 1", len(skills[0].ReferenceFiles))
	}
}

func TestScanLocal_SkipsHiddenDirs(t *testing.T) {
	dir := t.TempDir()

	// Hidden dir with SKILL.md should be skipped
	hidden := filepath.Join(dir, ".hidden-skill")
	os.MkdirAll(hidden, 0o755)
	os.WriteFile(filepath.Join(hidden, "SKILL.md"), []byte(`---
name: hidden
description: test
---
`), 0o644)

	skills, err := ScanLocal(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(skills) != 0 {
		t.Errorf("found %d skills, want 0 (hidden dirs should be skipped)", len(skills))
	}
}
