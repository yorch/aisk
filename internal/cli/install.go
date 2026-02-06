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
	"github.com/yorch/aisk/internal/tui"
)

var installCmd = &cobra.Command{
	Use:   "install [skill]",
	Short: "Install a skill to one or more AI clients",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runInstall,
}

var (
	installClient      string
	installScope       string
	installIncludeRefs bool
	installDryRun      bool
)

func init() {
	installCmd.Flags().StringVar(&installClient, "client", "", "target client (claude, gemini, codex, copilot, cursor, windsurf)")
	installCmd.Flags().StringVar(&installScope, "scope", "global", "installation scope (global or project)")
	installCmd.Flags().BoolVar(&installIncludeRefs, "include-refs", false, "inline reference files in output")
	installCmd.Flags().BoolVar(&installDryRun, "dry-run", false, "show what would be done without making changes")
}

func runInstall(_ *cobra.Command, args []string) error {
	paths, err := config.ResolvePaths()
	if err != nil {
		return err
	}

	// Discover available skills
	skills, err := skill.ScanLocal(paths.SkillsRepo)
	if err != nil {
		return fmt.Errorf("scanning skills: %w", err)
	}

	if len(skills) == 0 {
		return fmt.Errorf("no skills found in %s", paths.SkillsRepo)
	}

	// Resolve skill — TUI if no argument
	var target *skill.Skill
	if len(args) == 0 {
		selected, err := tui.RunSkillSelect(skills)
		if err != nil {
			return err
		}
		target = selected
	} else {
		skillArg := args[0]
		for _, s := range skills {
			if s.DirName == skillArg || s.Frontmatter.Name == skillArg {
				target = s
				break
			}
		}
		if target == nil {
			return fmt.Errorf("skill %q not found", skillArg)
		}
	}

	// Detect clients
	reg := client.NewRegistry()
	client.DetectAll(reg, paths.Home)

	// Resolve clients — TUI multi-select if no --client flag
	var targetClients []*client.Client
	if installClient == "" {
		detected := reg.Detected()
		if len(detected) == 0 {
			return fmt.Errorf("no AI clients detected on this system")
		}

		title := fmt.Sprintf("Install %q to:", target.Frontmatter.Name)
		selected, err := tui.RunClientSelect(title, detected)
		if err != nil {
			return err
		}
		if len(selected) == 0 {
			return fmt.Errorf("no clients selected")
		}
		targetClients = selected
	} else {
		clientID := client.ParseClientID(installClient)
		if clientID == "" {
			return fmt.Errorf("unknown client %q (valid: claude, gemini, codex, copilot, cursor, windsurf)", installClient)
		}
		c := reg.Get(clientID)
		if !c.Detected {
			return fmt.Errorf("client %s not detected on this system", c.Name)
		}
		targetClients = []*client.Client{c}
	}

	// Ensure dirs for manifest
	if err := paths.EnsureDirs(); err != nil {
		return err
	}

	// Build progress items
	progressItems := make([]tui.ProgressItem, len(targetClients))
	for i, c := range targetClients {
		tp := resolveTargetPath(c, installScope)
		progressItems[i] = tui.ProgressItem{
			Label:  c.Name,
			Detail: tp,
			Status: tui.StatusPending,
		}
	}

	// Install to each selected client
	lock := manifest.NewLock(paths.ManifestDB)
	if err := lock.Acquire(5 * time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not acquire lock: %v\n", err)
	} else {
		defer lock.Release()
	}

	m, err := manifest.Load(paths.ManifestDB)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	opts := adapter.InstallOpts{
		Scope:       installScope,
		IncludeRefs: installIncludeRefs,
		DryRun:      installDryRun,
	}

	var installed int
	for i, c := range targetClients {
		targetPath := resolveTargetPath(c, installScope)
		if targetPath == "" {
			progressItems[i].Status = tui.StatusError
			progressItems[i].Detail = fmt.Sprintf("does not support %s scope", installScope)
			fmt.Fprintf(os.Stderr, "  %s does not support %s scope, skipping\n", c.Name, installScope)
			continue
		}

		adp, err := adapter.ForClient(c.ID)
		if err != nil {
			progressItems[i].Status = tui.StatusError
			fmt.Fprintf(os.Stderr, "  no adapter for %s: %v\n", c.Name, err)
			continue
		}

		if installDryRun {
			desc := adp.Describe(target, targetPath, opts)
			fmt.Printf("[dry-run] %s: %s\n", c.Name, desc)
			progressItems[i].Status = tui.StatusDone
			installed++
			continue
		}

		progressItems[i].Status = tui.StatusActive
		if err := adp.Install(target, targetPath, opts); err != nil {
			progressItems[i].Status = tui.StatusError
			progressItems[i].Detail = err.Error()
			fmt.Fprintf(os.Stderr, "  error installing to %s: %v\n", c.Name, err)
			continue
		}

		m.Add(manifest.Installation{
			SkillName:    target.Frontmatter.Name,
			SkillVersion: target.DisplayVersion(),
			ClientID:     string(c.ID),
			Scope:        installScope,
			InstalledAt:  time.Now(),
			UpdatedAt:    time.Now(),
			InstallPath:  targetPath,
		})

		progressItems[i].Status = tui.StatusDone
		installed++
	}

	if !installDryRun {
		if err := m.Save(); err != nil {
			return fmt.Errorf("saving manifest: %w", err)
		}
	}

	// Print progress summary
	fmt.Println()
	tui.PrintProgress(fmt.Sprintf("Installing %q", target.Frontmatter.Name), progressItems)
	fmt.Printf("\n%d client(s) done.\n", installed)

	return nil
}

func resolveTargetPath(c *client.Client, scope string) string {
	switch scope {
	case "global":
		if c.SupportsGlobal {
			return c.GlobalPath
		}
	case "project":
		if c.SupportsProject {
			return c.ProjectPath
		}
	}
	return ""
}
