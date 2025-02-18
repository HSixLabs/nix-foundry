package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
)

func TestConfigIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	homeDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Setenv("HOME", homeDir)
	defer func() {
		os.Setenv("HOME", originalHome)
	}()

	configDir := filepath.Join(homeDir, ".config", "nix-foundry")
	os.Setenv("NIX_FOUNDRY_CONFIG_DIR", configDir)
	defer os.Unsetenv("NIX_FOUNDRY_CONFIG_DIR")

	t.Run("full configuration lifecycle", func(t *testing.T) {
		// Initialize service
		service := config.NewService()
		if err := service.Initialize(true); err != nil {
			t.Fatalf("Failed to initialize service: %v", err)
		}

		// Test initial save
		if err := service.Save(testConfig); err != nil {
			t.Fatalf("Initial Save() failed: %v", err)
		}

		// Modify configuration
		if err := service.SetValue("backup.maxBackups", "15"); err != nil {
			t.Fatalf("SetValue() failed: %v", err)
		}

		// Save modifications
		if err := service.Save(testConfig); err != nil {
			t.Fatalf("Save() after modification failed: %v", err)
		}

		// Create new service instance to test loading
		newService := config.NewService()

		// Load and verify
		_, err := newService.Load()
		if err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		value, err := newService.GetValue("backup.maxBackups")
		if err != nil {
			t.Fatalf("GetValue() failed: %v", err)
		}
		if value != "15" {
			t.Errorf("Expected maxBackups=15, got %v", value)
		}
	})

	t.Run("environment switching", func(t *testing.T) {
		service := config.NewService()
		if err := service.Initialize(true); err != nil {
			t.Fatalf("Failed to initialize service: %v", err)
		}

		nixConfig := &config.NixConfig{
			Shell: config.ShellConfig{
				Type:     "bash",
				InitFile: "~/.bashrc",
			},
			Editor: config.EditorConfig{
				Type:        "vim",
				ConfigPath:  "~/.vimrc",
				PackageName: "vim",
			},
		}

		// Apply environment settings
		if err := service.Apply(nixConfig, false); err != nil {
			t.Fatalf("Apply() failed: %v", err)
		}

		// Verify environment symlink
		currentLink := filepath.Join(configDir, "environments", "current")
		target, err := os.Readlink(currentLink)
		if err != nil {
			t.Fatalf("Failed to read symlink: %v", err)
		}

		expectedTarget := filepath.Join(configDir, "environments", "development")
		if target != expectedTarget {
			t.Errorf("Expected symlink target %s, got %s", expectedTarget, target)
		}
	})

	t.Run("backup creation and rotation", func(t *testing.T) {
		service := config.NewService()
		if err := service.Initialize(true); err != nil {
			t.Fatalf("Failed to initialize service: %v", err)
		}

		// Create multiple backups
		for i := 0; i < 3; i++ {
			if err := service.Save(testConfig); err != nil {
				t.Fatalf("Save() failed: %v", err)
			}
			time.Sleep(time.Millisecond * 100) // Ensure unique timestamps
		}

		// Verify backup directory
		backupDir := filepath.Join(configDir, "backups")
		entries, err := os.ReadDir(backupDir)
		if err != nil {
			t.Fatalf("Failed to read backup directory: %v", err)
		}

		if len(entries) == 0 {
			t.Error("No backups created")
		}
	})
}
