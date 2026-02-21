package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/yorch/aisk/internal/audit"
	"github.com/yorch/aisk/internal/config"
	"github.com/yorch/aisk/internal/skill"
	"github.com/yorch/aisk/internal/tui"
)

var lintCmd = &cobra.Command{
	Use:   "lint [path]",
	Short: "Validate a skill directory or SKILL.md file",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runLint,
}

func runLint(_ *cobra.Command, args []string) error {
	paths, err := config.ResolvePaths()
	if err != nil {
		return err
	}
	al := audit.New(paths.AiskDir, "lint")
	target := "."
	if len(args) > 0 {
		target = args[0]
	}
	al.Log("command.lint", "started", map[string]any{"target": target}, nil)

	info, err := os.Stat(target)
	if err != nil {
		al.Log("command.lint", "error", map[string]any{"target": target}, err)
		return fmt.Errorf("cannot access %s: %w", target, err)
	}

	var report *skill.LintReport

	if info.IsDir() {
		report, err = skill.LintSkillDir(target)
		if err != nil {
			al.Log("lint.run", "error", map[string]any{"target": target, "kind": "directory"}, err)
			return err
		}
	} else {
		// It's a file — lint as SKILL.md content
		data, err := os.ReadFile(target)
		if err != nil {
			al.Log("lint.run", "error", map[string]any{"target": target, "kind": "file"}, err)
			return fmt.Errorf("reading %s: %w", target, err)
		}
		report = skill.LintSkillMD(string(data))
	}

	if len(report.Results) == 0 {
		fmt.Println(lipgloss.NewStyle().Foreground(tui.Green).Render("No issues found."))
		al.Log("command.lint", "success", map[string]any{"target": target, "errors": 0, "warnings": 0}, nil)
		return nil
	}

	errStyle := lipgloss.NewStyle().Foreground(tui.Red)
	warnStyle := lipgloss.NewStyle().Foreground(tui.Yellow)

	for _, r := range report.Results {
		prefix := warnStyle.Render("warning")
		if r.Severity == skill.SeverityError {
			prefix = errStyle.Render("error")
		}
		field := r.Field
		if field != "" {
			field = "[" + field + "] "
		}
		fmt.Printf("  %s: %s%s\n", prefix, field, r.Message)
	}

	errs := report.Errors()
	warns := report.Warnings()
	var parts []string
	if len(errs) > 0 {
		parts = append(parts, fmt.Sprintf("%d error(s)", len(errs)))
	}
	if len(warns) > 0 {
		parts = append(parts, fmt.Sprintf("%d warning(s)", len(warns)))
	}
	fmt.Printf("\n%s\n", strings.Join(parts, ", "))

	if report.HasErrors() {
		al.Log("command.lint", "error", map[string]any{
			"target":   target,
			"errors":   len(errs),
			"warnings": len(warns),
		}, fmt.Errorf("lint failed"))
		os.Exit(1)
	}
	al.Log("command.lint", "success", map[string]any{
		"target":   target,
		"errors":   len(errs),
		"warnings": len(warns),
	}, nil)
	return nil
}
