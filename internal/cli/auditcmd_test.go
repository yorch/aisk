package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

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

func TestPruneByAge(t *testing.T) {
	now := time.Now()
	events := []audit.Event{
		{Action: "old", Timestamp: now.Add(-40 * 24 * time.Hour).Format(time.RFC3339Nano)},
		{Action: "new", Timestamp: now.Add(-5 * 24 * time.Hour).Format(time.RFC3339Nano)},
	}
	got := pruneByAge(events, 30)
	if len(got) != 1 || got[0].Action != "new" {
		t.Fatalf("unexpected prune result: %+v", got)
	}
}

func TestRunAuditPrune_DryRun(t *testing.T) {
	home := t.TempDir()
	logPath := filepath.Join(t.TempDir(), "audit.log")
	t.Setenv("HOME", home)
	t.Setenv("AISK_AUDIT_LOG_PATH", logPath)
	t.Setenv("AISK_AUDIT_MAX_BACKUPS", "2")

	if err := os.WriteFile(logPath, []byte(`{"timestamp":"2026-01-01T00:00:00Z","action":"a"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(logPath+".1", []byte(`{"timestamp":"2026-01-02T00:00:00Z","action":"b"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	origDays, origKeep, origDry := auditPruneKeepDays, auditPruneKeep, auditPruneDryRun
	t.Cleanup(func() {
		auditPruneKeepDays, auditPruneKeep, auditPruneDryRun = origDays, origKeep, origDry
	})
	auditPruneKeepDays = 0
	auditPruneKeep = 1
	auditPruneDryRun = true

	if err := runAuditPrune(nil, nil); err != nil {
		t.Fatalf("runAuditPrune error: %v", err)
	}

	events, err := loadAuditEventsWithBackups(logPath)
	if err != nil {
		t.Fatalf("reload events: %v", err)
	}
	// Dry-run should not mutate files.
	if len(events) != 2 {
		t.Fatalf("expected 2 events after dry-run, got %d", len(events))
	}
}

func TestFilterBySince_Duration(t *testing.T) {
	now := time.Date(2026, 2, 22, 12, 0, 0, 0, time.UTC)
	events := []audit.Event{
		{Action: "old", Timestamp: now.Add(-48 * time.Hour).Format(time.RFC3339Nano)},
		{Action: "new", Timestamp: now.Add(-2 * time.Hour).Format(time.RFC3339Nano)},
	}

	got, err := filterBySince(events, "24h", now)
	if err != nil {
		t.Fatalf("filterBySince error: %v", err)
	}
	if len(got) != 1 || got[0].Action != "new" {
		t.Fatalf("unexpected since filter result: %+v", got)
	}
}

func TestFilterBySince_RFC3339(t *testing.T) {
	events := []audit.Event{
		{Action: "a1", Timestamp: "2026-02-20T10:00:00Z"},
		{Action: "a2", Timestamp: "2026-02-22T10:00:00Z"},
	}
	got, err := filterBySince(events, "2026-02-21T00:00:00Z", time.Now())
	if err != nil {
		t.Fatalf("filterBySince error: %v", err)
	}
	if len(got) != 1 || got[0].Action != "a2" {
		t.Fatalf("unexpected RFC3339 since filter result: %+v", got)
	}
}

func TestComputeAuditStats(t *testing.T) {
	events := []audit.Event{
		{Command: "install", Action: "command.install", Status: "success", ClientID: "claude"},
		{Command: "install", Action: "install.adapter.apply", Status: "success", ClientID: "claude"},
		{Command: "status", Action: "command.status", Status: "error", ClientID: "cursor"},
	}
	stats := computeAuditStats(events)
	if stats.Total != 3 {
		t.Fatalf("expected total=3, got %d", stats.Total)
	}
	if stats.ByCommand["install"] != 2 || stats.ByCommand["status"] != 1 {
		t.Fatalf("unexpected ByCommand: %+v", stats.ByCommand)
	}
	if stats.ByStatus["success"] != 2 || stats.ByStatus["error"] != 1 {
		t.Fatalf("unexpected ByStatus: %+v", stats.ByStatus)
	}
	if stats.ByClientID["claude"] != 2 || stats.ByClientID["cursor"] != 1 {
		t.Fatalf("unexpected ByClientID: %+v", stats.ByClientID)
	}
}
