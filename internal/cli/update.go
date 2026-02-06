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

var updateCmd = &cobra.Command{
	Use:   "update [skill]",
	Short: "Re-install a skill with the latest version",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runUpdate,
}

var updateClient string

func init() {
	updateCmd.Flags().StringVar(&updateClient, "client", "", "specific client to update")
}

func runUpdate(_ *cobra.Command, args []string) error {
	paths, err := config.ResolvePaths()
	if err != nil {
		return err
	}

	m, err := manifest.Load(paths.ManifestDB)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	skills, err := skill.ScanLocal(paths.SkillsRepo)
	if err != nil {
		return fmt.Errorf("scanning skills: %w", err)
	}

	// Build skill lookup
	skillMap := make(map[string]*skill.Skill)
	for _, s := range skills {
		skillMap[s.Frontmatter.Name] = s
		skillMap[s.DirName] = s
	}

	// Filter installations to update
	var targets []manifest.Installation
	if len(args) > 0 {
		targets = m.Find(args[0], updateClient)
		if len(targets) == 0 {
			// Try by DirName
			if s, ok := skillMap[args[0]]; ok {
				targets = m.Find(s.Frontmatter.Name, updateClient)
			}
		}
	} else {
		targets = m.Installations
		if updateClient != "" {
			targets = m.FindByClient(updateClient)
		}
	}

	if len(targets) == 0 {
		fmt.Println("No matching installations to update.")
		return nil
	}

	lock := manifest.NewLock(paths.ManifestDB)
	if err := lock.Acquire(5 * time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not acquire lock: %v\n", err)
	} else {
		defer lock.Release()
	}

	updated := 0
	for _, inst := range targets {
		s := skillMap[inst.SkillName]
		if s == nil {
			fmt.Fprintf(os.Stderr, "warning: skill %q not found in repo, skipping\n", inst.SkillName)
			continue
		}

		clientID := client.ParseClientID(inst.ClientID)
		adp, err := adapter.ForClient(clientID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: no adapter for %s\n", inst.ClientID)
			continue
		}

		opts := adapter.InstallOpts{
			Scope: inst.Scope,
		}

		if err := adp.Install(s, inst.InstallPath, opts); err != nil {
			fmt.Fprintf(os.Stderr, "error updating %s on %s: %v\n", inst.SkillName, inst.ClientID, err)
			continue
		}

		m.Add(manifest.Installation{
			SkillName:    inst.SkillName,
			SkillVersion: s.DisplayVersion(),
			ClientID:     inst.ClientID,
			Scope:        inst.Scope,
			InstalledAt:  inst.InstalledAt,
			UpdatedAt:    time.Now(),
			InstallPath:  inst.InstallPath,
		})

		fmt.Printf("Updated %q on %s (%s -> %s)\n", inst.SkillName, inst.ClientID, inst.SkillVersion, s.DisplayVersion())
		updated++
	}

	if err := m.Save(); err != nil {
		return fmt.Errorf("saving manifest: %w", err)
	}

	fmt.Printf("\n%d installation(s) updated.\n", updated)
	return nil
}
