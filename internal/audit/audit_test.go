package audit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoggerWritesJSONL(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AISK_AUDIT_ENABLED", "true")
	t.Setenv("AISK_AUDIT_LOG_PATH", filepath.Join(dir, "audit.log"))

	l := New(filepath.Join(dir, ".aisk"), "install")
	l.Log("command.install", "started", map[string]any{"dry_run": true}, nil)
	l.LogEvent(Event{
		Action:   "install.adapter.apply",
		Status:   "success",
		Skill:    "skill-a",
		ClientID: "claude",
	})

	data, err := os.ReadFile(filepath.Join(dir, "audit.log"))
	if err != nil {
		t.Fatalf("reading audit log: %v", err)
	}

	lines := splitNonEmptyLines(string(data))
	if len(lines) != 2 {
		t.Fatalf("expected 2 log lines, got %d", len(lines))
	}

	var e Event
	if err := json.Unmarshal([]byte(lines[0]), &e); err != nil {
		t.Fatalf("invalid json line: %v", err)
	}
	if e.Command != "install" || e.Action != "command.install" || e.RunID == "" {
		t.Fatalf("unexpected first event: %+v", e)
	}
}

func TestLoggerDisabled(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")
	t.Setenv("AISK_AUDIT_ENABLED", "false")
	t.Setenv("AISK_AUDIT_LOG_PATH", logPath)

	l := New(filepath.Join(dir, ".aisk"), "install")
	l.Log("command.install", "started", nil, nil)

	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Fatalf("expected no log file when disabled, got err=%v", err)
	}
}

func TestLoggerRotate(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")
	t.Setenv("AISK_AUDIT_ENABLED", "true")
	t.Setenv("AISK_AUDIT_LOG_PATH", logPath)

	oldMax := maxLogSizeBytes
	maxLogSizeBytes = 32
	t.Cleanup(func() { maxLogSizeBytes = oldMax })

	if err := os.WriteFile(logPath, []byte("012345678901234567890123456789012345"), 0o644); err != nil {
		t.Fatalf("write seed file: %v", err)
	}

	l := New(filepath.Join(dir, ".aisk"), "install")
	l.Log("command.install", "started", nil, nil)

	if _, err := os.Stat(logPath + ".1"); err != nil {
		t.Fatalf("expected rotated file, got: %v", err)
	}
}

func TestLoggerRotate_MultipleBackups(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "audit.log")
	t.Setenv("AISK_AUDIT_ENABLED", "true")
	t.Setenv("AISK_AUDIT_LOG_PATH", logPath)
	t.Setenv("AISK_AUDIT_MAX_BACKUPS", "2")

	oldMax := maxLogSizeBytes
	maxLogSizeBytes = 16
	t.Cleanup(func() { maxLogSizeBytes = oldMax })

	if err := os.WriteFile(logPath, []byte("seed-seed-seed-seed"), 0o644); err != nil {
		t.Fatalf("write seed: %v", err)
	}

	l := New(filepath.Join(dir, ".aisk"), "install")
	l.Log("a", "b", nil, nil) // rotate log -> .1
	if err := os.WriteFile(logPath, []byte("seed-seed-seed-seed"), 0o644); err != nil {
		t.Fatalf("write seed 2: %v", err)
	}
	l.Log("a", "b", nil, nil) // rotate again -> .1, old .1 -> .2

	if _, err := os.Stat(logPath + ".1"); err != nil {
		t.Fatalf("missing .1: %v", err)
	}
	if _, err := os.Stat(logPath + ".2"); err != nil {
		t.Fatalf("missing .2: %v", err)
	}
}

func TestCandidateLogPaths(t *testing.T) {
	dir := t.TempDir()
	primary := filepath.Join(dir, "audit.log")
	t.Setenv("AISK_AUDIT_MAX_BACKUPS", "3")
	_ = os.WriteFile(primary+".2", []byte("x"), 0o644)
	_ = os.WriteFile(primary, []byte("x"), 0o644)
	_ = os.WriteFile(primary+".1", []byte("x"), 0o644)

	paths := CandidateLogPaths(primary)
	if len(paths) != 3 {
		t.Fatalf("expected 3 paths, got %d (%v)", len(paths), paths)
	}
	if paths[0] != primary+".2" || paths[1] != primary+".1" || paths[2] != primary {
		t.Fatalf("unexpected path order: %v", paths)
	}
}

func splitNonEmptyLines(s string) []string {
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
