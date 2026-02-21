package tui

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/lipgloss"
	"github.com/yorch/aisk/internal/skill"
)

// PrintUpdateTable renders available updates in a styled table.
func PrintUpdateTable(updates []skill.UpdateInfo) {
	if len(updates) == 0 {
		return
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(Yellow)
	fmt.Println()
	fmt.Println(titleStyle.Render("Updates available:"))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "SKILL\tINSTALLED\t\tAVAILABLE\tCLIENTS")
	fmt.Fprintln(w, strings.Repeat("-", 20)+"\t"+strings.Repeat("-", 12)+"\t\t"+strings.Repeat("-", 12)+"\t"+strings.Repeat("-", 20))

	arrowStyle := lipgloss.NewStyle().Foreground(Cyan)
	arrow := arrowStyle.Render("->")

	for _, u := range updates {
		clients := strings.Join(u.AffectedClients, ", ")
		installed := u.InstalledVersion
		if installed == "" {
			installed = "unversioned"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", u.SkillName, installed, arrow, u.AvailableVersion, clients)
	}

	w.Flush()

	hintStyle := lipgloss.NewStyle().Foreground(Gray)
	fmt.Println(hintStyle.Render("Run: aisk update"))
}
