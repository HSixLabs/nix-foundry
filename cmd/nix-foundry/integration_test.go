package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/testutil"
	"github.com/spf13/cobra"
)

func init() {
	// Remove the global initialization of rootCmd and its subcommands
	// Since we are now creating a new rootCmd instance in each test
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nix-foundry",
		Short: "A tool for setting up development environments",
		Long: `nix-foundry helps you set up consistent development environments
across different platforms using Nix, with smart defaults and easy customization.`,
	}

	// Initialize packages command hierarchy
	packagesCmd := &cobra.Command{Use: "packages", Short: "Manage custom packages"}
	packagesCmd.AddCommand(addPackageCmd)
	packagesCmd.AddCommand(removePackageCmd)
	packagesCmd.AddCommand(listPackagesCmd)

	// Initialize backup command hierarchy
	backupCmd.AddCommand(createCmd)
	backupCmd.AddCommand(listCmd)
	backupCmd.AddCommand(restoreCmd)
	backupCmd.AddCommand(deleteCmd)

	// Add all commands to root
	cmd.AddCommand(packagesCmd)
	cmd.AddCommand(backupCmd)
	cmd.AddCommand(initCmd)
	cmd.AddCommand(updateCmd)
	cmd.AddCommand(doctorCmd)

	return cmd
}

func TestPackageManagement(t *testing.T) {
	homeDir, cleanup := testutil.SetupTestHome(t)
	defer cleanup()

	// Create config directory and initial packages.json
	configDir := filepath.Join(homeDir, ".config", "nix-foundry")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Initialize empty packages.json as an empty JSON array
	initialPackages := []string{}
	initialData, marshalErr := json.MarshalIndent(initialPackages, "", "  ")
	if marshalErr != nil {
		t.Fatalf("Failed to marshal initial packages: %v", marshalErr)
	}
	if writeErr := os.WriteFile(filepath.Join(configDir, "packages.json"), initialData, 0644); writeErr != nil {
		t.Fatalf("Failed to write initial packages.json: %v", writeErr)
	}

	// Mock nix commands
	mockNix := testutil.MockCommand(t, "nix", "mock nix output")
	defer mockNix()

	tests := []struct {
		name     string
		packages []string
		want     []string
	}{
		{
			name:     "add single package",
			packages: []string{"vim"},
			want:     []string{"vim"},
		},
		{
			name:     "add multiple packages",
			packages: []string{"emacs", "git"},
			want:     []string{"vim", "emacs", "git"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a new instance of rootCmd for each test
			rootCmd := newRootCmd()

			// Suppress command output
			rootCmd.SetOut(io.Discard)
			rootCmd.SetErr(io.Discard)

			// Add packages by setting args on rootCmd and executing it
			for _, pkg := range tt.packages {
				rootCmd.SetArgs([]string{"packages", "add", pkg})

				if err := rootCmd.Execute(); err != nil {
					t.Fatalf("Failed to add package %s: %v", pkg, err)
				}

				// Reset rootCmd args after execution
				rootCmd.SetArgs([]string{})
			}

			// Verify packages were added
			packagesFile := filepath.Join(configDir, "packages.json")
			data, err := os.ReadFile(packagesFile)
			if err != nil {
				t.Fatalf("Failed to read packages file: %v", err)
			}

			var packages []string
			if err := json.Unmarshal(data, &packages); err != nil {
				t.Fatalf("Failed to parse packages file: %v", err)
			}

			if len(packages) != len(tt.want) {
				t.Errorf("Got %d packages, want %d", len(packages), len(tt.want))
			}

			packageMap := make(map[string]bool)
			for _, pkg := range packages {
				packageMap[pkg] = true
			}

			for _, pkg := range tt.want {
				if !packageMap[pkg] {
					t.Errorf("Package %s not found in result", pkg)
				}
			}
		})
	}
}

func TestBackupRestore(t *testing.T) {
	homeDir, cleanup := testutil.SetupTestHome(t)
	defer cleanup()

	// Create config directory structure
	configDir := filepath.Join(homeDir, ".config", "nix-foundry")
	backupDir := filepath.Join(configDir, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}

	// Create test configuration
	testConfig := []byte("test configuration")
	if err := os.WriteFile(filepath.Join(configDir, "test.nix"), testConfig, 0644); err != nil {
		t.Fatalf("Failed to write test configuration: %v", err)
	}

	// Use a new instance of rootCmd
	rootCmd := newRootCmd()

	// Suppress command output
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)

	// Execute backup command
	rootCmd.SetArgs([]string{"backup"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Failed to create backup: %v", err)
	}

	// Verify backup was created
	backupFiles, globErr := filepath.Glob(filepath.Join(configDir, "backups", "*.tar.gz"))
	if globErr != nil {
		t.Fatalf("Failed to list backups: %v", globErr)
	}
	if len(backupFiles) != 1 {
		t.Fatalf("Expected 1 backup, got %d", len(backupFiles))
	}

	// Remove test configuration
	if err := os.Remove(filepath.Join(configDir, "test.nix")); err != nil {
		t.Fatalf("Failed to remove test configuration: %v", err)
	}

	// Create a new rootCmd instance for restore
	rootCmd = newRootCmd()
	rootCmd.SetOut(io.Discard)
	rootCmd.SetErr(io.Discard)
	// Execute restore command
	rootCmd.SetArgs([]string{"restore", backupFiles[0]})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Failed to restore backup: %v", err)
	}

	// Verify restored configuration
	restored, err := os.ReadFile(filepath.Join(configDir, "test.nix"))
	if err != nil {
		t.Fatalf("Failed to read restored configuration: %v", err)
	}

	if string(restored) != string(testConfig) {
		t.Errorf("Restored configuration doesn't match original")
	}
}
