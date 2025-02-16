package backup

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/validation"
	"github.com/shawnkhoffman/nix-foundry/internal/services/backup"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
	"github.com/spf13/cobra"
)

// Updated command creation function
func NewBackupCommand() *cobra.Command {
	return NewCreateCmd(backup.NewService(
		config.NewService(),
		environment.NewService("", config.NewService(), validation.NewService(), platform.NewService()),
		project.NewService(config.NewService(), environment.NewService("", config.NewService(), validation.NewService(), platform.NewService()), packages.NewService("")),
	))
}

// Add maxBackups constant for rotation tests
const maxBackups = 10

func TestBackups(t *testing.T) {
	// Setup test service
	tempDir := t.TempDir()

	// Create config service with directory
	cfgManager := config.NewService()

	validator := validation.NewService()
	platformSvc := platform.NewService()
	envSvc := environment.NewService(
		tempDir,
		cfgManager,
		validator,
		platformSvc,
	)
	pkgSvc := packages.NewService(tempDir)
	projectSvc := project.NewService(cfgManager, envSvc, pkgSvc)
	svc := backup.NewService(cfgManager, envSvc, projectSvc)

	// Create test configuration directory structure
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

	// Set up environment for testing
	os.Setenv("NIX_FOUNDRY_CONFIG_DIR", tempDir)
	defer os.Unsetenv("NIX_FOUNDRY_CONFIG_DIR")

	t.Run("CreateBackup", func(t *testing.T) {
		cmd := NewCmd(svc)
		cmd.SetArgs([]string{"create", "test-backup"})
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Failed to execute backup command: %v", err)
		}

		// Verify backup file exists
		backupPath := filepath.Join(tempDir, "backups", "test-backup.tar.gz")
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			t.Errorf("Backup file was not created at %s", backupPath)
		}
	})

	t.Run("ListBackups", func(t *testing.T) {
		cmd := NewListCmd(svc)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Failed to execute list-backups command: %v", err)
		}

		output := buf.String()
		if output == "No backups found" {
			t.Error("Expected to find backups, but none were listed")
		}
	})

	t.Run("DeleteBackup", func(t *testing.T) {
		cmd := NewDeleteCmd(svc)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"test-backup", "--force"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Failed to execute delete-backup command: %v", err)
		}

		// Verify backup file was deleted
		backupPath := filepath.Join(tempDir, "backups", "test-backup.tar.gz")
		if _, err := os.Stat(backupPath); !os.IsNotExist(err) {
			t.Error("Backup file was not deleted")
		}
	})

	t.Run("RestoreBackup", func(t *testing.T) {
		// First create a backup to restore
		backupCmd := NewBackupCommand()
		backupCmd.SetArgs([]string{"restore-test"})
		if err := backupCmd.Execute(); err != nil {
			t.Fatalf("Failed to create test backup: %v", err)
		}

		// Modify a file
		if err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte("modified content"), 0644); err != nil {
			t.Fatalf("Failed to modify test file: %v", err)
		}

		// Restore the backup
		cmd := NewRestoreCmd(svc)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"restore-test", "--force", "--test"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Failed to execute restore command: %v", err)
		}

		// Verify file was restored
		content, err := os.ReadFile(filepath.Join(tempDir, "config.yaml"))
		if err != nil {
			t.Fatalf("Failed to read restored file: %v", err)
		}

		if string(content) != "test content" {
			t.Errorf("Expected restored content to be 'test content', got '%s'", string(content))
		}
	})

	t.Run("RotateBackups", func(t *testing.T) {
		// Create multiple backups
		backupCmd := NewBackupCommand()
		for i := 0; i < 15; i++ { // Create more than maxBackups
			backupName := fmt.Sprintf("test-backup-%d", i)
			backupCmd.SetArgs([]string{backupName})
			if err := backupCmd.Execute(); err != nil {
				t.Fatalf("Failed to create backup %s: %v", backupName, err)
			}
			time.Sleep(10 * time.Millisecond) // Ensure different timestamps
		}

		// Create some safety backups
		safetyBackups := []string{"pre-restore-backup1", "pre-restore-backup2"}
		for _, name := range safetyBackups {
			backupCmd.SetArgs([]string{name})
			if err := backupCmd.Execute(); err != nil {
				t.Fatalf("Failed to create safety backup %s: %v", name, err)
			}
		}

		// Execute rotate command
		cmd := NewRotateCmd(svc)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"--force"}) // Skip confirmation in tests

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Failed to execute rotate-backups command: %v", err)
		}

		// List backups and verify count
		listCmd := NewListCmd(svc)
		listBuf := new(bytes.Buffer)
		listCmd.SetOut(listBuf)
		if err := listCmd.Execute(); err != nil {
			t.Fatalf("Failed to list backups: %v", err)
		}

		// Count the backups in the output
		backupCount := 0
		safetyCount := 0
		for _, line := range strings.Split(listBuf.String(), "\n") {
			if strings.HasPrefix(line, "- pre-restore-") {
				safetyCount++
			} else if strings.HasPrefix(line, "- ") {
				backupCount++
			}
		}

		// Verify counts
		if backupCount > maxBackups {
			t.Errorf("Expected at most %d regular backups after rotation, got %d", maxBackups, backupCount)
		}
		if safetyCount != len(safetyBackups) {
			t.Errorf("Expected %d safety backups after rotation, got %d", len(safetyBackups), safetyCount)
		}
	})

	t.Run("ShowBackupConfig", func(t *testing.T) {
		cmd := newConfigShowCmd(svc)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Failed to execute show-backup-config command: %v", err)
		}

		output := buf.String()
		// Verify output contains expected sections
		expectedFields := []string{
			"Maximum backups:",
			"Maximum age:",
			"Compression level:",
		}
		for _, field := range expectedFields {
			if !strings.Contains(output, field) {
				t.Errorf("Expected output to contain %q", field)
			}
		}
	})

	t.Run("SetBackupConfig", func(t *testing.T) {
		// Test setting individual values
		testCases := []struct {
			name    string
			args    []string
			wantErr bool
		}{
			{"set max backups", []string{"--max-backups", "20"}, false},
			{"set max age", []string{"--max-age", "60"}, false},
			{"set compression", []string{"--compression", "9"}, false},
			{"invalid compression", []string{"--compression", "10"}, true},
			{"multiple settings", []string{"--max-backups", "15", "--max-age", "45", "--compression", "7"}, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cmd := newConfigSetCmd(svc)
				cmd.SetArgs(tc.args)

				err := cmd.Execute()
				if tc.wantErr {
					if err == nil {
						t.Error("Expected error but got none")
					}
				} else {
					if err != nil {
						t.Errorf("Unexpected error: %v", err)
					}
				}

				// Verify configuration was updated
				if err == nil {
					showCmd := newConfigShowCmd(svc)
					showBuf := new(bytes.Buffer)
					showCmd.SetOut(showBuf)
					if err := showCmd.Execute(); err != nil {
						t.Fatalf("Failed to show config: %v", err)
					}

					// Verify output contains updated values
					output := showBuf.String()
					for _, arg := range tc.args {
						if strings.HasPrefix(arg, "--") {
							continue
						}
						if !strings.Contains(output, arg) {
							t.Errorf("Expected output to contain %q", arg)
						}
					}
				}
			})
		}
	})
}
