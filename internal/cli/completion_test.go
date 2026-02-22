package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestGenerateCompletion_Bash(t *testing.T) {
	var buf bytes.Buffer
	if err := generateCompletion("bash", &buf); err != nil {
		t.Fatalf("generateCompletion bash error: %v", err)
	}
	out := buf.String()
	if out == "" || !strings.Contains(out, "complete -o default -F") {
		t.Fatalf("unexpected bash completion output")
	}
}

func TestGenerateCompletion_Zsh(t *testing.T) {
	var buf bytes.Buffer
	if err := generateCompletion("zsh", &buf); err != nil {
		t.Fatalf("generateCompletion zsh error: %v", err)
	}
	out := buf.String()
	if out == "" || !strings.Contains(out, "#compdef aisk") {
		t.Fatalf("unexpected zsh completion output")
	}
}

func TestGenerateCompletion_Fish(t *testing.T) {
	var buf bytes.Buffer
	if err := generateCompletion("fish", &buf); err != nil {
		t.Fatalf("generateCompletion fish error: %v", err)
	}
	out := buf.String()
	if out == "" || !strings.Contains(out, "complete -c aisk") {
		t.Fatalf("unexpected fish completion output")
	}
}

func TestGenerateCompletion_UnsupportedShell(t *testing.T) {
	var buf bytes.Buffer
	err := generateCompletion("powershell", &buf)
	if err == nil {
		t.Fatal("expected error for unsupported shell")
	}
}
