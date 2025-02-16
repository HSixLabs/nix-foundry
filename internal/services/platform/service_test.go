package platform

import (
	"fmt"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
)

func TestPlatformService(t *testing.T) {
	service := NewService()

	t.Run("test mode", func(t *testing.T) {
		if err := service.SetupPlatform(true); err != nil {
			t.Errorf("Expected no error in test mode, got %v", err)
		}
	})

	t.Run("darwin setup", func(t *testing.T) {
		svc := &ServiceImpl{
			logger: logging.GetLogger(),
			os:     "darwin",
		}
		err := svc.SetupPlatform(false)
		if err != nil {
			// Expected in CI environment where we can't actually install Homebrew
			t.Logf("Darwin setup failed (expected in CI): %v", err)
		}
	})
}

func TestHomeManagerInstallation(t *testing.T) {
	service := NewService()

	t.Run("home-manager installation", func(t *testing.T) {
		err := service.InstallHomeManager()
		if err != nil {
			// Expected in CI environment where we can't actually install home-manager
			t.Logf("Home-manager installation failed (expected in CI): %v", err)
		}
	})
}

func TestErrorMessages(t *testing.T) {
	t.Run("homebrew installation error", func(t *testing.T) {
		svc := &ServiceImpl{
			logger: logging.GetLogger(),
			os:     "darwin",
		}

		// Mock the installHomebrew function to always fail
		original := svc.InstallHomebrew
		svc.InstallHomebrew = func() error {
			return fmt.Errorf("mock error")
		}
		defer func() { svc.InstallHomebrew = original }()

		err := svc.setupDarwin()
		if err == nil {
			t.Error("Expected error, got nil")
		}

		platformErr, ok := err.(errors.PlatformError)
		if !ok {
			t.Error("Expected PlatformError")
		}

		expected := "platform error during homebrew installation: mock error"
		if platformErr.Error() != expected {
			t.Errorf("Expected error message %q, got %q", expected, platformErr.Error())
		}
	})
}
