package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch [environment]",
	Short: "Switch between environments",
	Long: `Switch between personal and team environments.
Examples:
  nix-foundry switch personal
  nix-foundry switch project`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]

		// Validate environment name
		if target != "personal" && target != "project" && target != "work" {
			return fmt.Errorf("invalid environment '%s'. Must be 'personal', 'project', or 'work'", target)
		}

		// If not forcing, check for conflicts
		if !forceSwitch {
			if err := checkConfigConflicts(); err != nil {
				return fmt.Errorf("configuration conflicts detected: %w\nUse --force to override", err)
			}
		}

		// Switch environment using existing function
		if err := switchEnvironment(target); err != nil {
			// If switch fails, attempt rollback
			fmt.Println("Switch failed, attempting rollback...")
			if rbErr := rollbackEnvironment(); rbErr != nil {
				return fmt.Errorf("switch failed and rollback failed: %v (rollback error: %v)", err, rbErr)
			}
			return fmt.Errorf("switch failed, rolled back to previous state: %w", err)
		}

		return nil
	},
}

func init() {
	switchCmd.Flags().BoolVarP(&forceSwitch, "force", "f", false, "Force switch even if conflicts exist")
}

func rollbackEnvironment() error {
	configDir := getConfigDir()
	backupDir := filepath.Join(configDir, "backups")

	// Find latest backup
	entries, readErr := os.ReadDir(backupDir)
	if readErr != nil {
		return fmt.Errorf("no backups found for rollback: %w", readErr)
	}

	var latestTime time.Time
	var latestBackup string
	for _, entry := range entries {
		// Look for .tar.gz files instead of directories
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tar.gz") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latestBackup = filepath.Join(backupDir, entry.Name())
		}
	}

	if latestBackup == "" {
		return fmt.Errorf("no valid backups found for rollback")
	}

	// Create a temporary directory for extraction
	tempDir, tempErr := os.MkdirTemp("", "nix-foundry-rollback-*")
	if tempErr != nil {
		return fmt.Errorf("failed to create temp directory: %w", tempErr)
	}
	defer os.RemoveAll(tempDir)

	// Extract the backup
	cmd := exec.Command("tar", "-xzf", latestBackup, "-C", tempDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract backup: %w", err)
	}

	currentEnv := filepath.Join(configDir, "environments", "current")

	// Remove existing symlink if it exists
	if _, err := os.Lstat(currentEnv); err == nil {
		if err := os.Remove(currentEnv); err != nil {
			return fmt.Errorf("failed to remove existing symlink: %w", err)
		}
	}

	// Get the target environment from the backup
	envFile := filepath.Join(tempDir, "current_env")
	target, err := os.ReadFile(envFile)
	if err != nil {
		return fmt.Errorf("failed to read environment from backup: %w", err)
	}

	// Create symlink to the appropriate environment
	envDir := filepath.Join(configDir, "environments", string(target))
	if err := os.Symlink(envDir, currentEnv); err != nil {
		return fmt.Errorf("failed to restore environment symlink: %w", err)
	}

	return nil
}

func switchEnvironment(target string) error {
	if target != "personal" && target != "project" && target != "work" {
		return fmt.Errorf("invalid environment '%s': must be 'personal', 'project', or 'work'", target)
	}

	// Create backup before switching
	if err := createBackup(); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	configDir := getConfigDir()
	envDir := filepath.Join(configDir, "environments", target)
	currentEnv := filepath.Join(configDir, "environments", "current")

	// Ensure target environment directory exists
	if err := os.MkdirAll(envDir, 0755); err != nil {
		return fmt.Errorf("failed to create environment directory: %w", err)
	}

	// Remove only the symlink, not the directory it points to
	if _, err := os.Lstat(currentEnv); err == nil {
		if err := os.Remove(currentEnv); err != nil {
			return fmt.Errorf("failed to remove existing symlink: %w", err)
		}
	}

	// Create symlink to new environment
	if err := os.Symlink(envDir, currentEnv); err != nil {
		return fmt.Errorf("failed to switch to %s environment: %w", target, err)
	}

	// Update current environment file
	envFile := filepath.Join(configDir, "current_env")
	if err := os.WriteFile(envFile, []byte(target), 0644); err != nil {
		return fmt.Errorf("failed to update environment file: %w", err)
	}

	fmt.Printf("✅ Successfully switched to %s environment\n", target)
	return nil
}
