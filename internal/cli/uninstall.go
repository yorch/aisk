package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yorch/aisk/internal/adapter"
	"github.com/yorch/aisk/internal/client"
	"github.com/yorch/aisk/internal/config"
	"github.com/yorch/aisk/internal/gitignore"
	"github.com/yorch/aisk/internal/manifest"
	"github.com/yorch/aisk/internal/skill"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <skill>",
	Short: "Remove a skill from one or all AI clients",
	Args:  cobra.ExactArgs(1),
	RunE:  runUninstall,
}

var uninstallClient string

func init() {
	uninstallCmd.Flags().StringVar(&uninstallClient, "client", "", "specific client to uninstall from")
}

func runUninstall(_ *cobra.Command, args []string) error {
	paths, err := config.ResolvePaths()
	if err != nil {
		return err
	}

	skillArg := args[0]

	// Load manifest to find installations
	m, err := manifest.Load(paths.ManifestDB)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	installations := m.Find(skillArg, uninstallClient)
	if len(installations) == 0 {
		// Try matching by directory name via skill scan
		skills, _ := skill.ScanLocal(paths.SkillsRepo)
		for _, s := range skills {
			if s.DirName == skillArg {
				installations = m.Find(s.Frontmatter.Name, uninstallClient)
				skillArg = s.Frontmatter.Name
				break
			}
		}
	}

	if len(installations) == 0 {
		return fmt.Errorf("no installations found for %q", skillArg)
	}

	// We need a minimal Skill for uninstall operations
	stub := &skill.Skill{}
	stub.Frontmatter.Name = skillArg

	// Try to find actual skill for DirName
	skills, _ := skill.ScanLocal(paths.SkillsRepo)
	for _, s := range skills {
		if s.Frontmatter.Name == skillArg || s.DirName == skillArg {
			stub = s
			break
		}
	}

	lock := manifest.NewLock(paths.ManifestDB)
	if err := lock.Acquire(5 * time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not acquire lock: %v\n", err)
	} else {
		defer lock.Release()
	}

	for _, inst := range installations {
		clientID := client.ParseClientID(inst.ClientID)
		adp, err := adapter.ForClient(clientID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: no adapter for %s: %v\n", inst.ClientID, err)
			continue
		}

		if err := adp.Uninstall(stub, inst.InstallPath); err != nil {
			fmt.Fprintf(os.Stderr, "warning: uninstall from %s: %v\n", inst.ClientID, err)
			continue
		}

		m.Remove(inst.SkillName, inst.ClientID, inst.Scope)
		fmt.Printf("Uninstalled %q from %s\n", inst.SkillName, inst.ClientID)
	}

	if err := m.Save(); err != nil {
		return fmt.Errorf("saving manifest: %w", err)
	}

	// Clean up .gitignore for project-scope uninstalls
	manageGitignoreOnUninstall(m, installations)

	return nil
}

func manageGitignoreOnUninstall(m *manifest.Manifest, removed []manifest.Installation) {
	// Collect client IDs from removed project-scope installations
	removedClients := make(map[string]bool)
	for _, inst := range removed {
		if inst.Scope == "project" {
			removedClients[inst.ClientID] = true
		}
	}
	if len(removedClients) == 0 {
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		return
	}
	projectRoot := config.FindProjectRoot(cwd)
	if projectRoot == "" {
		return
	}

	// Check which clients still have project-scope installs
	remaining := m.FindByScope("project")
	stillUsed := make(map[string]bool)
	for _, inst := range remaining {
		if isInstallationInProject(inst, projectRoot) {
			stillUsed[inst.ClientID] = true
		}
	}

	// Remove patterns for clients that no longer have project-scope installs
	var patternsToRemove []string
	for clientID := range removedClients {
		if !stillUsed[clientID] {
			patterns := gitignore.GitignorePatternsForClient(clientID, "")
			patternsToRemove = append(patternsToRemove, patterns...)
		}
	}

	if len(patternsToRemove) == 0 {
		return
	}

	giPath := filepath.Join(projectRoot, ".gitignore")
	removedEntries, err := gitignore.RemoveEntries(giPath, patternsToRemove)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not update .gitignore: %v\n", err)
		return
	}
	for _, r := range removedEntries {
		fmt.Printf("Removed %s from .gitignore\n", r)
	}
}

func isInstallationInProject(inst manifest.Installation, projectRoot string) bool {
	if inst.Scope != "project" {
		return false
	}
	if !filepath.IsAbs(inst.InstallPath) {
		// Backward compatibility for older manifests that stored relative project paths.
		// Assume these entries belong to the current project context.
		return true
	}
	rel, err := filepath.Rel(projectRoot, inst.InstallPath)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}
