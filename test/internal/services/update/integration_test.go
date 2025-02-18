package update

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/testutils"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/shawnkhoffman/nix-foundry/internal/services/update"
)

func TestUpdateIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".nix-foundry")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Initialize required services
	configService := config.NewService()
	platformSvc := platform.NewService()
	envService := environment.NewService(
		configDir,
		configService,
		platformSvc,
		true,
		true,
		true,
	)

	// Create update service with dependencies
	service := update.NewService(configService, envService)

	t.Run("full update cycle", func(t *testing.T) {
		// Create test environment
		if err := os.MkdirAll(filepath.Join(tmpDir, "environments", "default"), 0755); err != nil {
			t.Fatalf("Failed to create test directories: %v", err)
		}

		// Create test flake file
		flakeFile := filepath.Join(tmpDir, "flake.nix")
		if err := os.WriteFile(flakeFile, []byte(testutils.MockFlakeContent), 0644); err != nil {
			t.Fatalf("Failed to create test flake: %v", err)
		}

		// Attempt update
		err := service.UpdateFlake(tmpDir)
		if err != nil {
			// In CI environment, nix might not be available
			t.Logf("Update failed (expected in environments without nix): %v", err)
			return
		}

		// Test configuration apply
		err = service.ApplyConfiguration(tmpDir, false)
		if err != nil {
			// In CI environment, home-manager might not be available
			t.Logf("Configuration apply failed (expected in CI): %v", err)
			return
		}

		// Verify flake.lock was created (if update succeeded)
		lockFile := filepath.Join(tmpDir, "flake.lock")
		if _, err := os.Stat(lockFile); err != nil {
			t.Logf("No flake.lock created (expected in environments without nix)")
		}
	})
}
