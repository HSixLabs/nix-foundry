/*
Package config provides configuration management commands for Nix Foundry.
It implements Cobra commands for managing configuration files, including
initialization, modification, and cleanup of configuration settings across
different scopes (user, team, project).
*/
package config

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/pkg/config"
	"github.com/spf13/cobra"
)

/*
NewUninstallCmd creates a new uninstall command for removing Nix Foundry configuration.
It returns a Cobra command that when executed will remove all configuration files
and directories associated with Nix Foundry, effectively resetting the configuration
state of the system.
*/
func NewUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall Nix Foundry configuration",
		Long: `Uninstall Nix Foundry configuration.
This will remove all configuration files and directories.`,
		RunE: runUninstall,
	}
}

/*
runUninstall executes the uninstallation of Nix Foundry configuration files.
It removes all configuration files and directories using the configuration service.
Returns an error if the uninstallation process fails.
*/
func runUninstall(_ *cobra.Command, _ []string) error {
	configSvc := config.GetConfigService()
	if err := configSvc.UninstallConfig(); err != nil {
		return fmt.Errorf("failed to uninstall configuration: %w", err)
	}

	fmt.Println("✨ Configuration uninstalled successfully!")
	return nil
}
