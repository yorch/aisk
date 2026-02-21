package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
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

func runStatus(_ *cobra.Command, _ []string) error {
	paths, err := config.ResolvePaths()
	if err != nil {
		return err
	}

	m, err := manifest.Load(paths.ManifestDB)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	if len(m.Installations) == 0 {
		fmt.Println("No skills installed.")
		return nil
	}

	if statusJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(m.Installations)
	}

	entries := tui.BuildStatusEntries(m)
	tui.PrintStatusTable(entries)

	// Check for updates
	if statusCheckUpdates {
		checkAndPrintUpdates(paths, m)
	}

	return nil
}

func checkAndPrintUpdates(paths config.Paths, m *manifest.Manifest) {
	available, err := skill.ScanLocal(paths.SkillsRepo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nwarning: could not scan skills repo for updates: %v\n", err)
		return
	}

	updates := skill.CheckUpdates(m.Installations, available)
	if len(updates) > 0 {
		tui.PrintUpdateTable(updates)
	}
}
