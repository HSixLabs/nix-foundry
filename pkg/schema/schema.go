/*
Package schema provides configuration schema definitions for Nix Foundry.
It defines the structure and validation rules for configuration files,
supporting different configuration types and inheritance.
*/
package schema

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/shawnkhoffman/nix-foundry/pkg/platform"
	"gopkg.in/yaml.v3"
)

/*
MultiLineString is a string that will be marshaled as a multiline string in YAML.
It implements custom YAML marshaling to ensure proper formatting of multi-line content.
*/
type MultiLineString string

/*
MarshalYAML implements the yaml.Marshaler interface.
It ensures the string is formatted as a literal block scalar in YAML.
*/
func (s MultiLineString) MarshalYAML() (interface{}, error) {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Style: yaml.LiteralStyle,
		Value: string(s),
	}, nil
}

/*
ConfigType represents the type of configuration.
Valid types are UserConfig, TeamConfig, and ProjectConfig.
*/
type ConfigType string

const (
	UserConfig    ConfigType = "user"
	TeamConfig    ConfigType = "team"
	ProjectConfig ConfigType = "project"
)

/*
Config represents the configuration file structure.
It contains metadata, settings, and Nix-specific configuration.
*/
type Config struct {
	Version  string     `yaml:"version"`
	Kind     string     `yaml:"kind"`
	Type     ConfigType `yaml:"type"`
	Base     string     `yaml:"base,omitempty"`
	Metadata Metadata   `yaml:"metadata"`
	Settings Settings   `yaml:"settings"`
	Nix      Nix        `yaml:"nix"`
}

/*
Metadata contains configuration metadata including name, description,
and timestamps for creation and updates.
*/
type Metadata struct {
	Name        string    `yaml:"name"`
	Description string    `yaml:"description"`
	Created     time.Time `yaml:"created"`
	Updated     time.Time `yaml:"updated"`
}

/*
Settings contains global settings for the configuration.
This includes shell preferences, logging settings, and update configurations.
*/
type Settings struct {
	Shell          string        `yaml:"shell,omitempty"`
	LogLevel       string        `yaml:"logLevel"`
	AutoUpdate     bool          `yaml:"autoUpdate"`
	UpdateInterval time.Duration `yaml:"updateInterval"`
}

/*
Nix contains Nix-specific configuration.
This includes package manager settings, package lists, and shell scripts.
*/
type Nix struct {
	Manager  string   `yaml:"manager"`
	Packages Packages `yaml:"packages"`
	Scripts  []Script `yaml:"scripts,omitempty"`
}

/*
Packages contains package lists.
It separates packages into core (required) and optional packages.
*/
type Packages struct {
	Core     []string `yaml:"core,omitempty"`
	Optional []string `yaml:"optional,omitempty"`
}

/*
Script represents a shell script.
It includes the script's name, description, and commands to execute.
*/
type Script struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description"`
	Commands    MultiLineString `yaml:"commands"`
}

/*
NewDefaultConfig creates a new configuration with default values.
It sets up a basic user configuration with standard settings.
*/
func NewDefaultConfig() *Config {
	now := time.Now()
	return &Config{
		Version: "1.0.0",
		Kind:    "Config",
		Type:    UserConfig,
		Metadata: Metadata{
			Name:        "default",
			Description: "Default user configuration",
			Created:     now,
			Updated:     now,
		},
		Settings: Settings{
			LogLevel:       "info",
			AutoUpdate:     true,
			UpdateInterval: 24 * time.Hour,
		},
		Nix: Nix{
			Manager:  "nix-env",
			Packages: Packages{},
		},
	}
}

/*
NewTeamConfig creates a new team configuration.
It initializes a configuration for team-wide settings and packages.
*/
func NewTeamConfig(name string) *Config {
	now := time.Now()
	return &Config{
		Version: "1.0.0",
		Kind:    "Config",
		Type:    TeamConfig,
		Metadata: Metadata{
			Name:        name,
			Description: fmt.Sprintf("Team configuration for %s", name),
			Created:     now,
			Updated:     now,
		},
		Settings: Settings{
			LogLevel:       "info",
			AutoUpdate:     true,
			UpdateInterval: 24 * time.Hour,
		},
		Nix: Nix{
			Manager:  "nix-env",
			Packages: Packages{},
		},
	}
}

/*
NewProjectConfig creates a new project configuration.
It initializes a configuration for project-specific settings and packages.
*/
func NewProjectConfig(name string) *Config {
	now := time.Now()
	return &Config{
		Version: "1.0.0",
		Kind:    "Config",
		Type:    ProjectConfig,
		Metadata: Metadata{
			Name:        name,
			Description: fmt.Sprintf("Project configuration for %s", name),
			Created:     now,
			Updated:     now,
		},
		Settings: Settings{
			LogLevel:       "info",
			AutoUpdate:     true,
			UpdateInterval: 24 * time.Hour,
		},
		Nix: Nix{
			Manager:  "nix-env",
			Packages: Packages{},
		},
	}
}

/*
ValidateConfig validates the configuration.
It checks for required fields and enforces configuration type-specific rules.
*/
func ValidateConfig(config *Config) error {
	if config.Type == UserConfig && config.Settings.Shell == "" {
		return fmt.Errorf("shell setting is required for user configs")
	}

	if (config.Type == TeamConfig || config.Type == ProjectConfig) && len(config.Nix.Packages.Core) == 0 {
		return fmt.Errorf("core packages are required for team and project configs")
	}

	return nil
}

/*
GetConfigPath returns the path to the configuration file.
It constructs the path based on the user's home directory.
*/
func GetConfigPath() (string, error) {
	homeDir, err := platform.GetRealUserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(homeDir, ".config", "nix-foundry", "config.yaml"), nil
}

/*
PackageDiff represents the difference between two package configurations.
*/
type PackageDiff struct {
	ToInstall []string
	ToRemove  []string
}

/*
DiffPackages compares currently installed packages with desired packages and returns the differences.
installedPackages should be the result of querying nix-env -q.
desiredPackages is the Packages struct from the configuration.
*/
func DiffPackages(installedPackages []string, desiredPackages Packages) PackageDiff {
	var diff PackageDiff

	installedMap := make(map[string]bool)
	for _, pkg := range installedPackages {
		installedMap[pkg] = true
	}

	desiredMap := make(map[string]bool)
	for _, pkg := range desiredPackages.Core {
		desiredMap[pkg] = true
	}
	for _, pkg := range desiredPackages.Optional {
		desiredMap[pkg] = true
	}

	for pkg := range desiredMap {
		if !installedMap[pkg] {
			diff.ToInstall = append(diff.ToInstall, pkg)
		}
	}

	for pkg := range installedMap {
		if !desiredMap[pkg] {
			diff.ToRemove = append(diff.ToRemove, pkg)
		}
	}

	return diff
}
