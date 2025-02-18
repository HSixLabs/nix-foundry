package backup

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/services/backup"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
	"github.com/spf13/cobra"
)

// Command constructors for testing
func NewCreateCmd(svc backup.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new backup",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			force, _ := cmd.Flags().GetBool("force")
			return svc.Create(args[0], force)
		},
	}
	cmd.Flags().Bool("force", false, "Force backup creation")
	return cmd
}

func NewListCmd(svc backup.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all backups",
		RunE: func(cmd *cobra.Command, args []string) error {
			backups, err := svc.List()
			if err != nil {
				return err
			}
			if len(backups) == 0 {
				cmd.Println("No backups found")
				return nil
			}
			for _, b := range backups {
				cmd.Printf("%s (%s)\n", b.ID, b.Timestamp.Format(time.RFC3339))
			}
			return nil
		},
	}
}

func NewDeleteCmd(svc backup.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete a backup",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return svc.Delete(args[0])
		},
	}
	cmd.Flags().Bool("force", false, "Force deletion")
	return cmd
}

func NewRestoreCmd(svc backup.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore [name]",
		Short: "Restore a backup",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			force, _ := cmd.Flags().GetBool("force")
			return svc.Restore(args[0], force)
		},
	}
	cmd.Flags().Bool("force", false, "Force restore")
	cmd.Flags().Bool("test", false, "Test restore without applying")
	return cmd
}

func NewRotateCmd(svc backup.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "rotate",
		Short: "Rotate old backups",
		RunE: func(cmd *cobra.Command, args []string) error {
			days, _ := cmd.Flags().GetInt("days")
			return svc.Rotate(time.Duration(days) * 24 * time.Hour)
		},
	}
}

func newConfigCmd(svc backup.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage backup configuration",
	}
	cmd.AddCommand(
		newConfigShowCmd(svc),
		newConfigSetCmd(svc),
	)
	return cmd
}

func newConfigShowCmd(svc backup.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show backup configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := svc.GetConfig()
			cmd.Printf("MaxBackups: %d\n", cfg.MaxBackups)
			cmd.Printf("RetentionDays: %d\n", cfg.RetentionDays)
			return nil
		},
	}
}

func newConfigSetCmd(svc backup.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set backup configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			maxBackups, _ := cmd.Flags().GetInt("max-backups")
			retentionDays, _ := cmd.Flags().GetInt("retention-days")
			return svc.UpdateConfig(backup.Config{
				MaxBackups:    maxBackups,
				RetentionDays: retentionDays,
			})
		},
	}
	cmd.Flags().Int("max-backups", 10, "Maximum number of backups to keep")
	cmd.Flags().Int("retention-days", 30, "Number of days to keep backups")
	return cmd
}

func newEncryptCmd(svc backup.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "encrypt [name] [key-path]",
		Short: "Encrypt a backup",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return svc.EncryptBackup(args[0], args[1])
		},
	}
}

func newDecryptCmd(svc backup.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "decrypt [name] [key-path]",
		Short: "Decrypt a backup",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return svc.DecryptBackup(args[0], args[1])
		},
	}
}

// Add helper function for interface conversion
func asService(ps project.ProjectService) project.Service {
	return ps.(project.Service)
}

// Updated command creation function
func NewBackupCommand() *cobra.Command {
	configSvc := config.NewService()
	platformSvc := platform.NewService()

	envSvc := environment.NewService(
		"", // empty string for test
		configSvc,
		platformSvc,
		true, // Enable test mode for testing
		true, // Enable isolation for test environment
		true, // Enable auto-install
	)

	pkgSvc := packages.NewService("")
	projectSvc := project.NewService(configSvc, envSvc, pkgSvc)

	return NewCreateCmd(backup.NewService(
		configSvc,
		envSvc,
		asService(projectSvc),
	))
}

// Add maxBackups constant for rotation tests
const maxBackups = 10

