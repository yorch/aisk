package audit

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	defaultLogFilename = "audit.log"
)

var maxLogSizeBytes int64 = 5 << 20 // 5 MiB

// Event is a single structured audit entry.
type Event struct {
	Timestamp string         `json:"timestamp"`
	RunID     string         `json:"run_id"`
	Command   string         `json:"command"`
	Action    string         `json:"action"`
	Status    string         `json:"status"`
	Skill     string         `json:"skill,omitempty"`
	ClientID  string         `json:"client_id,omitempty"`
	Scope     string         `json:"scope,omitempty"`
	Target    string         `json:"target_path,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
	Error     string         `json:"error,omitempty"`
}

// Logger writes append-only JSONL audit events.
type Logger struct {
	mu      sync.Mutex
	enabled bool
	path    string
	runID   string
	command string
}

// New creates a new logger rooted at appDir unless overridden by env vars.
func New(appDir, command string) *Logger {
	if isDisabled() {
		return &Logger{enabled: false}
	}

	logPath := strings.TrimSpace(os.Getenv("AISK_AUDIT_LOG_PATH"))
	if logPath == "" {
		logPath = filepath.Join(appDir, defaultLogFilename)
	}

	return &Logger{
		enabled: true,
		path:    logPath,
		runID:   newRunID(),
		command: command,
	}
}

// Log writes an event entry. Errors are intentionally ignored.
func (l *Logger) Log(action, status string, details map[string]any, err error) {
	l.LogEvent(Event{
		Action:  action,
		Status:  status,
		Details: details,
		Error:   errorString(err),
	})
}

// LogEvent writes a full event with optional fields.
func (l *Logger) LogEvent(e Event) {
	if l == nil || !l.enabled {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(l.path), 0o755); err != nil {
		return
	}
	if err := rotateIfNeeded(l.path); err != nil {
		return
	}

	e.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	e.RunID = l.runID
	e.Command = l.command
	e.Error = strings.TrimSpace(e.Error)

	line, err := json.Marshal(e)
	if err != nil {
		return
	}

	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()

	_, _ = f.Write(append(line, '\n'))
}

// RunID returns the current invocation identifier.
func (l *Logger) RunID() string {
	if l == nil {
		return ""
	}
	return l.runID
}

func isDisabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("AISK_AUDIT_ENABLED")))
	switch v {
	case "", "1", "true", "yes", "on":
		return false
	default:
		return true
	}
}

func newRunID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err == nil {
		return hex.EncodeToString(b[:])
	}
	return time.Now().UTC().Format("20060102150405.000000000")
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func rotateIfNeeded(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.Size() < maxLogSizeBytes {
		return nil
	}

	backup := path + ".1"
	_ = os.Remove(backup)
	return os.Rename(path, backup)
}
