package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/config"
	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/testutil"
	"github.com/spf13/cobra"
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
		name        string
		args        []string
		flags       map[string]string
		wantFiles   []string
		wantErr     bool
		wantErrMsg  string
		wantContent string
	}{
		{
			name: "auto config with defaults",
			args: []string{},
			flags: map[string]string{
				"auto":      "true",
				"test-mode": "true",
				"force":     "true",
			},
			wantFiles: []string{
				".config/nix-foundry/flake.nix",
				".config/nix-foundry/home.nix",
			},
			wantErr:     false,
			wantErrMsg:  "",
			wantContent: "package = pkgs.zsh",
		},
		{
			name: "custom shell and editor",
			args: []string{},
			flags: map[string]string{
				"auto":      "true",
				"shell":     "bash",
				"editor":    "nvim",
				"test-mode": "true",
				"force":     "true",
			},
			wantFiles: []string{
				".config/nix-foundry/flake.nix",
				".config/nix-foundry/home.nix",
			},
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name: "with git config",
			args: []string{},
			flags: map[string]string{
				"auto":      "true",
				"git-name":  "Test User",
				"git-email": "test@example.com",
				"test-mode": "true",
				"force":     "true",
			},
			wantFiles: []string{
				".config/nix-foundry/flake.nix",
				".config/nix-foundry/home.nix",
			},
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name: "with team config",
			args: []string{},
			flags: map[string]string{
				"auto":      "true",
				"team":      "backend",
				"test-mode": "true",
				"force":     "true",
			},
			wantFiles: []string{
				".config/nix-foundry/flake.nix",
				".config/nix-foundry/home.nix",
			},
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name: "from_config_file",
			args: []string{},
			flags: map[string]string{
				"test-mode": "true",
				"force":     "true",
			},
			wantFiles: []string{
				".config/nix-foundry/flake.nix",
				".config/nix-foundry/home.nix",
			},
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name: "no_config_file_and_no_auto",
			args: []string{},
			flags: map[string]string{
				"test-mode": "true",
			},
			wantFiles:  nil,
			wantErr:    true,
			wantErrMsg: "either --auto flag or config file path is required",
		},
		{
			name: "invalid shell",
			args: []string{},
			flags: map[string]string{
				"auto":      "true",
				"shell":     "fish",
				"test-mode": "true",
				"force":     "true",
			},
			wantErr:    true,
			wantErrMsg: "",
		},
		{
			name: "invalid editor",
			args: []string{},
			flags: map[string]string{
				"auto":      "true",
				"editor":    "notepad",
				"test-mode": "true",
				"force":     "true",
			},
			wantErr:    true,
			wantErrMsg: "invalid editor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize config manager for each test
			var err error
			configManager, err = config.NewConfigManager()
			require.NoError(t, err, "Failed to initialize config manager")

			// Clean up after this specific test
			t.Cleanup(func() {
				configManager = nil
				autoConfig = false
				testMode = false
				shell = "zsh"
				editor = "nano"
				gitName = ""
				gitEmail = ""
			})

			// Clean the config directory before each test
			configDir := filepath.Join(tmpDir, ".config", "nix-foundry")
			testutil.CleanDir(t, configDir)

			// For config file test, write the config to the correct location
			if tt.name == "from_config_file" {
				testConfig := &config.NixConfig{
					Version: "1.0",
					Shell: config.ShellConfig{
						Type:    "zsh",
						Plugins: []string{},
					},
					Editor: config.EditorConfig{
						Type: "nano",
					},
				}

				// Write config to the nix-foundry config directory
				configDir := filepath.Join(tmpDir, ".config", "nix-foundry")
				if mkdirErr := os.MkdirAll(configDir, 0755); mkdirErr != nil {
					t.Fatal(mkdirErr)
				}
				configPath := filepath.Join(configDir, "config.yaml")
				err = configManager.WriteConfig(configPath, testConfig)
				require.NoError(t, err, "Failed to write test config")

				// Use the config path directly
				tt.args = []string{configPath}
			}

			cmd := createTestInitCmd()
			args := []string{"init"}
			args = append(args, tt.args...)
			if tt.flags != nil {
				for k, v := range tt.flags {
					args = append(args, fmt.Sprintf("--%s=%s", k, v))
				}
			}
			cmd.SetArgs(args)

			// Execute command
			err = cmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			// Check generated files
			for _, file := range tt.wantFiles {
				path := filepath.Join(tmpDir, file)
				assert.FileExists(t, path)

				// Validate file contents based on flags
				content, err := os.ReadFile(path)
				require.NoError(t, err)

				// Convert flags to config for validation
				config := make(map[string]string)
				for k, v := range tt.flags {
					if k != "test-mode" && k != "auto" { // Skip non-config flags
						config[k] = v
					}
				}

				switch filepath.Base(path) {
				case "home.nix":
					if shellType, hasShell := tt.flags["shell"]; hasShell {
						expectedShell := fmt.Sprintf("package = pkgs.%s", shellType)
						assert.Contains(t, string(content), expectedShell,
							"Expected shell configuration %q not found in content: %s", expectedShell, content)
					}
					if name, ok := tt.flags["git-name"]; ok {
						assert.Contains(t, string(content), fmt.Sprintf("userName = \"%s\"", name))
					}
				case "flake.nix":
					if team, ok := tt.flags["team"]; ok {
						assert.Contains(t, string(content), fmt.Sprintf("team = \"%s\"", team))
					}
				}
			}

			// Check file contents if specified
			if tt.wantContent != "" {
				for _, file := range tt.wantFiles {
					content, err := os.ReadFile(filepath.Join(tmpDir, file))
					assert.NoError(t, err, "Failed to read file: %s", file)
					assert.Contains(t, string(content), tt.wantContent,
						"File %s does not contain expected content", file)
				}
			}
		})
	}
}

func TestInitCmd_ConfigValidation(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmpDir)
	defer func() { os.Setenv("HOME", origHome) }()

	// Initialize config manager
	var err error
	configManager, err = config.NewConfigManager()
	if err != nil {
		t.Fatalf("Failed to initialize config manager: %v", err)
	}

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
			// Clean the config directory
			configDir := filepath.Join(tmpDir, ".config", "nix-foundry")
			testutil.CleanDir(t, configDir)

			// Create test config
			testConfig := &config.NixConfig{
				Version: "1.0",
				Shell: config.ShellConfig{
					Type:    "zsh",
					Plugins: []string{},
				},
				Editor: config.EditorConfig{
					Type: "nano",
				},
			}

			// Add more fields for full config
			if tt.configName == "full" {
				testConfig.Git = config.GitConfig{
					Enable: true,
					User: struct {
						Name  string `yaml:"name"`
						Email string `yaml:"email"`
					}{
						Name:  "Test User",
						Email: "test@example.com",
					},
				}
				testConfig.Shell.Plugins = []string{"zsh-autosuggestions"}
			}

			err := configManager.WriteConfig("config.yaml", testConfig)
			require.NoError(t, err, "Failed to write test config")

			cmd := createTestInitCmd()
			cmd.SetArgs([]string{"init", "--test-mode", "--force", "--auto"})
			err = cmd.Execute()

			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}

	// Reset configManager after tests
	t.Cleanup(func() {
		configManager = nil
	})
}

func createTestInitCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "nix-foundry",
		Short: "Development environment manager",
	}

	// Add the actual initCmd from the main package
	rootCmd.AddCommand(initCmd)

	return rootCmd
}
