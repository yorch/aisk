package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/yorch/aisk/internal/adapter"
	"github.com/yorch/aisk/internal/client"
	"github.com/yorch/aisk/internal/config"
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

	return nil
}
