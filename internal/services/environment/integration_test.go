package environment

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/validation"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
)

func TestEnvironmentIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	homeDir := t.TempDir()
	configDir := filepath.Join(homeDir, ".nix-foundry")

	t.Run("full environment initialization", func(t *testing.T) {
		service := NewService(
			configDir,
			config.NewService(),
			validation.NewService(),
			platform.NewService(),
		)

		// Test initialization
		if err := service.Initialize(true); err != nil {
			t.Fatalf("Environment initialization failed: %v", err)
		}

		// Verify directory structure
		dirs := []string{
			configDir,
			filepath.Join(configDir, "environments"),
			filepath.Join(configDir, "environments", "default"),
			filepath.Join(configDir, "backups"),
			filepath.Join(configDir, "logs"),
		}

		for _, dir := range dirs {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				t.Errorf("Expected directory %s to exist", dir)
			}
		}

		// Verify symlink
		currentEnv := filepath.Join(configDir, "environments", "current")
		fi, err := os.Lstat(currentEnv)
		if err != nil {
			t.Fatalf("Failed to stat current environment symlink: %v", err)
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			t.Error("Expected current environment to be a symlink")
		}

		target, err := os.Readlink(currentEnv)
		if err != nil {
			t.Fatalf("Failed to read symlink target: %v", err)
		}
		expectedTarget := filepath.Join(configDir, "environments", "default")
		if target != expectedTarget {
			t.Errorf("Symlink points to %s, expected %s", target, expectedTarget)
		}
	})

	t.Run("prerequisite checks", func(t *testing.T) {
		service := NewService(
			configDir,
			config.NewService(),
			validation.NewService(),
			platform.NewService(),
		)

		// Test with test mode enabled (should skip checks)
		if err := service.CheckPrerequisites(true); err != nil {
			t.Errorf("Expected prerequisite check to pass in test mode: %v", err)
		}

		// Test with test mode disabled (should perform actual checks)
		err := service.CheckPrerequisites(false)
		if err == nil {
			t.Log("Prerequisites check passed (nix and home-manager are installed)")
		} else {
			t.Log("Prerequisites check failed as expected (nix or home-manager not installed)")
		}
	})
}
