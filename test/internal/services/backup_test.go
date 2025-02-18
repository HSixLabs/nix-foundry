package services

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/services/backup"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
)

// Define maxBackups constant
const maxBackups = 10

// Add type conversion for the test
func asService(ps project.ProjectService) project.Service {
	return ps.(project.Service)
}

func TestBackupService(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "nix-foundry-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFiles := []string{
		"config.yaml",
		"environments/default.nix",
		"packages/custom.nix",
	}

	for _, file := range testFiles {
		path := filepath.Join(tempDir, file)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create directory for %s: %v", file, err)
		}
		if err := os.WriteFile(path, []byte("test content"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Create service dependencies
	configSvc := config.NewService()
	platformSvc := platform.NewService()
	envSvc := environment.NewService(tempDir, configSvc, platformSvc, true, true, true)
	projectSvc := project.NewService(configSvc, envSvc, nil)

	t.Run("CreateAndListBackups", func(t *testing.T) {
		svc := backup.NewService(configSvc, envSvc, asService(projectSvc))

		// Create a backup
		if err := svc.Create("test-backup", false); err != nil {
			t.Fatalf("Failed to create backup: %v", err)
		}

		// List backups
		backups, err := svc.List()
		if err != nil {
			t.Fatalf("Failed to list backups: %v", err)
		}

		if len(backups) != 1 || backups[0].ID != "test-backup" {
			t.Errorf("Expected one backup named 'test-backup', got %v", backups)
		}
	})

	t.Run("RestoreBackup", func(t *testing.T) {
		svc := backup.NewService(configSvc, envSvc, asService(projectSvc))

		// Create and modify test files
		modifiedContent := "modified content"
		if err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte(modifiedContent), 0644); err != nil {
			t.Fatalf("Failed to modify test file: %v", err)
		}

		// Restore the backup
		if err := svc.Restore("test-backup", true); err != nil {
			t.Fatalf("Failed to restore backup: %v", err)
		}

		// Verify restored content
		content, err := os.ReadFile(filepath.Join(tempDir, "config.yaml"))
		if err != nil {
			t.Fatalf("Failed to read restored file: %v", err)
		}

		if string(content) != "test content" {
			t.Errorf("Expected restored content to be 'test content', got '%s'", content)
		}
	})

	t.Run("DeleteBackup", func(t *testing.T) {
		svc := backup.NewService(configSvc, envSvc, asService(projectSvc))

		// Delete the backup
		if err := svc.Delete("test-backup"); err != nil {
			t.Fatalf("Failed to delete backup: %v", err)
		}

		// Verify backup was deleted
		backups, err := svc.List()
		if err != nil {
			t.Fatalf("Failed to list backups: %v", err)
		}

		if len(backups) != 0 {
			t.Errorf("Expected no backups after deletion, got %v", backups)
		}
	})

	t.Run("RotateBackups", func(t *testing.T) {
		svc := backup.NewService(configSvc, envSvc, asService(projectSvc))

		// Create more than maxBackups backups
		for i := 0; i < maxBackups+5; i++ {
			backupName := fmt.Sprintf("test-backup-%d", i)
			if err := svc.Create(backupName, false); err != nil {
				t.Fatalf("Failed to create backup %s: %v", backupName, err)
			}
			// Add a small delay to ensure different modification times
			time.Sleep(10 * time.Millisecond)
		}

		// Create some safety backups that should be preserved
		safetyBackups := []string{
			"pre-restore-backup1",
			"pre-restore-backup2",
		}
		for _, name := range safetyBackups {
			if err := svc.Create(name, false); err != nil {
				t.Fatalf("Failed to create safety backup %s: %v", name, err)
			}
		}

		// Rotate backups with 7 day retention
		if err := svc.Rotate(7 * 24 * time.Hour); err != nil {
			t.Fatalf("Failed to rotate backups: %v", err)
		}

		// List backups and verify count
		backups, err := svc.List()
		if err != nil {
			t.Fatalf("Failed to list backups: %v", err)
		}

		// Should have maxBackups regular backups + safety backups
		expectedCount := maxBackups + len(safetyBackups)
		if len(backups) != expectedCount {
			t.Errorf("Expected %d backups after rotation, got %d", expectedCount, len(backups))
		}

		// Verify safety backups were preserved
		for _, name := range safetyBackups {
			found := false
			for _, backup := range backups {
				if backup.ID == name {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Safety backup %s was not preserved", name)
			}
		}
	})
}
