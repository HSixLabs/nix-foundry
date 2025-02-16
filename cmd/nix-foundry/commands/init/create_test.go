package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/commands/config"
)

func TestInitCommand(t *testing.T) {
	testCases := []struct {
		name        string
		args        []string
		force       bool
		expectError bool
	}{
		{"happy path", []string{}, false, false},
		{"force existing", []string{"--force"}, true, false},
		{"invalid config dir", []string{"--config", "/invalid/path"}, false, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			cmd := config.NewInitCommand()

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetArgs(append(tc.args, "--config", tempDir))

			err := cmd.Execute()
			if (err != nil) != tc.expectError {
				t.Errorf("Unexpected error state: %v", err)
			}

			if !tc.expectError {
				if _, err := os.Stat(filepath.Join(tempDir, "config.yaml")); os.IsNotExist(err) {
					t.Errorf("Expected config file not found")
				}
			}
		})
	}
}
