package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/yorch/aisk/internal/audit"
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

func runList(_ *cobra.Command, _ []string) (retErr error) {
	paths, err := config.ResolvePaths()
	if err != nil {
		return err
	}
	al := audit.New(paths.AiskDir, "list")
	al.Log("command.list", "started", map[string]any{
		"remote": listRemote,
		"repo":   listRepo,
		"json":   listJSON,
	}, nil)
	defer func() {
		status := "success"
		if retErr != nil {
			status = "error"
		}
		al.Log("command.list", status, nil, retErr)
	}()

	skills, err := skill.ScanLocal(paths.SkillsRepo)
	if err != nil {
		al.Log("skill.scan_local", "error", map[string]any{"path": paths.SkillsRepo}, err)
		return fmt.Errorf("scanning skills: %w", err)
	}
	al.Log("skill.scan_local", "success", map[string]any{"path": paths.SkillsRepo, "count": len(skills)}, nil)

	if listRemote {
		repo := listRepo
		if repo == "" {
			repo = os.Getenv("AISK_REMOTE_REPO")
		}
		if repo != "" {
			parts := strings.SplitN(repo, "/", 2)
			if len(parts) == 2 {
				al.Log("list.remote.fetch", "started", map[string]any{"repo": repo}, nil)
				fmt.Fprintf(os.Stderr, "Fetching skills from %s...\n", repo)
				remote, err := skill.FetchRemoteList(parts[0], parts[1])
				if err != nil {
					al.Log("list.remote.fetch", "error", map[string]any{"repo": repo}, err)
					fmt.Fprintf(os.Stderr, "warning: remote fetch failed: %v\n", err)
				} else {
					skills = append(skills, remote...)
					al.Log("list.remote.fetch", "success", map[string]any{"repo": repo, "count": len(remote)}, nil)
				}
			}
		} else {
			al.Log("list.remote.fetch", "skipped", map[string]any{"reason": "missing repo"}, nil)
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
