package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitCmd(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	defer func() { os.Setenv("HOME", origHome) }()

	tests := []struct {
		name      string
		args      []string
		flags     map[string]string
		wantFiles []string
		wantErr   bool
	}{
		{
			name: "auto config with defaults",
			args: []string{},
			flags: map[string]string{
				"auto": "true",
			},
			wantFiles: []string{
				".config/nix-foundry/flake.nix",
				".config/nix-foundry/home.nix",
			},
			wantErr: false,
		},
		{
			name: "custom shell and editor",
			args: []string{},
			flags: map[string]string{
				"auto":   "true",
				"shell":  "bash",
				"editor": "nvim",
			},
			wantFiles: []string{
				".config/nix-foundry/flake.nix",
				".config/nix-foundry/home.nix",
			},
			wantErr: false,
		},
		{
			name: "with git config",
			args: []string{},
			flags: map[string]string{
				"auto":      "true",
				"git-name":  "Test User",
				"git-email": "test@example.com",
			},
			wantFiles: []string{
				".config/nix-foundry/flake.nix",
				".config/nix-foundry/home.nix",
			},
			wantErr: false,
		},
		{
			name: "with team config",
			args: []string{},
			flags: map[string]string{
				"auto": "true",
				"team": "backend",
			},
			wantFiles: []string{
				".config/nix-foundry/flake.nix",
				".config/nix-foundry/home.nix",
			},
			wantErr: false,
		},
		{
			name:  "from config file",
			args:  []string{"testdata/config.yaml"},
			flags: map[string]string{},
			wantFiles: []string{
				".config/nix-foundry/flake.nix",
				".config/nix-foundry/home.nix",
			},
			wantErr: false,
		},
		{
			name:    "no config file and no auto",
			args:    []string{},
			flags:   map[string]string{},
			wantErr: true,
		},
		{
			name: "invalid shell",
			args: []string{},
			flags: map[string]string{
				"auto":  "true",
				"shell": "fish",
			},
			wantErr: true,
		},
		{
			name: "invalid editor",
			args: []string{},
			flags: map[string]string{
				"auto":   "true",
				"editor": "code",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean test directory
			testutil.CleanDir(t, filepath.Join(tmpDir, ".config", "nix-foundry"))

			// Create test command
			cmd := newRootCmd()
			cmd.SetArgs(append([]string{"init"}, tt.args...))

			// Set flags
			for k, v := range tt.flags {
				err := cmd.Flags().Set(k, v)
				require.NoError(t, err)
			}

			// Run command
			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Check generated files
			for _, file := range tt.wantFiles {
				path := filepath.Join(tmpDir, file)
				assert.FileExists(t, path)

				// Validate file contents based on flags
				content, err := os.ReadFile(path)
				require.NoError(t, err)

				switch filepath.Base(path) {
				case "home.nix":
					if shell, ok := tt.flags["shell"]; ok {
						assert.Contains(t, string(content), shell)
					}
					if name, ok := tt.flags["git-name"]; ok {
						assert.Contains(t, string(content), name)
					}
				case "flake.nix":
					if team, ok := tt.flags["team"]; ok {
						assert.Contains(t, string(content), team)
					}
				}
			}
		})
	}
}

func TestInitCmd_ConfigValidation(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	tests := []struct {
		name       string
		configName string
		wantErr    string
	}{
		{
			name:       "valid minimal config",
			configName: "minimal",
			wantErr:    "",
		},
		{
			name:       "valid full config",
			configName: "full",
			wantErr:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.CleanDir(t, filepath.Join(tmpDir, ".config", "nix-foundry"))
			configPath := filepath.Join(tmpDir, "config.yaml")
			testutil.WriteTestConfig(t, configPath, tt.configName)

			cmd := newRootCmd()
			cmd.SetArgs([]string{"init", configPath})
			err := cmd.Execute()

			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
