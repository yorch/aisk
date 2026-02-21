package gitignore

import (
	"os"
	"strings"
)

const (
	sectionStart = "# aisk managed"
	sectionEnd   = "# end aisk managed"
)

// EnsureEntries adds entries to the aisk-managed section of a .gitignore file.
// Creates the file if it doesn't exist. Returns the entries that were actually added.
func EnsureEntries(gitignorePath string, entries []string) ([]string, error) {
	content, err := readOrEmpty(gitignorePath)
	if err != nil {
		return nil, err
	}

	existing := parseManagedEntries(content)
	var added []string
	for _, e := range entries {
		if !existing[e] {
			added = append(added, e)
		}
	}

	if len(added) == 0 {
		return nil, nil
	}

	// Build new managed section
	allEntries := mergeEntries(existing, entries)
	newContent := replaceManagedSection(content, allEntries)

	if err := os.WriteFile(gitignorePath, []byte(newContent), 0o644); err != nil {
		return nil, err
	}

	return added, nil
}

// RemoveEntries removes entries from the aisk-managed section.
// Returns the entries that were actually removed. Cleans up empty section.
func RemoveEntries(gitignorePath string, entries []string) ([]string, error) {
	content, err := readOrEmpty(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	existing := parseManagedEntries(content)
	if len(existing) == 0 {
		return nil, nil
	}

	removeSet := make(map[string]bool)
	for _, e := range entries {
		removeSet[e] = true
	}

	var removed []string
	remaining := make(map[string]bool)
	for e := range existing {
		if removeSet[e] {
			removed = append(removed, e)
		} else {
			remaining[e] = true
		}
	}

	if len(removed) == 0 {
		return nil, nil
	}

	var remainingList []string
	for e := range remaining {
		remainingList = append(remainingList, e)
	}

	var newContent string
	if len(remainingList) == 0 {
		// Remove the entire section
		newContent = removeManagedSection(content)
	} else {
		newContent = replaceManagedSection(content, remainingList)
	}

	if err := os.WriteFile(gitignorePath, []byte(newContent), 0o644); err != nil {
		return nil, err
	}

	return removed, nil
}

// GitignorePatternsForClient returns gitignore patterns for a given client's install path.
func GitignorePatternsForClient(clientID, installPath string) []string {
	switch clientID {
	case "claude":
		return []string{".claude/skills/"}
	case "cursor":
		return []string{".cursor/rules/"}
	case "windsurf":
		return []string{".windsurf/rules/"}
	case "copilot":
		return []string{".github/copilot-instructions.md"}
	case "gemini":
		return []string{"GEMINI.md"}
	case "codex":
		return []string{"AGENTS.md"}
	default:
		if installPath != "" {
			return []string{installPath}
		}
		return nil
	}
}

func readOrEmpty(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(data), nil
}

func parseManagedEntries(content string) map[string]bool {
	entries := make(map[string]bool)
	inSection := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == sectionStart {
			inSection = true
			continue
		}
		if trimmed == sectionEnd {
			inSection = false
			continue
		}
		if inSection && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			entries[trimmed] = true
		}
	}
	return entries
}

func mergeEntries(existing map[string]bool, newEntries []string) []string {
	merged := make(map[string]bool)
	for e := range existing {
		merged[e] = true
	}
	for _, e := range newEntries {
		merged[e] = true
	}
	var result []string
	for e := range merged {
		result = append(result, e)
	}
	// Sort for deterministic output
	sortStrings(result)
	return result
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

func replaceManagedSection(content string, entries []string) string {
	section := buildSection(entries)

	// If section already exists, replace it
	startIdx := strings.Index(content, sectionStart)
	endIdx := strings.Index(content, sectionEnd)
	if startIdx >= 0 && endIdx >= 0 {
		before := content[:startIdx]
		after := content[endIdx+len(sectionEnd):]
		after = strings.TrimLeft(after, "\n")
		result := before + section
		if after != "" {
			result += "\n" + after
		}
		return result
	}

	// Append new section
	result := strings.TrimRight(content, "\n")
	if result != "" {
		result += "\n\n"
	}
	result += section + "\n"
	return result
}

func removeManagedSection(content string) string {
	startIdx := strings.Index(content, sectionStart)
	endIdx := strings.Index(content, sectionEnd)
	if startIdx < 0 || endIdx < 0 {
		return content
	}

	before := strings.TrimRight(content[:startIdx], "\n")
	after := strings.TrimLeft(content[endIdx+len(sectionEnd):], "\n")

	if before == "" && after == "" {
		return ""
	}
	if before == "" {
		return after
	}
	if after == "" {
		return before + "\n"
	}
	return before + "\n\n" + after
}

func buildSection(entries []string) string {
	var b strings.Builder
	b.WriteString(sectionStart)
	b.WriteString("\n")
	for _, e := range entries {
		b.WriteString(e)
		b.WriteString("\n")
	}
	b.WriteString(sectionEnd)
	return b.String()
}
