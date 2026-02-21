package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindProjectRoot_Git(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0o755)
	sub := filepath.Join(dir, "src", "pkg")
	os.MkdirAll(sub, 0o755)

	root := FindProjectRoot(sub)
	if root != dir {
		t.Errorf("expected %s, got %s", dir, root)
	}
}

func TestFindProjectRoot_GoMod(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0o644)

	root := FindProjectRoot(dir)
	if root != dir {
		t.Errorf("expected %s, got %s", dir, root)
	}
}

func TestFindProjectRoot_Nested(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0o755)
	nested := filepath.Join(dir, "a", "b", "c")
	os.MkdirAll(nested, 0o755)

	root := FindProjectRoot(nested)
	if root != dir {
		t.Errorf("expected %s, got %s", dir, root)
	}
}

func TestFindProjectRoot_NotFound(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "empty")
	os.MkdirAll(sub, 0o755)

	root := FindProjectRoot(sub)
	// The temp dir itself won't have markers, but we can't guarantee
	// nothing above it does, so just check it's either "" or a valid path
	if root == sub {
		t.Error("should not return the start dir when it has no markers")
	}
}

func TestFindProjectRoot_MultipleMarkers(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".git"), 0o755)
	os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0o644)

	root := FindProjectRoot(dir)
	if root != dir {
		t.Errorf("expected %s, got %s", dir, root)
	}
}
