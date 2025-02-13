package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/config"
	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply the current configuration",
	Long: `Apply the current nix-foundry configuration to your system.

This command will:
1. Read your configuration from ~/.config/nix-foundry/
2. Generate necessary Nix and home-manager files
3. Apply the configuration using home-manager

Example:
  nix-foundry apply

Note: Run this command after making changes to your configuration or after restoring from a backup.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get config directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configDir := filepath.Join(home, ".config", "nix-foundry")

		// Check if config exists
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			return fmt.Errorf("no configuration found at %s. Please run './nix-foundry init' first", configDir)
		}

		// Apply configuration
		spin := progress.NewSpinner("Applying configuration...")
		spin.Start()
		if err := config.Apply(configDir); err != nil {
			spin.Fail("Failed to apply configuration")
			return fmt.Errorf("failed to apply configuration: %w", err)
		}
		spin.Success("Configuration applied")

		fmt.Println("\n✨ Environment updated successfully!")
		return nil
	},
}
