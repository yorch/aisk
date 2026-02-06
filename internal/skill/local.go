package skill

import (
	"os"
	"path/filepath"
	"strings"
)

// ScanLocal discovers skills in the given directory.
// It looks for subdirectories containing a SKILL.md file.
func ScanLocal(repoPath string) ([]*Skill, error) {
	entries, err := os.ReadDir(repoPath)
	if err != nil {
		return nil, err
	}

	var skills []*Skill
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Skip hidden dirs and common non-skill dirs
		if strings.HasPrefix(entry.Name(), ".") || entry.Name() == "node_modules" {
			continue
		}

		skillDir := filepath.Join(repoPath, entry.Name())
		skillFile := filepath.Join(skillDir, "SKILL.md")

		data, err := os.ReadFile(skillFile)
		if err != nil {
			continue // no SKILL.md, not a skill directory
		}

		fm, body, err := ParseFrontmatter(string(data))
		if err != nil {
			continue // malformed SKILL.md, skip
		}

		s := &Skill{
			Frontmatter:  fm,
			DirName:      entry.Name(),
			Path:         skillDir,
			Source:       SourceLocal,
			MarkdownBody: body,
		}

		// Discover reference files (check both singular and plural)
		s.ReferenceFiles = discoverFiles(skillDir, "reference")
		if len(s.ReferenceFiles) == 0 {
			s.ReferenceFiles = discoverFiles(skillDir, "references")
		}

		// Discover example files
		s.ExampleFiles = discoverFiles(skillDir, "examples")

		// Discover asset files
		s.AssetFiles = discoverFiles(skillDir, "assets")

		skills = append(skills, s)
	}

	return skills, nil
}

// discoverFiles lists files recursively under a subdirectory, returning relative paths.
func discoverFiles(skillDir, subdir string) []string {
	dir := filepath.Join(skillDir, subdir)
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return nil
	}

	var files []string
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(skillDir, path)
		if err != nil {
			return nil
		}
		files = append(files, rel)
		return nil
	})

	return files
}
