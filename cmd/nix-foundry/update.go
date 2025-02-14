package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/config"
	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
)

func init() {
	updateCmd.Flags().BoolVarP(&forceUpdate, "force", "f", false, "Force update")
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the development environment",
	Long: `Update your nix-foundry development environment to the latest version.
This will preserve your custom configurations while updating the core components.

This command will:
1. Update all Nix packages to their latest versions
2. Preserve your custom configurations
3. Apply the updated environment

Example:
  nix-foundry update

Note: Your existing configuration will be preserved during the update.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get config directory
		homeDir, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return fmt.Errorf("failed to get home directory: %w", homeErr)
		}
		configDir := filepath.Join(homeDir, ".config", "nix-foundry")

		// Check if config exists
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			return fmt.Errorf("no configuration found at %s. Please run 'nix-foundry init' first", configDir)
		}

		// Update flake inputs
		spin := progress.NewSpinner("Updating Nix packages...")
		spin.Start()

		// Convert the path to an absolute path
		absPath, err := filepath.Abs(configDir)
		if err != nil {
			spin.Fail("Failed to resolve config path")
			return fmt.Errorf("failed to resolve config path: %w", err)
		}

		// Use the correct flake reference syntax
		updateCmd := exec.Command("nix", "flake", "update", "--flake", absPath)
		updateCmd.Stdout = os.Stdout
		updateCmd.Stderr = os.Stderr
		if err := updateCmd.Run(); err != nil {
			spin.Fail("Failed to update packages")
			return fmt.Errorf("failed to update flake: %w\nTry running 'nix flake update --flake %s' manually", err, absPath)
		}
		spin.Success("Packages updated")

		// Apply updated configuration
		spin = progress.NewSpinner("Applying configuration...")
		spin.Start()
		if err := config.Apply(configDir, nil, testMode); err != nil {
			spin.Fail("Failed to apply configuration")
			return fmt.Errorf("failed to apply configuration: %w", err)
		}
		spin.Success("Configuration applied")

		fmt.Println("\n✨ Environment updated successfully!")
		return nil
	},
}
