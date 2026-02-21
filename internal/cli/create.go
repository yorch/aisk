package cli

import (
	"fmt"

	"github.com/spf13/cobra"
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

func runCreate(_ *cobra.Command, args []string) error {
	name := args[0]

	parentDir := createPath
	if parentDir == "" {
		paths, err := config.ResolvePaths()
		if err != nil {
			return err
		}
		parentDir = paths.SkillsRepo
	}

	skillDir, err := skill.Scaffold(parentDir, name)
	if err != nil {
		return err
	}

	fmt.Printf("Created skill %q at %s\n", name, skillDir)
	fmt.Println("Next steps:")
	fmt.Println("  1. Edit SKILL.md — fill in description, version, and instructions")
	fmt.Println("  2. Add reference files to reference/")
	fmt.Println("  3. Add example files to examples/")
	fmt.Printf("  4. Run: aisk lint %s\n", skillDir)
	return nil
}
