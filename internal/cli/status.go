package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yorch/aisk/internal/audit"
	"github.com/yorch/aisk/internal/config"
	"github.com/yorch/aisk/internal/manifest"
	"github.com/yorch/aisk/internal/skill"
	"github.com/yorch/aisk/internal/tui"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show installed skills per client",
	RunE:  runStatus,
}

var (
	statusJSON         bool
	statusCheckUpdates bool
)

func init() {
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "output as JSON")
	statusCmd.Flags().BoolVar(&statusCheckUpdates, "check-updates", true, "check for available updates")
}

func runStatus(_ *cobra.Command, _ []string) (retErr error) {
	paths, err := config.ResolvePaths()
	if err != nil {
		return err
	}
	al := audit.New(paths.AiskDir, "status")
	al.Log("command.status", "started", map[string]any{
		"json":          statusJSON,
		"check_updates": statusCheckUpdates,
	}, nil)
	defer func() {
		status := "success"
		if retErr != nil {
			status = "error"
		}
		al.Log("command.status", status, nil, retErr)
	}()

	m, err := manifest.Load(paths.ManifestDB)
	if err != nil {
		al.Log("manifest.load", "error", nil, err)
		return fmt.Errorf("loading manifest: %w", err)
	}
	al.Log("manifest.load", "success", map[string]any{"installations": len(m.Installations)}, nil)

	if len(m.Installations) == 0 {
		fmt.Println("No skills installed.")
		al.Log("status.render", "success", map[string]any{"installations": 0}, nil)
		return nil
	}

	if statusJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		al.Log("status.render", "success", map[string]any{"format": "json", "installations": len(m.Installations)}, nil)
		return enc.Encode(m.Installations)
	}

	entries := tui.BuildStatusEntries(m)
	tui.PrintStatusTable(entries)

	// Check for updates
	if statusCheckUpdates {
		checkAndPrintUpdates(paths, m, al)
	}

	return nil
}

func checkAndPrintUpdates(paths config.Paths, m *manifest.Manifest, al *audit.Logger) {
	al.Log("status.updates.check", "started", nil, nil)
	available, err := skill.ScanLocal(paths.SkillsRepo)
	if err != nil {
		al.Log("status.updates.check", "error", map[string]any{"skills_repo": paths.SkillsRepo}, err)
		fmt.Fprintf(os.Stderr, "\nwarning: could not scan skills repo for updates: %v\n", err)
		return
	}

	updates := skill.CheckUpdates(m.Installations, available)
	if len(updates) > 0 {
		tui.PrintUpdateTable(updates)
	}
	al.Log("status.updates.check", "success", map[string]any{"updates": len(updates)}, nil)
}
