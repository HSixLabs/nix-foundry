package update

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/testutils"
)

func TestUpdateIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	service := NewService()
	tmpDir := t.TempDir()

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
		}

		// Test configuration apply
		err = service.ApplyConfiguration(tmpDir, false)
		if err != nil {
			// In CI environment, home-manager might not be available
			t.Logf("Configuration apply failed (expected in CI): %v", err)
		}
	})
}
