// Package schema provides configuration schema definitions.
package schema

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// MultiLineString is a string that will be marshaled as a multiline string in YAML.
type MultiLineString string

// MarshalYAML implements the yaml.Marshaler interface.
func (s MultiLineString) MarshalYAML() (interface{}, error) {
	// Create a yaml.Node with style set to literal (|)
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Style: yaml.LiteralStyle,
		Value: string(s),
	}, nil
}

// Config represents the configuration file structure.
type Config struct {
	Version  string   `yaml:"version"`
	Kind     string   `yaml:"kind"`
	Metadata Metadata `yaml:"metadata"`
	Settings Settings `yaml:"settings"`
	Nix      Nix      `yaml:"nix"`
}

// Metadata contains configuration metadata.
type Metadata struct {
	Name        string    `yaml:"name"`
	Description string    `yaml:"description"`
	Created     time.Time `yaml:"created"`
	Updated     time.Time `yaml:"updated"`
}

// Settings contains global settings.
type Settings struct {
	Shell          string        `yaml:"shell"`
	LogLevel       string        `yaml:"logLevel"`
	AutoUpdate     bool          `yaml:"autoUpdate"`
	UpdateInterval time.Duration `yaml:"updateInterval"`
}

// Nix contains Nix-specific configuration.
type Nix struct {
	Manager  string   `yaml:"manager"`
	Packages Packages `yaml:"packages"`
	Scripts  []Script `yaml:"scripts,omitempty"`
}

// Packages contains package lists.
type Packages struct {
	Core     []string `yaml:"core"`
	Optional []string `yaml:"optional,omitempty"`
}

// Script represents a shell script.
type Script struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description,omitempty"`
	Commands    MultiLineString `yaml:"commands"`
}

// NewDefaultConfig creates a new configuration with default values.
func NewDefaultConfig() *Config {
	now := time.Now()
	return &Config{
		Version: "v1",
		Kind:    "NixConfig",
		Metadata: Metadata{
			Name:        "default",
			Description: "Default configuration",
			Created:     now,
			Updated:     now,
		},
		Settings: Settings{
			Shell:          "bash",
			LogLevel:       "info",
			AutoUpdate:     true,
			UpdateInterval: 24 * time.Hour,
		},
		Nix: Nix{
			Manager: "nix-env",
			Packages: Packages{
				Core: []string{
					"git",
					"curl",
					"wget",
				},
			},
		},
	}
}

// ValidateConfig validates the configuration.
func ValidateConfig(config *Config) error {
	if config.Version == "" {
		return fmt.Errorf("version is required")
	}

	if config.Kind == "" {
		return fmt.Errorf("kind is required")
	}

	if config.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}

	if config.Settings.Shell == "" {
		return fmt.Errorf("settings.shell is required")
	}

	if config.Settings.LogLevel == "" {
		return fmt.Errorf("settings.logLevel is required")
	}

	if config.Nix.Manager == "" {
		return fmt.Errorf("nix.manager is required")
	}

	if len(config.Nix.Packages.Core) == 0 {
		return fmt.Errorf("nix.packages.core must not be empty")
	}

	return nil
}

// GetConfigPath returns the path to the configuration file.
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(homeDir, ".config", "nix-foundry", "config.yaml"), nil
}
