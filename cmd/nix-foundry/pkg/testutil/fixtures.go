package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/config"
	"gopkg.in/yaml.v3"
)

// TestConfigs contains sample configurations for testing
var TestConfigs = map[string]config.NixConfig{
	"minimal": {
		Shell: config.ShellConfig{
			Type: "zsh",
		},
		Editor: config.EditorConfig{
			Type: "vim",
		},
		Git: config.GitConfig{
			Enable: true,
		},
		Packages: config.PackagesConfig{
			Additional:  []string{},
			Development: []string{"git"},
		},
	},
	"full": {
		Shell: config.ShellConfig{
			Type: "zsh",
		},
		Editor: config.EditorConfig{
			Type:       "nvim",
			Extensions: []string{"telescope.nvim"},
			Settings: map[string]interface{}{
				"lineNumbers": true,
			},
		},
		Git: config.GitConfig{
			Enable: true,
			User: struct {
				Name  string `yaml:"name"`
				Email string `yaml:"email"`
			}{
				Name:  "Test User",
				Email: "test@example.com",
			},
			Config: map[string]string{
				"pull.rebase": "true",
			},
		},
		Packages: config.PackagesConfig{
			Additional: []string{"ripgrep", "fd"},
			PlatformSpecific: map[string][]string{
				"darwin": {"mas"},
				"linux":  {"inotify-tools"},
			},
			Development: []string{
				"git",
				"jq",
				"curl",
			},
			Team: map[string][]string{
				"backend": {"postgresql", "redis"},
			},
		},
		Team: config.TeamConfig{
			Enable: true,
			Name:   "backend",
			Settings: map[string]string{
				"database": "postgresql",
			},
		},
	},
}

// WriteTestConfig writes a test configuration to a file
func WriteTestConfig(t *testing.T, path string, name string) {
	cfg, ok := TestConfigs[name]
	if !ok {
		t.Fatalf("Test config %s not found", name)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("Failed to marshal test config: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}
}

// CreateTestBase creates a base configuration for testing
func CreateTestBase(t *testing.T, configDir string) {
	baseConfig := config.UserConfig{
		Shell: "zsh",
		Git: config.GitConfig{
			Enable: true,
			User: struct {
				Name  string `yaml:"name"`
				Email string `yaml:"email"`
			}{
				Name:  "Base User",
				Email: "base@example.com",
			},
		},
		Packages: []string{"git", "curl"},
	}

	basePath := filepath.Join(configDir, "bases", "base.yaml")
	if err := os.MkdirAll(filepath.Dir(basePath), 0755); err != nil {
		t.Fatalf("Failed to create base directory: %v", err)
	}

	data, err := yaml.Marshal(baseConfig)
	if err != nil {
		t.Fatalf("Failed to marshal base config: %v", err)
	}

	if err := os.WriteFile(basePath, data, 0644); err != nil {
		t.Fatalf("Failed to write base config: %v", err)
	}
}

// CleanDir removes and recreates a directory
func CleanDir(t *testing.T, dir string) {
	if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to remove directory %s: %v", dir, err)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}
}
