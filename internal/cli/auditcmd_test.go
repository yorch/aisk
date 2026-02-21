package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yorch/aisk/internal/audit"
)

func TestTailAuditEvents(t *testing.T) {
	events := []audit.Event{
		{Action: "a1"},
		{Action: "a2"},
		{Action: "a3"},
	}
	got := tailAuditEvents(events, 2)
	if len(got) != 2 || got[0].Action != "a2" || got[1].Action != "a3" {
		t.Fatalf("unexpected tail result: %+v", got)
	}
}

func TestFilterAuditEvents(t *testing.T) {
	events := []audit.Event{
		{RunID: "r1", Action: "x", Status: "success"},
		{RunID: "r2", Action: "y", Status: "error"},
	}
	got := filterAuditEvents(events, "r2", "y", "error")
	if len(got) != 1 || got[0].RunID != "r2" {
		t.Fatalf("unexpected filtered result: %+v", got)
	}
}

func TestLoadAuditEvents(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")
	content := `{"timestamp":"2026-01-01T00:00:00Z","run_id":"r1","command":"install","action":"command.install","status":"started"}
invalid-json
{"timestamp":"2026-01-01T00:00:01Z","run_id":"r1","command":"install","action":"command.install","status":"success"}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write audit log: %v", err)
	}

	events, err := loadAuditEvents(path)
	if err != nil {
		t.Fatalf("loadAuditEvents error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 parsed events, got %d", len(events))
	}
}

func TestLoadAuditEventsWithBackups(t *testing.T) {
	dir := t.TempDir()
	primary := filepath.Join(dir, "audit.log")
	t.Setenv("AISK_AUDIT_MAX_BACKUPS", "2")
	if err := os.WriteFile(primary+".2", []byte(`{"action":"oldest"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(primary+".1", []byte(`{"action":"older"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(primary, []byte(`{"action":"newest"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	events, err := loadAuditEventsWithBackups(primary)
	if err != nil {
		t.Fatalf("loadAuditEventsWithBackups error: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	if events[0].Action != "oldest" || events[1].Action != "older" || events[2].Action != "newest" {
		t.Fatalf("unexpected event order: %+v", events)
	}
}
