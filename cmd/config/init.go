// Package config provides configuration management commands for Nix Foundry.
package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewInitCmd creates a new init command.
func NewInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize Nix Foundry configuration",
		Long: `Initialize Nix Foundry configuration.
This will create a new configuration file with default settings.`,
		RunE: runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	configSvc := getConfigService()
	if err := configSvc.InitConfig(); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	fmt.Println("✨ Configuration initialized successfully!")
	return nil
}
