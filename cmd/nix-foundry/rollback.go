package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var (
	forceRollback bool
	list          bool
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback [backup-id]",
	Short: "Revert to last working state",
	Long: `Revert to the last known working state of your environment.
This uses automatic backups created during environment switches and updates.

Examples:
  nix-foundry rollback          # Roll back to most recent backup
  nix-foundry rollback --list   # List available backups
  nix-foundry rollback abc123   # Roll back to specific backup`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if list {
			return listBackups()
		}
		return performRollback(args)
	},
}

func init() {
	rollbackCmd.Flags().BoolVarP(&forceRollback, "force", "f", false, "Skip confirmation prompt")
	rollbackCmd.Flags().BoolVarP(&list, "list", "l", false, "List available backups")
}

func listBackups() error {
	backupDir := filepath.Join(getConfigDir(), "backups")
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("no backups found: %w", err)
	}

	fmt.Println("\n📦 Available Backups")
	fmt.Println("===================")

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		fmt.Printf("ID: %s\n  Created: %s\n  Size: %d bytes\n\n",
			entry.Name(),
			info.ModTime().Format("2006-01-02 15:04:05"),
			info.Size())
	}
	return nil
}

func performRollback(args []string) error {
	configDir := getConfigDir()
	backupDir := filepath.Join(configDir, "backups")

	var backupPath string
	if len(args) > 0 {
		// Specific backup requested
		backupPath = filepath.Join(backupDir, args[0])
		if _, err := os.Stat(backupPath); err != nil {
			return fmt.Errorf("backup %s not found", args[0])
		}
	} else {
		// Find latest backup
		entries, err := os.ReadDir(backupDir)
		if err != nil {
			return fmt.Errorf("no backups found: %w", err)
		}

		var latestTime time.Time
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if info.ModTime().After(latestTime) {
				latestTime = info.ModTime()
				backupPath = filepath.Join(backupDir, entry.Name())
			}
		}
	}

	if backupPath == "" {
		return fmt.Errorf("no valid backups found")
	}

	// Confirm unless forced
	if !forceRollback {
		fmt.Printf("Rolling back to backup: %s\n", filepath.Base(backupPath))
		fmt.Print("Continue? [y/N]: ")
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			return fmt.Errorf("failed to read user input: %w", err)
		}
		if response != "y" && response != "Y" {
			fmt.Println("Rollback cancelled")
			return nil
		}
	}

	currentEnv := filepath.Join(configDir, "environments", "current")

	// Create backup of current state first
	if err := createBackup(); err != nil {
		return fmt.Errorf("failed to backup current state: %w", err)
	}

	// Restore from backup
	if err := os.RemoveAll(currentEnv); err != nil {
		return fmt.Errorf("failed to clean current environment: %w", err)
	}

	if err := os.Rename(backupPath, currentEnv); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	fmt.Println("✅ Successfully rolled back to previous state")
	return nil
}
