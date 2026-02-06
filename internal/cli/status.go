package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yorch/aisk/internal/config"
	"github.com/yorch/aisk/internal/manifest"
	"github.com/yorch/aisk/internal/tui"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show installed skills per client",
	RunE:  runStatus,
}

var statusJSON bool

func init() {
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "output as JSON")
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
	return nil
}
