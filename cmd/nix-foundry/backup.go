package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/backup"
	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
)

// Use the same backup directory path as defined in the backup package
const backupDir = ".config/nix-foundry/backups"

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Manage configuration backups",
	Long: `Manage nix-foundry configuration backups.

This command allows you to create, list, and restore configuration backups.`,
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new backup",
	Long: `Create a backup of your current nix-foundry configuration.

Backups are stored in ~/.config/nix-foundry/backups/ with timestamps.

Example:
  nix-foundry backup create`,
	RunE: func(cmd *cobra.Command, args []string) error {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configDir := filepath.Join(home, ".config", "nix-foundry")

		spin := progress.NewSpinner("Creating backup...")
		spin.Start()
		backupFile, err := backup.Create(configDir)
		if err != nil {
			spin.Fail("Backup failed")
			return fmt.Errorf("failed to create backup: %w", err)
		}
		spin.Success(fmt.Sprintf("Backup created at %s", backupFile))
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available backups",
	Long:  `List all available nix-foundry configuration backups.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		backups, err := backup.ListBackups()
		if err != nil {
			return fmt.Errorf("failed to list backups: %w", err)
		}

		if len(backups) == 0 {
			fmt.Println("No backups found")
			return nil
		}

		fmt.Println("Available backups:")
		for i, backup := range backups {
			fmt.Printf("%d. %s\n", i+1, filepath.Base(backup))
		}
		return nil
	},
}

var restoreCmd = &cobra.Command{
	Use:   "restore [backup-file]",
	Short: "Restore configuration from backup",
	Long: `Restore your nix-foundry configuration from a backup file.

The backup file can be specified by name or full path. If only a name is provided,
it will look in the default backup directory (~/.config/nix-foundry/backups/).

Examples:
  # Restore from specific backup
  nix-foundry backup restore config-2024-03-15.tar.gz

  # Restore using full path
  nix-foundry backup restore /path/to/backup.tar.gz

Note: Run 'nix-foundry apply' after restoring to apply the configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("please specify the backup file to restore from")
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		// If the backup file doesn't exist as specified, try looking in the default backup directory
		backupFile := args[0]
		if _, err := os.Stat(backupFile); os.IsNotExist(err) {
			defaultBackupPath := filepath.Join(home, backupDir, filepath.Base(backupFile))
			if _, err := os.Stat(defaultBackupPath); err == nil {
				backupFile = defaultBackupPath
			}
		}

		configDir := filepath.Join(home, ".config", "nix-foundry")
		spin := progress.NewSpinner("Restoring from backup...")
		spin.Start()
		if err := backup.Restore(backupFile, configDir); err != nil {
			spin.Fail("Restore failed")
			return fmt.Errorf("failed to restore backup: %w", err)
		}
		spin.Success("Configuration restored")

		fmt.Println("\nℹ️  Run 'nix-foundry apply' to apply the restored configuration")
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [backup-file]",
	Short: "Delete a backup file",
	Long: `Delete a specific backup file.

The backup file can be specified by name or full path. If only a name is provided,
it will look in the default backup directory (~/.config/nix-foundry/backups/).

Examples:
  # Delete specific backup
  nix-foundry backup delete config-2024-03-15.tar.gz

  # Delete using full path
  nix-foundry backup delete /path/to/backup.tar.gz`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("please specify the backup file to delete")
		}

		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		// If the backup file doesn't exist as specified, try looking in the default backup directory
		backupFile := args[0]
		if _, err := os.Stat(backupFile); os.IsNotExist(err) {
			defaultBackupPath := filepath.Join(home, backupDir, filepath.Base(backupFile))
			if _, err := os.Stat(defaultBackupPath); err == nil {
				backupFile = defaultBackupPath
			} else {
				return fmt.Errorf("backup file not found: %s", args[0])
			}
		}

		spin := progress.NewSpinner("Deleting backup...")
		spin.Start()
		if err := os.Remove(backupFile); err != nil {
			spin.Fail("Failed to delete backup")
			return fmt.Errorf("failed to delete backup: %w", err)
		}
		spin.Success("Backup deleted")
		return nil
	},
}

func init() {
	// Add subcommands to backupCmd
	backupCmd.AddCommand(createCmd)
	backupCmd.AddCommand(listCmd)
	backupCmd.AddCommand(restoreCmd)
	backupCmd.AddCommand(deleteCmd)

	// Set up flags
	backupCmd.Flags().BoolVarP(&forceBackup, "force", "f", false, "Skip confirmation prompt")
}
