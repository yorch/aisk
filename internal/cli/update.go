package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/yorch/aisk/internal/adapter"
	"github.com/yorch/aisk/internal/audit"
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

func runUpdate(_ *cobra.Command, args []string) (retErr error) {
	paths, err := config.ResolvePaths()
	if err != nil {
		return err
	}
	al := audit.New(paths.AiskDir, "update")
	al.Log("command.update", "started", map[string]any{
		"args":   args,
		"client": updateClient,
	}, nil)
	defer func() {
		status := "success"
		if retErr != nil {
			status = "error"
		}
		al.Log("command.update", status, nil, retErr)
	}()

	m, err := manifest.Load(paths.ManifestDB)
	if err != nil {
		al.Log("manifest.load", "error", nil, err)
		return fmt.Errorf("loading manifest: %w", err)
	}
	al.Log("manifest.load", "success", map[string]any{"installations": len(m.Installations)}, nil)

	skills, err := skill.ScanLocal(paths.SkillsRepo)
	if err != nil {
		al.Log("skill.scan_local", "error", map[string]any{"path": paths.SkillsRepo}, err)
		return fmt.Errorf("scanning skills: %w", err)
	}
	al.Log("skill.scan_local", "success", map[string]any{"path": paths.SkillsRepo, "count": len(skills)}, nil)

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
		al.Log("update.targets.resolve", "success", map[string]any{"count": 0}, nil)
		return nil
	}
	al.Log("update.targets.resolve", "success", map[string]any{"count": len(targets)}, nil)

	lock := manifest.NewLock(paths.ManifestDB)
	al.Log("manifest.lock", "started", map[string]any{"path": paths.ManifestDB + ".lock"}, nil)
	if err := lock.Acquire(5 * time.Second); err != nil {
		al.Log("manifest.lock", "error", nil, err)
		fmt.Fprintf(os.Stderr, "warning: could not acquire lock: %v\n", err)
	} else {
		al.Log("manifest.lock", "success", nil, nil)
		defer lock.Release()
		defer al.Log("manifest.lock", "released", nil, nil)
	}

	updated := 0
	for _, inst := range targets {
		s := skillMap[inst.SkillName]
		if s == nil {
			fmt.Fprintf(os.Stderr, "warning: skill %q not found in repo, skipping\n", inst.SkillName)
			al.LogEvent(audit.Event{
				Action:   "update.adapter.apply",
				Status:   "skipped",
				Skill:    inst.SkillName,
				ClientID: inst.ClientID,
				Scope:    inst.Scope,
				Target:   inst.InstallPath,
				Error:    "skill not found in local repo",
			})
			continue
		}

		clientID := client.ParseClientID(inst.ClientID)
		adp, err := adapter.ForClient(clientID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: no adapter for %s\n", inst.ClientID)
			al.LogEvent(audit.Event{
				Action:   "update.adapter.apply",
				Status:   "error",
				Skill:    inst.SkillName,
				ClientID: inst.ClientID,
				Scope:    inst.Scope,
				Target:   inst.InstallPath,
				Error:    err.Error(),
			})
			continue
		}

		opts := adapter.InstallOpts{
			Scope: inst.Scope,
		}

		al.LogEvent(audit.Event{
			Action:   "update.adapter.apply",
			Status:   "started",
			Skill:    inst.SkillName,
			ClientID: inst.ClientID,
			Scope:    inst.Scope,
			Target:   inst.InstallPath,
		})
		if err := adp.Install(s, inst.InstallPath, opts); err != nil {
			fmt.Fprintf(os.Stderr, "error updating %s on %s: %v\n", inst.SkillName, inst.ClientID, err)
			al.LogEvent(audit.Event{
				Action:   "update.adapter.apply",
				Status:   "error",
				Skill:    inst.SkillName,
				ClientID: inst.ClientID,
				Scope:    inst.Scope,
				Target:   inst.InstallPath,
				Error:    err.Error(),
			})
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
		al.LogEvent(audit.Event{
			Action:   "update.adapter.apply",
			Status:   "success",
			Skill:    inst.SkillName,
			ClientID: inst.ClientID,
			Scope:    inst.Scope,
			Target:   inst.InstallPath,
			Details: map[string]any{
				"from_version": inst.SkillVersion,
				"to_version":   s.DisplayVersion(),
			},
		})
	}

	if err := m.Save(); err != nil {
		al.Log("manifest.save", "error", nil, err)
		return fmt.Errorf("saving manifest: %w", err)
	}
	al.Log("manifest.save", "success", map[string]any{"installations": len(m.Installations), "updated": updated}, nil)

	fmt.Printf("\n%d installation(s) updated.\n", updated)
	return nil
}
