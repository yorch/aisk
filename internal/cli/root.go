package cli

import (
	"github.com/spf13/cobra"
	"github.com/yorch/aisk/internal/config"
)

var rootCmd = &cobra.Command{
	Use:   "aisk",
	Short: "AI Skill Manager — install coding skills across AI clients",
	Long: `aisk manages AI coding assistant skills across multiple clients
(Claude Code, Gemini CLI, Codex CLI, VS Code Copilot, Cursor, Windsurf).

Each client gets skills in its native format via dedicated adapters.`,
	Version: config.AppVersion,
}

var assumeYes bool

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&assumeYes, "yes", "y", false, "disable interactive prompts; require explicit inputs")
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(planCmd)
	rootCmd.AddCommand(clientsCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(lintCmd)
	rootCmd.AddCommand(auditCmd)
	rootCmd.AddCommand(completionCmd)
}
