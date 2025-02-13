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

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show environment status",
	Long: `Display current environment status, active profile, and configuration state.
Includes information about:
- Current environment (personal/project)
- Configuration status
- Last backup time
- Active packages
- Environment health`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return showStatus()
	},
}

func showStatus() error {
	configDir := getConfigDir()

	fmt.Println("🔍 Environment Status")
	fmt.Println("====================")

	// 1. Current Environment
	envFile := filepath.Join(configDir, "current_env")
	envData, readErr := os.ReadFile(envFile)
	if readErr != nil {
		envData = []byte("personal") // Default environment
	}
	fmt.Printf("\n🌍 Current Environment: %s\n", strings.TrimSpace(string(envData)))

	// 2. Configuration Status
	if err := checkConfigConflicts(); err != nil {
		fmt.Printf("⚠️  Configuration Conflicts: %v\n", err)
	} else {
		fmt.Println("✅ Configuration: No conflicts")
	}

	// 3. Last Backup
	backupDir := filepath.Join(configDir, "backups")
	entries, err := os.ReadDir(backupDir)
	if err == nil && len(entries) > 0 {
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
			}
		}
		if !latestTime.IsZero() {
			fmt.Printf("💾 Last Backup: %s\n", latestTime.Format("2006-01-02 15:04:05"))
		}
	}

	// 4. Active Profile
	currentEnv := filepath.Join(configDir, "environments", "current")
	if _, err := os.Stat(currentEnv); err == nil {
		fmt.Println("✅ Environment: Active")
	} else {
		fmt.Println("❌ Environment: Not initialized")
	}

	// 5. System Health
	if err := checkSystemRequirements(); err != nil {
		fmt.Printf("❌ System Health: %v\n", err)
	} else {
		fmt.Println("✅ System Health: OK")
	}

	// 6. Environment Isolation
	if err := checkEnvironmentIsolation(); err != nil {
		fmt.Printf("⚠️  Environment Isolation: %v\n", err)
	} else {
		fmt.Println("✅ Environment Isolation: Active")
	}

	return nil
}

func checkSystemRequirements() error {
	// Check Nix installation
	if _, err := exec.LookPath("nix"); err != nil {
		return fmt.Errorf("nix not found in PATH: %w", err)
	}

	// Check home-manager
	if _, err := exec.LookPath("home-manager"); err != nil {
		return fmt.Errorf("home-manager not found: %w", err)
	}

	// Check required directories
	configDir := getConfigDir()
	requiredDirs := []string{
		filepath.Join(configDir, "environments"),
		filepath.Join(configDir, "backups"),
		filepath.Join(configDir, "logs"),
	}

	for _, dir := range requiredDirs {
		if _, err := os.Stat(dir); err != nil {
			return fmt.Errorf("required directory missing: %s", dir)
		}
	}

	return nil
}

func checkEnvironmentIsolation() error {
	configDir := getConfigDir()

	// Check environment directory structure
	envDir := filepath.Join(configDir, "environments")
	entries, err := os.ReadDir(envDir)
	if err != nil {
		return fmt.Errorf("failed to read environments directory: %w", err)
	}

	// Verify environment separation
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check permissions
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("failed to get environment info: %w", err)
		}

		// Ensure directory is not world-writable
		if info.Mode().Perm()&0002 != 0 {
			return fmt.Errorf("environment %s has unsafe permissions", entry.Name())
		}
	}

	return nil
}
