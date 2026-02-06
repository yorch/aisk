package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/yorch/aisk/internal/client"
	"github.com/yorch/aisk/internal/config"
)

var clientsCmd = &cobra.Command{
	Use:   "clients",
	Short: "Show detected AI clients",
	RunE:  runClients,
}

var clientsJSON bool

func init() {
	clientsCmd.Flags().BoolVar(&clientsJSON, "json", false, "output as JSON")
}

func runClients(_ *cobra.Command, _ []string) error {
	paths, err := config.ResolvePaths()
	if err != nil {
		return err
	}

	reg := client.NewRegistry()
	client.DetectAll(reg, paths.Home)

	if clientsJSON {
		return printClientsJSON(reg)
	}

	return printClientsTable(reg)
}

func printClientsTable(reg *client.Registry) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "CLIENT\tDETECTED\tGLOBAL PATH\tPROJECT PATH\n")

	for _, c := range reg.All() {
		detected := " "
		if c.Detected {
			detected = "*"
		}

		globalPath := c.GlobalPath
		if globalPath == "" {
			globalPath = "(n/a)"
		}
		projectPath := c.ProjectPath
		if projectPath == "" {
			projectPath = "(n/a)"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			c.Name,
			detected,
			globalPath,
			projectPath,
		)
	}

	return w.Flush()
}

type clientJSON struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Detected        bool   `json:"detected"`
	GlobalPath      string `json:"global_path,omitempty"`
	ProjectPath     string `json:"project_path,omitempty"`
	SupportsGlobal  bool   `json:"supports_global"`
	SupportsProject bool   `json:"supports_project"`
}

func printClientsJSON(reg *client.Registry) error {
	var items []clientJSON
	for _, c := range reg.All() {
		items = append(items, clientJSON{
			ID:              string(c.ID),
			Name:            c.Name,
			Detected:        c.Detected,
			GlobalPath:      c.GlobalPath,
			ProjectPath:     c.ProjectPath,
			SupportsGlobal:  c.SupportsGlobal,
			SupportsProject: c.SupportsProject,
		})
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(items)
}
