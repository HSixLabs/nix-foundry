package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
)

func NewApplyCommand() *cobra.Command {
	var testMode bool

	cmd := &cobra.Command{
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
				return fmt.Errorf("no configuration found at %s. Please run 'nix-foundry init' first", configDir)
			}

			service := config.NewService()

			// Apply configuration
			spin := progress.NewSpinner("Applying configuration...")
			spin.Start()
			defer spin.Stop()

			if err := service.Load(); err != nil {
				spin.Fail("Failed to load configuration")
				return fmt.Errorf("failed to load configuration: %w", err)
			}
			config := service.GetConfig()

			if err := service.Apply(config, testMode); err != nil {
				spin.Fail("Failed to apply configuration")
				return fmt.Errorf("failed to apply configuration: %w", err)
			}
			spin.Success("Configuration applied")

			fmt.Println("\n✨ Environment updated successfully!")
			return nil
		},
	}

	cmd.Flags().BoolVar(&testMode, "test", false, "Run in test mode")
	return cmd
}

func NewConfigCommandGroup(cfgSvc config.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage application configuration",
	}

	cmd.AddCommand(
		newGetCmd(cfgSvc),
		newSetCmd(cfgSvc),
		newResetCmd(cfgSvc),
		newInitCommand(cfgSvc),
	)

	return cmd
}
