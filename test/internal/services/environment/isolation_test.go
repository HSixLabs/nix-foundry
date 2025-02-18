package environment_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/test/internal/services/environment/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsolationSetupWithMocks(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Initialize the required services with mocks
	cfgSvc := config.NewService()
	platformSvc := &mocks.MockPlatformService{}

	// Create the environment service
	svc := environment.NewService(
		tempDir,
		cfgSvc,
		platformSvc,
		true,
		true,
		true,
	)

	t.Run("test mode skips isolation", func(t *testing.T) {
		err := svc.SetupIsolation(true, false)
		assert.NoError(t, err, "SetupIsolation should succeed in test mode")

		// Verify that no directories were created in test mode
		_, err = os.Stat(filepath.Join(tempDir, "environments"))
		assert.True(t, os.IsNotExist(err), "No directories should be created in test mode")
	})

	t.Run("creates directory structure", func(t *testing.T) {
		err := svc.SetupIsolation(false, false)
		require.NoError(t, err, "SetupIsolation should succeed")

		// Verify directory structure
		dirs := []string{
			filepath.Join(tempDir, "environments"),
			filepath.Join(tempDir, "environments", "default"),
			filepath.Join(tempDir, "storage"),
			filepath.Join(tempDir, "cache"),
			filepath.Join(tempDir, "backups"),
		}

		for _, dir := range dirs {
			info, err := os.Stat(dir)
			assert.NoError(t, err, "Directory should exist: %s", dir)
			assert.True(t, info.IsDir(), "Path should be a directory: %s", dir)
		}

		// Verify default environment files
		files := []string{
			filepath.Join(tempDir, "environments", "default", "flake.nix"),
			filepath.Join(tempDir, "environments", "default", "home.nix"),
		}

		for _, file := range files {
			info, err := os.Stat(file)
			assert.NoError(t, err, "File should exist: %s", file)
			assert.False(t, info.IsDir(), "Path should be a file: %s", file)
		}
	})
}
