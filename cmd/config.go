// Package cmd provides the command-line interface for Nix Foundry.
package cmd

import (
	"github.com/shawnkhoffman/nix-foundry/cmd/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Nix Foundry configuration",
	Long: `Manage Nix Foundry configuration.
This command provides subcommands for managing your Nix Foundry configuration.`,
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(
		config.NewApplyCmd(),
		config.NewSetCmd(),
		config.NewInitCmd(),
		config.NewUninstallCmd(),
		config.NewScriptCmd(),
		config.NewSetupCmd(),
	)
}
