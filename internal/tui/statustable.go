package tui

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/lipgloss"
	"github.com/yorch/aisk/internal/client"
	"github.com/yorch/aisk/internal/manifest"
)

// StatusEntry represents one row in the status table.
type StatusEntry struct {
	SkillName    string
	SkillVersion string
	Installations map[client.ClientID]string // clientID -> installed version
}

// BuildStatusEntries creates status entries from manifest data.
func BuildStatusEntries(m *manifest.Manifest) []StatusEntry {
	// Group by skill name
	skillMap := make(map[string]*StatusEntry)
	var order []string

	for _, inst := range m.Installations {
		entry, ok := skillMap[inst.SkillName]
		if !ok {
			entry = &StatusEntry{
				SkillName:     inst.SkillName,
				SkillVersion:  inst.SkillVersion,
				Installations: make(map[client.ClientID]string),
			}
			skillMap[inst.SkillName] = entry
			order = append(order, inst.SkillName)
		}
		entry.Installations[client.ClientID(inst.ClientID)] = inst.SkillVersion
	}

	result := make([]StatusEntry, len(order))
	for i, name := range order {
		result[i] = *skillMap[name]
	}
	return result
}

// PrintStatusTable prints a formatted status table.
func PrintStatusTable(entries []StatusEntry) {
	if len(entries) == 0 {
		fmt.Println(lipgloss.NewStyle().Foreground(Gray).Render("No skills installed."))
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Header
	header := "SKILL\tCLAUDE\tGEMINI\tCODEX\tCOPILOT\tCURSOR\tWINDSURF"
	fmt.Fprintln(w, header)

	// Separator
	sep := strings.Repeat("-", 12) + "\t" +
		strings.Repeat("-", 8) + "\t" +
		strings.Repeat("-", 8) + "\t" +
		strings.Repeat("-", 8) + "\t" +
		strings.Repeat("-", 8) + "\t" +
		strings.Repeat("-", 8) + "\t" +
		strings.Repeat("-", 8)
	fmt.Fprintln(w, sep)

	for _, entry := range entries {
		skillLabel := fmt.Sprintf("%s (%s)", entry.SkillName, entry.SkillVersion)

		var cells []string
		cells = append(cells, skillLabel)

		for _, id := range client.AllClientIDs {
			ver, ok := entry.Installations[id]
			if ok {
				cells = append(cells, ver)
			} else {
				cells = append(cells, "")
			}
		}

		fmt.Fprintln(w, strings.Join(cells, "\t"))
	}

	w.Flush()
}
