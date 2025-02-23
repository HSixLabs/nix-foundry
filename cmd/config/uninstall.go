// Package config provides configuration management commands for Nix Foundry.
package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewUninstallCmd creates a new uninstall command.
func NewUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall Nix Foundry configuration",
		Long: `Uninstall Nix Foundry configuration.
This will remove all configuration files and directories.`,
		RunE: runUninstall,
	}
}

func runUninstall(cmd *cobra.Command, args []string) error {
	configSvc := getConfigService()
	if err := configSvc.UninstallConfig(); err != nil {
		return fmt.Errorf("failed to uninstall configuration: %w", err)
	}

	fmt.Println("✨ Configuration uninstalled successfully!")
	return nil
}
