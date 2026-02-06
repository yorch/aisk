package adapter

import (
	"fmt"

	"github.com/yorch/aisk/internal/client"
	"github.com/yorch/aisk/internal/skill"
)

// InstallOpts controls how a skill is installed.
type InstallOpts struct {
	Scope       string // "global" or "project"
	IncludeRefs bool   // inline reference files
	DryRun      bool   // just describe, don't write
}

// Adapter transforms and installs a skill for a specific client.
type Adapter interface {
	Install(s *skill.Skill, targetPath string, opts InstallOpts) error
	Uninstall(s *skill.Skill, targetPath string) error
	Describe(s *skill.Skill, targetPath string, opts InstallOpts) string
}

// ForClient returns the appropriate adapter for the given client ID.
func ForClient(id client.ClientID) (Adapter, error) {
	switch id {
	case client.Claude:
		return &ClaudeAdapter{}, nil
	case client.Gemini:
		return &MarkdownAdapter{ClientName: "Gemini"}, nil
	case client.Codex:
		return &MarkdownAdapter{ClientName: "Codex"}, nil
	case client.Copilot:
		return &MarkdownAdapter{ClientName: "Copilot"}, nil
	case client.Cursor:
		return &CursorAdapter{}, nil
	case client.Windsurf:
		return &WindsurfAdapter{}, nil
	default:
		return nil, fmt.Errorf("no adapter for client %q", id)
	}
}
