package client

import (
	"os"
	"os/exec"
	"path/filepath"
)

// DetectAll runs detection for all clients in the registry.
func DetectAll(reg *Registry, home string) {
	detectors := map[ClientID]func(*Client, string){
		Claude:   detectClaude,
		Gemini:   detectGemini,
		Codex:    detectCodex,
		Copilot:  detectCopilot,
		Cursor:   detectCursor,
		Windsurf: detectWindsurf,
	}

	for id, detect := range detectors {
		c := reg.Get(id)
		detect(c, home)
	}
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func binaryExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func detectClaude(c *Client, home string) {
	configDir := filepath.Join(home, ".claude")
	c.Detected = dirExists(configDir) || binaryExists("claude")
	if c.Detected {
		c.GlobalPath = filepath.Join(configDir, "skills")
		c.ProjectPath = filepath.Join(".claude", "skills")
	}
}

func detectGemini(c *Client, home string) {
	configDir := filepath.Join(home, ".gemini")
	c.Detected = dirExists(configDir) || binaryExists("gemini")
	if c.Detected {
		c.GlobalPath = filepath.Join(configDir, "GEMINI.md")
		c.ProjectPath = "GEMINI.md"
	}
}

func detectCodex(c *Client, home string) {
	configDir := filepath.Join(home, ".codex")
	c.Detected = dirExists(configDir) || binaryExists("codex")
	if c.Detected {
		c.GlobalPath = filepath.Join(configDir, "instructions.md")
		c.ProjectPath = "AGENTS.md"
	}
}

func detectCopilot(c *Client, home string) {
	configDir := filepath.Join(home, ".vscode")
	c.Detected = dirExists(configDir) || binaryExists("code")
	if c.Detected {
		c.ProjectPath = filepath.Join(".github", "copilot-instructions.md")
	}
}

func detectCursor(c *Client, home string) {
	configDir := filepath.Join(home, ".cursor")
	c.Detected = dirExists(configDir) || binaryExists("cursor")
	if c.Detected {
		c.ProjectPath = filepath.Join(".cursor", "rules")
	}
}

func detectWindsurf(c *Client, home string) {
	configDir := filepath.Join(home, ".codeium", "windsurf")
	c.Detected = dirExists(configDir) || binaryExists("windsurf")
	if c.Detected {
		c.GlobalPath = filepath.Join(configDir, "memories", "global_rules.md")
		c.ProjectPath = filepath.Join(".windsurf", "rules")
	}
}
