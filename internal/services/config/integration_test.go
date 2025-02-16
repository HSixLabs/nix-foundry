package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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
	configPath := filepath.Join(configDir, "config.yaml")

	t.Run("full configuration lifecycle", func(t *testing.T) {
		// Initialize service
		service := &ServiceImpl{
			path:   configPath,
			config: NewDefaultConfig(),
		}

		// Test initial save
		if err := service.Save(); err != nil {
			t.Fatalf("Initial Save() failed: %v", err)
		}

		// Modify configuration
		if err := service.SetValue("backup.maxBackups", "15"); err != nil {
			t.Fatalf("SetValue() failed: %v", err)
		}

		// Save modifications
		if err := service.Save(); err != nil {
			t.Fatalf("Save() after modification failed: %v", err)
		}

		// Create new service instance to test loading
		newService := &ServiceImpl{
			path: configPath,
		}

		// Load and verify
		if err := newService.Load(); err != nil {
			t.Fatalf("Load() failed: %v", err)
		}

		value, err := newService.GetValue("backup.maxBackups")
		if err != nil {
			t.Fatalf("GetValue() failed: %v", err)
		}
		if value != 15 {
			t.Errorf("Expected maxBackups=15, got %v", value)
		}
	})

	t.Run("environment switching", func(t *testing.T) {
		service := &ServiceImpl{
			path: configPath,
			config: &Config{
				LastUpdated: time.Now(),
				Version:     "1.0",
				Environment: EnvironmentSettings{
					Default:    "development",
					AutoSwitch: true,
				},
			},
		}

		// Apply environment settings
		if err := service.Apply(&NixConfig{
			Shell:  struct{ Type string }{Type: "bash"},
			Editor: struct{ Type string }{Type: "vim"},
		}, false); err != nil {
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
		service := &ServiceImpl{
			path:   configPath,
			config: NewDefaultConfig(),
		}

		// Create multiple backups
		for i := 0; i < 3; i++ {
			if err := service.Save(); err != nil {
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
