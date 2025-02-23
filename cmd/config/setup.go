// Package config provides configuration management commands for Nix Foundry.
package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewSetupCmd creates a new setup command.
func NewSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Setup Nix Foundry configuration",
		Long: `Setup Nix Foundry configuration.
This will create a new configuration file with default settings.`,
		RunE: runSetup,
	}
}

func runSetup(cmd *cobra.Command, args []string) error {
	configSvc := getConfigService()
	if err := configSvc.InitConfig(); err != nil {
		return fmt.Errorf("failed to setup configuration: %w", err)
	}

	fmt.Println("✨ Configuration setup successfully!")
	return nil
}