func TestBackups(t *testing.T) {
	// Setup test service
	tempDir := t.TempDir()

	// Create config service with directory
	cfgManager := config.NewService()
	platformSvc := platform.NewService()

	envSvc := environment.NewService(
		tempDir,
		cfgManager,
		platformSvc,
		true, // Enable test mode for testing
		true, // Enable isolation for test environment
		true, // Enable auto-install
	)

	pkgSvc := packages.NewService(tempDir)
	projectSvc := project.NewService(cfgManager, envSvc, pkgSvc)

	// Use asService helper function for interface conversion
	svc := backup.NewService(cfgManager, envSvc, asService(projectSvc))

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
		cmd := NewCreateCmd(svc)
		cmd.SetArgs([]string{"test-backup"})
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

	t.Run("EncryptBackup", func(t *testing.T) {
		// First create a backup to encrypt
		createCmd := NewCreateCmd(svc)
		createCmd.SetArgs([]string{"test-encrypt-backup"})
		if err := createCmd.Execute(); err != nil {
			t.Fatalf("Failed to create test backup: %v", err)
		}

		// Create a temporary key file
		keyPath := filepath.Join(tempDir, "test-key.pem")
		if err := os.WriteFile(keyPath, []byte("test-key-content"), 0600); err != nil {
			t.Fatalf("Failed to create test key file: %v", err)
		}

		// Execute encrypt command
		cmd := newEncryptCmd(svc)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"test-encrypt-backup", keyPath})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Failed to execute encrypt command: %v", err)
		}

		// Verify encryption (backup file should be different)
		backupPath := filepath.Join(tempDir, "backups", "test-encrypt-backup.tar.gz.enc")
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			t.Errorf("Encrypted backup file was not created at %s", backupPath)
		}
	})

	t.Run("DecryptBackup", func(t *testing.T) {
		// Use the previously encrypted backup
		keyPath := filepath.Join(tempDir, "test-key.pem")

		// Execute decrypt command
		cmd := newDecryptCmd(svc)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"test-encrypt-backup", keyPath})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Failed to execute decrypt command: %v", err)
		}

		// Verify decryption (original backup file should be restored)
		backupPath := filepath.Join(tempDir, "backups", "test-encrypt-backup.tar.gz")
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			t.Errorf("Decrypted backup file was not created at %s", backupPath)
		}
	})

	t.Run("ConfigCommands", func(t *testing.T) {
		// Create and test the config command group
		configCmd := newConfigCmd(svc)
		if configCmd == nil {
			t.Fatal("Config command should not be nil")
		}

		// Test show subcommand
		showBuf := new(bytes.Buffer)
		showCmd := configCmd.Commands()[0] // show command
		showCmd.SetOut(showBuf)

		if err := showCmd.Execute(); err != nil {
			t.Fatalf("Failed to execute config show command: %v", err)
		}

		output := showBuf.String()
		if !strings.Contains(output, "MaxBackups") {
			t.Error("Config show output should contain MaxBackups")
		}

		// Test set subcommand
		setBuf := new(bytes.Buffer)
		setCmd := configCmd.Commands()[1] // set command
		setCmd.SetOut(setBuf)
		setCmd.SetArgs([]string{"--max-backups=15", "--retention-days=45"})

		if err := setCmd.Execute(); err != nil {
			t.Fatalf("Failed to execute config set command: %v", err)
		}

		// Verify the changes by running show again
		showBuf.Reset()
		if err := showCmd.Execute(); err != nil {
			t.Fatalf("Failed to execute config show command after update: %v", err)
		}

		updatedOutput := showBuf.String()
		if !strings.Contains(updatedOutput, "MaxBackups: 15") {
			t.Error("Config was not updated correctly")
		}
	})
}

func TestBackupCommands(t *testing.T) {
	configService := config.NewService()
	platformSvc := platform.NewService()

	envService := environment.NewService(
		configService.GetConfigDir(),
		configService,
		platformSvc,
		true, // Enable test mode for testing
		true, // Enable isolation for test environment
		true, // Enable auto-install
	)

	pkgSvc := packages.NewService(configService.GetConfigDir())
	projectSvc := project.NewService(configService, envService, pkgSvc)

	// Use asService helper function for interface conversion
	backupSvc := backup.NewService(configService, envService, asService(projectSvc))

	t.Run("create backup", func(t *testing.T) {
		cmd := NewCreateCmd(backupSvc)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"test-backup", "--force"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Failed to execute backup command: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Backup created") {
			t.Errorf("Expected success message in output")
		}
	})
}

func TestRestoreCommand(t *testing.T) {
	configService := config.NewService()
	platformSvc := platform.NewService()

	envService := environment.NewService(
		configService.GetConfigDir(),
		configService,
		platformSvc,
		true, // Enable test mode for testing
		true, // Enable isolation for test environment
		true, // Enable auto-install
	)

	pkgSvc := packages.NewService(configService.GetConfigDir())
	projectSvc := project.NewService(configService, envService, pkgSvc)

	// Use asService helper function here as well
	backupSvc := backup.NewService(configService, envService, asService(projectSvc))

	t.Run("restore backup", func(t *testing.T) {
		// First create a backup to restore
		createCmd := NewCreateCmd(backupSvc)
		createCmd.SetArgs([]string{"test-backup", "--force"})
		if err := createCmd.Execute(); err != nil {
			t.Fatalf("Failed to create test backup: %v", err)
		}

		// Then try to restore it
		cmd := NewRestoreCmd(backupSvc)
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetArgs([]string{"test-backup", "--force"})

		if err := cmd.Execute(); err != nil {
			t.Fatalf("Failed to execute restore command: %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "Backup restored") {
			t.Errorf("Expected success message in output")
		}
	})
}
