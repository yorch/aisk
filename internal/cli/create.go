package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yorch/aisk/internal/audit"
	"github.com/yorch/aisk/internal/config"
	"github.com/yorch/aisk/internal/skill"
)

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Scaffold a new skill directory with template files",
	Args:  cobra.ExactArgs(1),
	RunE:  runCreate,
}

var createPath string

func init() {
	createCmd.Flags().StringVar(&createPath, "path", "", "parent directory for the new skill (default: skills repo path)")
}

func runCreate(_ *cobra.Command, args []string) (retErr error) {
	name := args[0]

	parentDir := createPath
	if parentDir == "" {
		paths, err := config.ResolvePaths()
		if err != nil {
			return err
		}
		al := audit.New(paths.AiskDir, "create")
		al.Log("command.create", "started", map[string]any{
			"name": name,
			"path": parentDir,
		}, nil)
		defer func() {
			status := "success"
			if retErr != nil {
				status = "error"
			}
			al.Log("command.create", status, nil, retErr)
		}()
		parentDir = paths.SkillsRepo
		al.Log("create.parent.resolve", "success", map[string]any{"path": parentDir}, nil)
		return runCreateWithAudit(name, parentDir, al)
	}

	// createPath explicitly set by user; we still write audit logs to default app dir.
	paths, err := config.ResolvePaths()
	if err != nil {
		return err
	}
	al := audit.New(paths.AiskDir, "create")
	al.Log("command.create", "started", map[string]any{
		"name": name,
		"path": parentDir,
	}, nil)
	defer func() {
		status := "success"
		if retErr != nil {
			status = "error"
		}
		al.Log("command.create", status, nil, retErr)
	}()
	al.Log("create.parent.resolve", "success", map[string]any{"path": parentDir}, nil)
	return runCreateWithAudit(name, parentDir, al)
}

func runCreateWithAudit(name, parentDir string, al *audit.Logger) error {
	skillDir, err := skill.Scaffold(parentDir, name)
	if err != nil {
		al.Log("create.scaffold", "error", map[string]any{"name": name, "path": parentDir}, err)
		return err
	}
	al.Log("create.scaffold", "success", map[string]any{"name": name, "path": skillDir}, nil)

	fmt.Printf("Created skill %q at %s\n", name, skillDir)
	fmt.Println("Next steps:")
	fmt.Println("  1. Edit SKILL.md — fill in description, version, and instructions")
	fmt.Println("  2. Add reference files to reference/")
	fmt.Println("  3. Add example files to examples/")
	fmt.Printf("  4. Run: aisk lint %s\n", skillDir)
	return nil
}
