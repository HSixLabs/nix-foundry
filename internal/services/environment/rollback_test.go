package environment

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/validation"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
)

func TestRollback(t *testing.T) {
	tempDir := t.TempDir()
	service := NewService(
		tempDir,
		config.NewService(),
		validation.NewService(),
		platform.NewService(),
	)

	// Create test backup
	backupTime := time.Now().Add(-1 * time.Hour)
	backupPath := filepath.Join(tempDir, "backups", backupTime.Format("20060102-150405"))
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		t.Fatalf("Failed to create test backup: %v", err)
	}

	t.Run("valid rollback", func(t *testing.T) {
		err := service.Rollback(backupTime, false)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("missing backup", func(t *testing.T) {
		err := service.Rollback(time.Now().Add(-24*time.Hour), false)
		if err == nil {
			t.Error("Expected error for missing backup")
		}
	})
}
