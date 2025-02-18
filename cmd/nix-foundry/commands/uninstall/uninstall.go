package uninstall

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/internal/flags"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	"github.com/shawnkhoffman/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
)

func NewUninstallCommand() *cobra.Command {
	var (
		keepBackups bool
		autoConfirm bool
	)
	logger := logging.GetLogger()

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove nix-foundry and all associated data",
		Long: `Uninstall nix-foundry and clean up all associated data.
This will remove:
- All nix-foundry environments
- Project configurations
- Cached packages
- Backup files (unless --keep-backups is specified)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get config directory
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			configDir := filepath.Join(home, ".config", "nix-foundry")

			// Show warning and confirmation
			fmt.Println("🚨 This will permanently remove:")
			fmt.Println(" - All nix-foundry environments")
			fmt.Println(" - Project configurations")
			fmt.Println(" - Cached packages")
			if !keepBackups {
				fmt.Println(" - All backups")
			}

			// Get confirmation unless auto-confirmed
			if !autoConfirm {
				fmt.Print("\nContinue? [y/N]: ")
				var response string
				if _, err := fmt.Scanln(&response); err != nil {
					response = "n"
				}
				if response != "y" && response != "Y" {
					fmt.Println("Uninstall cancelled")
					return nil
				}
			}

			// Start progress spinner
			spin := progress.NewSpinner("Uninstalling nix-foundry...")
			spin.Start()
			defer spin.Stop()

			// Remove directories in a specific order
			dirsToRemove := []string{
				filepath.Join(configDir, "environments"),
				filepath.Join(configDir, "projects"),
				filepath.Join(configDir, "cache"),
			}

			if !keepBackups {
				backupDir := filepath.Join(configDir, "backups")
				if _, err := os.Stat(backupDir); err == nil {
					dirsToRemove = append(dirsToRemove, backupDir)
				}
			}

			// Remove each directory with logging
			for _, dir := range dirsToRemove {
				logger.Debug("Removing directory", "path", dir)
				if err := os.RemoveAll(dir); err != nil {
					logger.Error("Failed to remove directory", "path", dir, "error", err)
					spin.Fail("Uninstall failed")
					return fmt.Errorf("failed to remove %s: %w", dir, err)
				}
			}

			// Finally remove the config directory itself
			logger.Debug("Removing config directory", "path", configDir)
			if err := os.RemoveAll(configDir); err != nil {
				logger.Error("Failed to remove config directory", "error", err)
				spin.Fail("Uninstall failed")
				return fmt.Errorf("failed to remove config directory: %w", err)
			}

			spin.Success("Uninstall complete")
			fmt.Println("\n✅ nix-foundry successfully uninstalled")
			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&keepBackups, "keep-backups", false, "Preserve backup history during uninstall")
	flags.AddYesFlag(cmd) // This will set up the --yes/-y flag and bind it to autoConfirm

	return cmd
}
