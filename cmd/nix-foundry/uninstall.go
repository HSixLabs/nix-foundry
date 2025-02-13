package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
)

func init() {
	uninstallCmd.Flags().BoolVarP(&keepConfig, "keep-config", "k", false, "Keep configuration files")
	uninstallCmd.Flags().BoolVarP(&forceUninstall, "force", "f", false, "Skip confirmation prompt")
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove nix-foundry configuration and clean up",
	Long: `Completely remove nix-foundry configuration and clean up all created files.

This will:
1. Remove the nix-foundry configuration directory (~/.config/nix-foundry)
2. Remove any created backups
3. Clean up nix profile entries
4. Remove generated home-manager configuration

Note: This will NOT uninstall the Nix package manager itself.
For complete Nix removal, follow the official uninstall instructions:
https://nixos.org/manual/nix/stable/installation/uninstall.html

Example:
  nix-foundry uninstall`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !forceUninstall {
			fmt.Print("Are you sure you want to uninstall nix-foundry? [y/N]: ")
			var response string
			if _, err := fmt.Scanln(&response); err != nil {
				return fmt.Errorf("failed to read user input: %w", err)
			}
			if !strings.EqualFold(response, "y") && !strings.EqualFold(response, "yes") {
				fmt.Println("Uninstall cancelled")
				return nil
			}
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		// Paths to clean up
		configDir := filepath.Join(home, ".config", "nix-foundry")
		backupDir := filepath.Join(home, ".config", "nix-foundry", "backups")

		// Check if config exists
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			return fmt.Errorf("no configuration found at %s. Nothing to uninstall", configDir)
		}

		// Show warning and confirmation prompt
		fmt.Println("\n⚠️  Warning: This action cannot be undone!")
		fmt.Println("\nThe following changes will be made:")
		fmt.Printf("1. Remove configuration directory: %s\n", configDir)
		fmt.Printf("2. Remove backup directory: %s\n", backupDir)
		fmt.Println("3. Remove home-manager generations created by nix-foundry")
		fmt.Println("4. Clean up nix profile entries")
		fmt.Println("\nNote: This will not uninstall the Nix package manager itself.")

		// Remove nix profile entries
		spin := progress.NewSpinner("Cleaning up nix profiles...")
		spin.Start()

		// Remove home-manager generation
		hmCmd := exec.Command("home-manager", "generations")
		if err := hmCmd.Run(); err == nil { // Only if home-manager is installed
			removeCmd := exec.Command("home-manager", "remove-generations", "all")
			if err := removeCmd.Run(); err != nil {
				spin.Stop()
				fmt.Println("⚠️  Could not remove home-manager generations")
			}
		}

		// Remove nix profile entries
		nixCmd := exec.Command("nix", "profile", "remove", "all")
		if err := nixCmd.Run(); err != nil {
			spin.Stop()
			fmt.Println("⚠️  Could not remove nix profile entries")
		}
		spin.Success("Cleaned up nix profiles")

		// Remove configuration directories
		spin = progress.NewSpinner("Removing configuration files...")
		spin.Start()

		if err := os.RemoveAll(backupDir); err != nil && !os.IsNotExist(err) {
			spin.Stop()
			fmt.Printf("⚠️  Could not remove backup directory: %v\n", err)
		}

		if err := os.RemoveAll(configDir); err != nil && !os.IsNotExist(err) {
			spin.Fail("Failed to remove configuration directory")
			return fmt.Errorf("failed to remove configuration directory: %w", err)
		}
		spin.Success("Removed configuration files")

		fmt.Println("\n✨ nix-foundry has been uninstalled successfully!")
		fmt.Println("\nNote: The nix package manager itself was not removed.")
		fmt.Println("To remove nix completely, please follow the official uninstall instructions:")
		fmt.Println("https://nixos.org/manual/nix/stable/installation/uninstall.html")

		return nil
	},
}
