package client

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectAll_WithConfigDirs(t *testing.T) {
	home := t.TempDir()

	// Create config directories for some clients
	os.MkdirAll(filepath.Join(home, ".claude"), 0o755)
	os.MkdirAll(filepath.Join(home, ".gemini"), 0o755)
	os.MkdirAll(filepath.Join(home, ".cursor"), 0o755)

	reg := NewRegistry()
	DetectAll(reg, home)

	// Claude should be detected (config dir exists)
	c := reg.Get(Claude)
	if !c.Detected {
		t.Error("Claude should be detected with .claude/ dir")
	}
	if c.GlobalPath == "" {
		t.Error("Claude GlobalPath should be set")
	}

	// Gemini should be detected
	g := reg.Get(Gemini)
	if !g.Detected {
		t.Error("Gemini should be detected with .gemini/ dir")
	}

	// Cursor should be detected
	cu := reg.Get(Cursor)
	if !cu.Detected {
		t.Error("Cursor should be detected with .cursor/ dir")
	}
}

func TestDetectAll_NoConfigDirs(t *testing.T) {
	home := t.TempDir()
	// Don't create any config dirs

	reg := NewRegistry()
	DetectAll(reg, home)

	// Clients may still be detected if their binary is in PATH
	// We can't reliably test that, but we can verify structure
	for _, c := range reg.All() {
		if c.ID == "" {
			t.Error("client ID should not be empty")
		}
		if c.Name == "" {
			t.Error("client Name should not be empty")
		}
	}
}

func TestParseClientID(t *testing.T) {
	tests := []struct {
		input string
		want  ClientID
	}{
		{"claude", Claude},
		{"gemini", Gemini},
		{"codex", Codex},
		{"copilot", Copilot},
		{"cursor", Cursor},
		{"windsurf", Windsurf},
		{"unknown", ""},
		{"", ""},
	}

	for _, tt := range tests {
		got := ParseClientID(tt.input)
		if got != tt.want {
			t.Errorf("ParseClientID(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestRegistry_Detected(t *testing.T) {
	reg := NewRegistry()
	home := t.TempDir()

	os.MkdirAll(filepath.Join(home, ".claude"), 0o755)
	DetectAll(reg, home)

	detected := reg.Detected()
	foundClaude := false
	for _, c := range detected {
		if c.ID == Claude {
			foundClaude = true
		}
	}

	if !foundClaude {
		t.Error("Claude should be in detected list")
	}
}
