package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Lock provides basic file-based locking for manifest access.
type Lock struct {
	path string
}

// NewLock creates a lock for the given manifest path.
func NewLock(manifestPath string) *Lock {
	return &Lock{path: manifestPath + ".lock"}
}

// Acquire tries to acquire the lock, waiting up to timeout.
func (l *Lock) Acquire(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		f, err := os.OpenFile(l.path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			_ = f.Close()
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("could not acquire lock %s within %v", filepath.Base(l.path), timeout)
		}

		// Check if lock is stale (older than 30 seconds)
		info, err := os.Stat(l.path)
		if err == nil && time.Since(info.ModTime()) > 30*time.Second {
			_ = os.Remove(l.path)
			continue
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// Release removes the lock file.
func (l *Lock) Release() {
	_ = os.Remove(l.path)
}
