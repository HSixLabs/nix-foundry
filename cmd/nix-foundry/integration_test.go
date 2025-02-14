package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/testutil"
	"github.com/spf13/cobra"
)

// Test-specific variables
var (
	testRootCmd   *cobra.Command
	testConfigDir string
	testHomeDir   string
)

func init() {
	// Initialize test mode flag only if not already set
	if os.Getenv("NIX_FOUNDRY_TEST_MODE") == "" {
		os.Setenv("NIX_FOUNDRY_TEST_MODE", "true")
	}
}

func newTestRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nix-foundry",
		Short: "A tool for setting up development environments",
		Long: `nix-foundry helps you set up consistent development environments
across different platforms using Nix, with smart defaults and easy customization.`,
	}

	// Initialize flags
	var testMode bool
	cmd.PersistentFlags().BoolVar(&testMode, "test-mode", false, "Run in test mode")

	// Add all commands
	cmd.AddCommand(initCmd)
	cmd.AddCommand(configCmd)
	cmd.AddCommand(updateCmd)
	cmd.AddCommand(applyCmd)
	cmd.AddCommand(backupCmd)
	cmd.AddCommand(doctorCmd)
	cmd.AddCommand(backupRestoreCmd)

	// Initialize packages command hierarchy
	packagesCmd := &cobra.Command{Use: "packages", Short: "Manage custom packages"}
	packagesCmd.AddCommand(addPackageCmd)
	packagesCmd.AddCommand(removePackageCmd)
	packagesCmd.AddCommand(listPackagesCmd)
	cmd.AddCommand(packagesCmd)

	// Set test mode for all commands
	cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if testMode {
			os.Setenv("NIX_FOUNDRY_TEST_MODE", "true")
		}
	}

	return cmd
}

func TestPackageManagement(t *testing.T) {
	// Setup test-specific root command
	testRootCmd = newTestRootCmd()
	testRootCmd.SetOut(io.Discard)
	testRootCmd.SetErr(io.Discard)
	t.Cleanup(func() {
		if testRootCmd != nil {
			testRootCmd.ResetCommands()
			testRootCmd = nil
		}
	})

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
			// Add packages by setting args on rootCmd and executing it
			for _, pkg := range tt.packages {
				testRootCmd.SetArgs([]string{"packages", "add", pkg})

				if err := testRootCmd.Execute(); err != nil {
					t.Fatalf("Failed to add package %s: %v", pkg, err)
				}

				// Reset args after execution
				testRootCmd.SetArgs([]string{})
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
	// Create a test-specific temporary directory
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Setup test-specific root command
	testRootCmd = newTestRootCmd()
	testRootCmd.SetOut(io.Discard)
	testRootCmd.SetErr(io.Discard)

	// Create test configuration
	testConfig := []byte("test configuration")
	configDir := filepath.Join(tmpDir, ".config", "nix-foundry")
	backupDir := filepath.Join(configDir, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		t.Fatalf("Failed to create backup directory: %v", err)
	}

	testFile := filepath.Join(configDir, "test.nix")
	if err := os.WriteFile(testFile, testConfig, 0644); err != nil {
		t.Fatalf("Failed to write test configuration: %v", err)
	}

	// Mock tar command with a function that simulates the backup/restore
	mockTar := testutil.MockCommand(t, "tar", `#!/bin/sh
	case "$1" in
		"-czf")
			# Creating backup - copy the test file to the backup location
			cp "`+testFile+`" "$2"
			;;
		"-xzf")
			# Extracting backup - copy the backup file back to the test location
			mkdir -p "$(dirname "`+testFile+`")"
			cp "$2" "`+testFile+`"
			;;
	esac
	exit 0`)
	defer mockTar()

	// Mock rsync command
	mockRsync := testutil.MockCommand(t, "rsync", `#!/bin/sh
	# Just pretend to sync files
	exit 0`)
	defer mockRsync()

	// Reset args before test
	testRootCmd.SetArgs([]string{})

	// Execute backup command with test mode
	testRootCmd.SetArgs([]string{"backup", "create", "--test-mode"})
	if err := testRootCmd.Execute(); err != nil {
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

	// Reset the root command for restore
	testRootCmd.ResetCommands()
	testRootCmd = newTestRootCmd()
	testRootCmd.SetOut(io.Discard)
	testRootCmd.SetErr(io.Discard)

	// Execute restore command
	testRootCmd.SetArgs([]string{"restore", backupFiles[0], "--test-mode"})
	if err := testRootCmd.Execute(); err != nil {
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

	// Ensure cleanup of test resources
	t.Cleanup(func() {
		if testRootCmd != nil {
			testRootCmd.ResetCommands()
			testRootCmd = nil
		}
		// Clean up test directory is handled by t.TempDir()
	})
}

// TestMain handles test initialization and cleanup
func TestMain(m *testing.M) {
	// Store original state
	origEnv := os.Getenv("NIX_FOUNDRY_TEST_MODE")
	origHome := os.Getenv("HOME")
	origRootCmd := testRootCmd

	// Create base temporary directory for all tests
	baseTestDir, err := os.MkdirTemp("", fmt.Sprintf("nix-foundry-test-%d-*", time.Now().UnixNano()))
	if err != nil {
		fmt.Printf("Failed to create base test directory: %v\n", err)
		os.Exit(1)
	}

	// Set up test directories
	testHomeDir = filepath.Join(baseTestDir, "home")
	testConfigDir = filepath.Join(testHomeDir, ".config", "nix-foundry-test")

	// Create test home directory structure
	if err := os.MkdirAll(testConfigDir, 0755); err != nil {
		fmt.Printf("Failed to create test directories: %v\n", err)
		os.RemoveAll(baseTestDir)
		os.Exit(1)
	}

	// Set HOME environment variable to test home directory
	os.Setenv("HOME", testHomeDir)

	// Ensure cleanup runs even if tests panic
	defer func() {
		// Recover from panics
		if r := recover(); r != nil {
			fmt.Printf("Recovered from panic in TestMain: %v\n", r)
		}

		cleanupOrder := []func(){
			func() {
				if testRootCmd != nil {
					testRootCmd.ResetCommands()
					testRootCmd = nil
				}
			},
			func() {
				if err := os.RemoveAll(baseTestDir); err != nil {
					fmt.Printf("Failed to cleanup test directories: %v\n", err)
				}
			},
			func() {
				if origEnv != "" {
					os.Setenv("NIX_FOUNDRY_TEST_MODE", origEnv)
				} else {
					os.Unsetenv("NIX_FOUNDRY_TEST_MODE")
				}
				os.Setenv("HOME", origHome)
				testRootCmd = origRootCmd
			},
		}

		for _, cleanup := range cleanupOrder {
			cleanup()
		}
	}()

	// Run tests and capture the exit code
	code := m.Run()

	// Exit with the proper code after cleanup
	os.Exit(code)
}
