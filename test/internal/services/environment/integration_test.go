package environment_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnvironmentIntegration tests the environment service integration
func TestEnvironmentIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Setup test environment
	homeDir := t.TempDir()
	configDir := filepath.Join(homeDir, ".nix-foundry")

	// Create test services
	cfgSvc := config.NewService()
	platformSvc := platform.NewService()

	t.Run("full environment initialization", func(t *testing.T) {
		svc := environment.NewService(
			configDir,
			cfgSvc,
			platformSvc,
		)

		// Test initialization
		err := svc.Initialize(true)
		require.NoError(t, err, "Environment initialization should succeed")

		// Verify directory structure
		dirs := []string{
			configDir,
			filepath.Join(configDir, "environments"),
			filepath.Join(configDir, "environments", "default"),
			filepath.Join(configDir, "backups"),
			filepath.Join(configDir, "logs"),
		}

		for _, dir := range dirs {
			exists, err := dirExists(dir)
			assert.NoError(t, err, "Directory check should not error")
			assert.True(t, exists, "Directory %s should exist", dir)
		}

		// Verify symlink
		currentEnv := filepath.Join(configDir, "environments", "current")
		fi, err := os.Lstat(currentEnv)
		require.NoError(t, err, "Should be able to stat current environment symlink")
		assert.True(t, fi.Mode()&os.ModeSymlink != 0, "Current environment should be a symlink")

		target, err := os.Readlink(currentEnv)
		require.NoError(t, err, "Should be able to read symlink target")
		expectedTarget := filepath.Join(configDir, "environments", "default")
		assert.Equal(t, expectedTarget, target, "Symlink should point to default environment")
	})

	t.Run("prerequisite checks", func(t *testing.T) {
		svc := environment.NewService(
			configDir,
			cfgSvc,
			platformSvc,
		)

		// Test with test mode enabled
		err := svc.CheckPrerequisites(true)
		assert.NoError(t, err, "Prerequisite check should pass in test mode")

		// Test with test mode disabled
		err = svc.CheckPrerequisites(false)
		if err != nil {
			t.Logf("Prerequisites check failed as expected in non-test mode: %v", err)
		} else {
			t.Log("Prerequisites check passed (nix and home-manager are installed)")
		}
	})
}

// Helper function to check if directory exists
func dirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}
