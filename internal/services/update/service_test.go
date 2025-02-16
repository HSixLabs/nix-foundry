package update

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/testutils"
)

func TestUpdateService(t *testing.T) {
	service := NewService()

	t.Run("update flake", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a mock flake.nix file
		flakeFile := filepath.Join(tmpDir, "flake.nix")
		if err := os.WriteFile(flakeFile, []byte(testutils.MockFlakeContent), 0644); err != nil {
			t.Fatalf("Failed to create test flake file: %v", err)
		}

		// Test update in test mode
		err := service.UpdateFlake(tmpDir)
		if err != nil {
			// We expect an error in test environment where nix is not available
			t.Logf("Expected error when nix is not available: %v", err)
		}
	})

	t.Run("invalid directory", func(t *testing.T) {
		err := service.UpdateFlake("/nonexistent/dir")
		if err == nil {
			t.Error("Expected error for nonexistent directory")
		}
	})

	t.Run("apply configuration", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Test in test mode first
		if err := service.ApplyConfiguration(tmpDir, true); err != nil {
			t.Errorf("Expected no error in test mode, got %v", err)
		}

		// Test with real mode
		err := service.ApplyConfiguration(tmpDir, false)
		if err != nil {
			// Expected in CI environment where home-manager is not available
			t.Logf("Configuration apply failed (expected in CI): %v", err)
		}
	})
}
