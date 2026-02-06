package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadFullContent reads the SKILL.md body and optionally inlines reference files.
func ReadFullContent(s *Skill, includeRefs bool) (string, error) {
	var b strings.Builder

	b.WriteString(s.MarkdownBody)

	if includeRefs && len(s.ReferenceFiles) > 0 {
		b.WriteString("\n\n---\n\n")
		for _, ref := range s.ReferenceFiles {
			absPath := filepath.Join(s.Path, ref)
			data, err := os.ReadFile(absPath)
			if err != nil {
				return "", fmt.Errorf("reading reference %s: %w", ref, err)
			}
			// Extract filename without extension for header
			name := filepath.Base(ref)
			name = strings.TrimSuffix(name, filepath.Ext(name))
			b.WriteString(fmt.Sprintf("## Reference: %s\n\n", name))
			b.WriteString(string(data))
			b.WriteString("\n\n")
		}
	}

	return b.String(), nil
}
