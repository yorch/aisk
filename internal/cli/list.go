package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/yorch/aisk/internal/config"
	"github.com/yorch/aisk/internal/skill"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available skills",
	RunE:  runList,
}

var (
	listRemote bool
	listJSON   bool
	listRepo   string
)

func init() {
	listCmd.Flags().BoolVar(&listRemote, "remote", false, "also fetch remote skills from GitHub")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "output as JSON")
	listCmd.Flags().StringVar(&listRepo, "repo", "", "GitHub repo to fetch from (owner/repo)")
}

func runList(_ *cobra.Command, _ []string) error {
	paths, err := config.ResolvePaths()
	if err != nil {
		return err
	}

	skills, err := skill.ScanLocal(paths.SkillsRepo)
	if err != nil {
		return fmt.Errorf("scanning skills: %w", err)
	}

	if listRemote {
		repo := listRepo
		if repo == "" {
			repo = os.Getenv("AISK_REMOTE_REPO")
		}
		if repo != "" {
			parts := strings.SplitN(repo, "/", 2)
			if len(parts) == 2 {
				fmt.Fprintf(os.Stderr, "Fetching skills from %s...\n", repo)
				remote, err := skill.FetchRemoteList(parts[0], parts[1])
				if err != nil {
					fmt.Fprintf(os.Stderr, "warning: remote fetch failed: %v\n", err)
				} else {
					skills = append(skills, remote...)
				}
			}
		} else {
			fmt.Fprintln(os.Stderr, "hint: set --repo or AISK_REMOTE_REPO to fetch remote skills")
		}
	}

	if len(skills) == 0 {
		fmt.Println("No skills found.")
		fmt.Printf("Set AISK_SKILLS_PATH or run from a directory containing skill folders.\n")
		return nil
	}

	if listJSON {
		return printSkillsJSON(skills)
	}

	return printSkillsTable(skills)
}

func printSkillsTable(skills []*skill.Skill) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "NAME\tVERSION\tDIRECTORY\tSOURCE\n")
	for _, s := range skills {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			s.Frontmatter.Name,
			s.DisplayVersion(),
			s.DirName,
			s.Source,
		)
	}
	return w.Flush()
}

type skillJSON struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	DirName     string   `json:"dir_name"`
	Source      string   `json:"source"`
	References  []string `json:"references,omitempty"`
	Examples    []string `json:"examples,omitempty"`
}

func printSkillsJSON(skills []*skill.Skill) error {
	items := make([]skillJSON, len(skills))
	for i, s := range skills {
		items[i] = skillJSON{
			Name:        s.Frontmatter.Name,
			Version:     s.DisplayVersion(),
			Description: s.Frontmatter.Description,
			DirName:     s.DirName,
			Source:      s.Source.String(),
			References:  s.ReferenceFiles,
			Examples:    s.ExampleFiles,
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(items)
}
