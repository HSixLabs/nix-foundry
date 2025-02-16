package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/commands"
)

func TestInitCommand(t *testing.T) {
	testCases := []struct {
		name        string
		args        []string
		expectError bool
		setup       func(string) error
	}{
		{
			name:        "fresh initialization",
			args:        []string{"--yes"}, // Auto-confirm
			expectError: false,
		},
		{
			name:        "force reinitialization",
			args:        []string{"--force", "--yes"},
			expectError: false,
			setup: func(dir string) error {
				// Create existing config
				return os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("existing: true"), 0644)
			},
		},
		{
			name:        "test mode initialization",
			args:        []string{"--test", "--yes"},
			expectError: false,
		},
		{
			name:        "invalid config directory",
			args:        []string{"--config", "/invalid/path"},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temp directory for test
			tempDir := t.TempDir()

			// Run setup if provided
			if tc.setup != nil {
				if err := tc.setup(tempDir); err != nil {
					t.Fatalf("Failed to setup test: %v", err)
				}
			}

			// Create command
			cmd := commands.NewInitCommand()

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)

			// Set config directory and add test args
			args := append(tc.args, "--config", tempDir)
			cmd.SetArgs(args)

			// Execute command
			err := cmd.Execute()

			// Verify error state
			if (err != nil) != tc.expectError {
				t.Errorf("Expected error: %v, got: %v", tc.expectError, err)
			}

			// Verify expected files/directories exist
			if !tc.expectError {
				// Check config file
				if _, err := os.Stat(filepath.Join(tempDir, "config.yaml")); os.IsNotExist(err) {
					t.Error("Config file not created")
				}

				// Check required directories
				requiredDirs := []string{
					"environments",
					"backups",
					"packages",
				}
				for _, dir := range requiredDirs {
					if _, err := os.Stat(filepath.Join(tempDir, dir)); os.IsNotExist(err) {
						t.Errorf("Required directory %s not created", dir)
					}
				}

				// Verify output contains success message
				if !bytes.Contains(buf.Bytes(), []byte("Initialization complete")) {
					t.Error("Expected success message in output")
				}
			}
		})
	}
}
