package skill

import (
	"testing"

	"github.com/yorch/aisk/internal/manifest"
)

func TestCheckUpdates_NoUpdates(t *testing.T) {
	installations := []manifest.Installation{
		{SkillName: "skill-a", SkillVersion: "1.0.0", ClientID: "claude"},
	}
	available := []*Skill{
		{Frontmatter: Frontmatter{Name: "skill-a", Version: "1.0.0"}, DirName: "skill-a"},
	}
	updates := CheckUpdates(installations, available)
	if len(updates) != 0 {
		t.Errorf("expected 0 updates, got %d", len(updates))
	}
}

func TestCheckUpdates_SingleUpdate(t *testing.T) {
	installations := []manifest.Installation{
		{SkillName: "skill-a", SkillVersion: "1.0.0", ClientID: "claude"},
	}
	available := []*Skill{
		{Frontmatter: Frontmatter{Name: "skill-a", Version: "2.0.0"}, DirName: "skill-a"},
	}
	updates := CheckUpdates(installations, available)
	if len(updates) != 1 {
		t.Fatalf("expected 1 update, got %d", len(updates))
	}
	if updates[0].InstalledVersion != "1.0.0" || updates[0].AvailableVersion != "2.0.0" {
		t.Errorf("version mismatch: got %s -> %s", updates[0].InstalledVersion, updates[0].AvailableVersion)
	}
}

func TestCheckUpdates_MultiClient(t *testing.T) {
	installations := []manifest.Installation{
		{SkillName: "skill-a", SkillVersion: "1.0.0", ClientID: "claude"},
		{SkillName: "skill-a", SkillVersion: "1.0.0", ClientID: "cursor"},
	}
	available := []*Skill{
		{Frontmatter: Frontmatter{Name: "skill-a", Version: "1.1.0"}, DirName: "skill-a"},
	}
	updates := CheckUpdates(installations, available)
	if len(updates) != 1 {
		t.Fatalf("expected 1 update (grouped), got %d", len(updates))
	}
	if len(updates[0].AffectedClients) != 2 {
		t.Errorf("expected 2 affected clients, got %d", len(updates[0].AffectedClients))
	}
}

func TestCheckUpdates_NotInRepo(t *testing.T) {
	installations := []manifest.Installation{
		{SkillName: "orphan-skill", SkillVersion: "1.0.0", ClientID: "claude"},
	}
	available := []*Skill{
		{Frontmatter: Frontmatter{Name: "other-skill", Version: "1.0.0"}, DirName: "other-skill"},
	}
	updates := CheckUpdates(installations, available)
	if len(updates) != 0 {
		t.Errorf("expected 0 updates for skill not in repo, got %d", len(updates))
	}
}

func TestCheckUpdates_Unversioned(t *testing.T) {
	installations := []manifest.Installation{
		{SkillName: "skill-a", SkillVersion: "unversioned", ClientID: "claude"},
	}
	available := []*Skill{
		{Frontmatter: Frontmatter{Name: "skill-a", Version: "1.0.0"}, DirName: "skill-a"},
	}
	updates := CheckUpdates(installations, available)
	if len(updates) != 1 {
		t.Fatalf("expected 1 update for unversioned skill, got %d", len(updates))
	}
}

func TestCheckUpdates_Empty(t *testing.T) {
	updates := CheckUpdates(nil, nil)
	if len(updates) != 0 {
		t.Errorf("expected 0 updates for empty input, got %d", len(updates))
	}
}

func TestCheckUpdates_MixedInstalledVersions(t *testing.T) {
	installations := []manifest.Installation{
		{SkillName: "skill-a", SkillVersion: "2.0.0", ClientID: "claude"},
		{SkillName: "skill-a", SkillVersion: "1.0.0", ClientID: "cursor"},
	}
	available := []*Skill{
		{Frontmatter: Frontmatter{Name: "skill-a", Version: "2.0.0"}, DirName: "skill-a"},
	}

	updates := CheckUpdates(installations, available)
	if len(updates) != 1 {
		t.Fatalf("expected 1 update group, got %d", len(updates))
	}
	if updates[0].InstalledVersion != "1.0.0" || updates[0].AvailableVersion != "2.0.0" {
		t.Fatalf("unexpected update versions: %s -> %s", updates[0].InstalledVersion, updates[0].AvailableVersion)
	}
	if len(updates[0].AffectedClients) != 1 || updates[0].AffectedClients[0] != "cursor" {
		t.Fatalf("expected only cursor to be affected, got %v", updates[0].AffectedClients)
	}
}
