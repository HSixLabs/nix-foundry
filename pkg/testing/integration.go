// Package testing provides testing utilities.
package testing

import (
	"os"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/pkg/schema"
)

// TestConfig represents a test configuration.
type TestConfig struct {
	Files map[string]string
}

// T is a minimal testing.T interface for our needs.
type T interface {
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

// SetupTest sets up a test environment.
func SetupTest(t T, config TestConfig) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "nix-foundry-test-*")
	if err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	for path, content := range config.Files {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("failed to create directory %s: %v", filepath.Dir(fullPath), err)
		}

		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			os.RemoveAll(tmpDir)
			t.Fatalf("failed to write file %s: %v", fullPath, err)
		}
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

// GetTestConfig returns a test configuration.
func GetTestConfig() *schema.Config {
	return schema.NewDefaultConfig()
}

// CompareConfigs compares two configurations.
func CompareConfigs(t T, expected, actual *schema.Config) {
	if expected.Version != actual.Version {
		t.Errorf("version mismatch: expected %s, got %s", expected.Version, actual.Version)
	}

	if expected.Kind != actual.Kind {
		t.Errorf("kind mismatch: expected %s, got %s", expected.Kind, actual.Kind)
	}

	if expected.Metadata.Name != actual.Metadata.Name {
		t.Errorf("metadata.name mismatch: expected %s, got %s", expected.Metadata.Name, actual.Metadata.Name)
	}

	if expected.Settings.Shell != actual.Settings.Shell {
		t.Errorf("settings.shell mismatch: expected %s, got %s", expected.Settings.Shell, actual.Settings.Shell)
	}

	if expected.Settings.LogLevel != actual.Settings.LogLevel {
		t.Errorf("settings.logLevel mismatch: expected %s, got %s", expected.Settings.LogLevel, actual.Settings.LogLevel)
	}

	if expected.Settings.AutoUpdate != actual.Settings.AutoUpdate {
		t.Errorf("settings.autoUpdate mismatch: expected %v, got %v", expected.Settings.AutoUpdate, actual.Settings.AutoUpdate)
	}

	if expected.Settings.UpdateInterval != actual.Settings.UpdateInterval {
		t.Errorf("settings.updateInterval mismatch: expected %v, got %v", expected.Settings.UpdateInterval, actual.Settings.UpdateInterval)
	}

	if expected.Nix.Manager != actual.Nix.Manager {
		t.Errorf("nix.manager mismatch: expected %s, got %s", expected.Nix.Manager, actual.Nix.Manager)
	}

	if len(expected.Nix.Packages.Core) != len(actual.Nix.Packages.Core) {
		t.Errorf("nix.packages.core length mismatch: expected %d, got %d", len(expected.Nix.Packages.Core), len(actual.Nix.Packages.Core))
	}

	if len(expected.Nix.Packages.Optional) != len(actual.Nix.Packages.Optional) {
		t.Errorf("nix.packages.optional length mismatch: expected %d, got %d", len(expected.Nix.Packages.Optional), len(actual.Nix.Packages.Optional))
	}

	if len(expected.Nix.Scripts) != len(actual.Nix.Scripts) {
		t.Errorf("nix.scripts length mismatch: expected %d, got %d", len(expected.Nix.Scripts), len(actual.Nix.Scripts))
	}
}
