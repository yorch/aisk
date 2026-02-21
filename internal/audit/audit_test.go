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
