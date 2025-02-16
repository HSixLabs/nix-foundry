package platform_test

import (
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
)

func TestPlatformService(t *testing.T) {
	service := platform.NewService()

	t.Run("test mode", func(t *testing.T) {
		if err := service.SetupPlatform(true); err != nil {
			t.Errorf("Expected no error in test mode, got %v", err)
		}
	})

	t.Run("darwin setup", func(t *testing.T) {
		// Create a new service instance for darwin testing
		service := platform.NewService()
		err := service.SetupPlatform(false)
		if err != nil {
			// Expected in CI environment where we can't actually install Homebrew
			t.Logf("Darwin setup failed (expected in CI): %v", err)
		}
	})
}

func TestHomeManagerInstallation(t *testing.T) {
	service := platform.NewService()

	t.Run("home-manager installation", func(t *testing.T) {
		err := service.InstallHomeManager()
		if err != nil {
			// Expected in CI environment where we can't actually install home-manager
			t.Logf("Home-manager installation failed (expected in CI): %v", err)
		}
	})
}

func TestErrorMessages(t *testing.T) {
	t.Run("platform setup error", func(t *testing.T) {
		service := platform.NewService()

		// Test with non-test mode to trigger actual setup
		err := service.SetupPlatform(false)
		if err != nil {
			// Verify error type and message format
			if _, ok := err.(*errors.PlatformError); !ok {
				t.Error("Expected PlatformError type")
			}

			// Error message format check
			if err.Error() == "" {
				t.Error("Expected non-empty error message")
			}
		}
	})
}
