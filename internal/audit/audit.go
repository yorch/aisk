package audit

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultLogFilename = "audit.log"
)

var maxLogSizeBytes int64 = 5 << 20 // 5 MiB
var maxBackups = 3

var sensitiveKeyPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)token`),
	regexp.MustCompile(`(?i)secret`),
	regexp.MustCompile(`(?i)password`),
	regexp.MustCompile(`(?i)authorization`),
	regexp.MustCompile(`(?i)api[_-]?key`),
}

var bearerTokenPattern = regexp.MustCompile(`(?i)\b(bearer\s+)[A-Za-z0-9\-\._~\+\/]+=*`)
var keyValueSecretPattern = regexp.MustCompile(`(?i)\b(token|secret|password|api[_-]?key|authorization)\s*[:=]\s*([^\s,;]+)`)

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
	e = sanitizeEvent(e)

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
	maxSize := configuredMaxLogSizeBytes()
	if info.Size() < maxSize {
		return nil
	}

	backups := configuredMaxBackups()
	if backups <= 0 {
		return os.Remove(path)
	}

	// Shift older backups up: .2 -> .3, .1 -> .2
	oldest := backupPath(path, backups)
	_ = os.Remove(oldest)
	for i := backups - 1; i >= 1; i-- {
		src := backupPath(path, i)
		dst := backupPath(path, i+1)
		if _, err := os.Stat(src); err == nil {
			_ = os.Rename(src, dst)
		}
	}
	return os.Rename(path, backupPath(path, 1))
}

func backupPath(path string, n int) string {
	return fmt.Sprintf("%s.%d", path, n)
}

func configuredMaxLogSizeBytes() int64 {
	v := strings.TrimSpace(os.Getenv("AISK_AUDIT_MAX_SIZE_MB"))
	if v == "" {
		return maxLogSizeBytes
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return maxLogSizeBytes
	}
	return int64(n) << 20
}

func configuredMaxBackups() int {
	v := strings.TrimSpace(os.Getenv("AISK_AUDIT_MAX_BACKUPS"))
	if v == "" {
		return maxBackups
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return maxBackups
	}
	return n
}

// CandidateLogPaths returns existing log files in chronological order.
func CandidateLogPaths(primary string) []string {
	backups := configuredMaxBackups()
	var paths []string
	for i := backups; i >= 1; i-- {
		p := backupPath(primary, i)
		if _, err := os.Stat(p); err == nil {
			paths = append(paths, p)
		}
	}
	if _, err := os.Stat(primary); err == nil {
		paths = append(paths, primary)
	}
	return paths
}

func sanitizeEvent(e Event) Event {
	e.Error = sanitizeString(strings.TrimSpace(e.Error))
	e.Target = sanitizeString(e.Target)
	if e.Details != nil {
		e.Details = sanitizeMap(e.Details)
	}
	return e
}

func sanitizeMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		if isSensitiveKey(k) {
			out[k] = "[REDACTED]"
			continue
		}
		out[k] = sanitizeValue(v)
	}
	return out
}

func sanitizeValue(v any) any {
	switch t := v.(type) {
	case string:
		return sanitizeString(t)
	case map[string]any:
		return sanitizeMap(t)
	case []any:
		out := make([]any, len(t))
		for i := range t {
			out[i] = sanitizeValue(t[i])
		}
		return out
	default:
		return v
	}
}

func isSensitiveKey(k string) bool {
	for _, p := range sensitiveKeyPatterns {
		if p.MatchString(k) {
			return true
		}
	}
	return false
}

func sanitizeString(s string) string {
	out := bearerTokenPattern.ReplaceAllString(s, "${1}[REDACTED]")
	out = keyValueSecretPattern.ReplaceAllString(out, "${1}=[REDACTED]")
	return out
}
