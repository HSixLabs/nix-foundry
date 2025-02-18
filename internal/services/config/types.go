package config

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/types"
)

// Define wrapper types that embed base types
// type Settings struct {
// 	AutoUpdate     bool   `yaml:"autoUpdate"`
// 	UpdateInterval string `yaml:"updateInterval"`
// 	LogLevel       string `yaml:"logLevel"`
// }

// LocalSettings wraps the base Settings type
type LocalSettings struct {
	types.Settings
}

// NewSettings creates a new LocalSettings from base Settings
func NewSettings(base types.Settings) LocalSettings {
	return LocalSettings{Settings: base}
}

// ToBaseSettings converts LocalSettings back to base Settings
func (s LocalSettings) ToBaseSettings() types.Settings {
	return s.Settings
}

type EnvironmentSettings struct {
	types.EnvironmentSettings
}

// Add conversion methods
func NewEnvironmentSettings(base types.EnvironmentSettings) EnvironmentSettings {
	return EnvironmentSettings{EnvironmentSettings: base}
}

func (e EnvironmentSettings) ToBaseEnvironmentSettings() types.EnvironmentSettings {
	return e.EnvironmentSettings
}

// LocalConfig wraps the imported Config type to allow adding methods
type LocalConfig struct {
	*types.Config
}

// NewLocalConfig creates a new LocalConfig instance
func NewLocalConfig(cfg *types.Config) *LocalConfig {
	return &LocalConfig{Config: cfg}
}

// GetValue retrieves a configuration value by key
func (c *LocalConfig) GetValue(key string) (interface{}, error) {
	if c == nil || c.Config == nil {
		return nil, fmt.Errorf("config is nil")
	}

	switch key {
	case "settings.autoUpdate":
		return c.Settings.AutoUpdate, nil
	case "settings.updateInterval":
		return c.Settings.UpdateInterval, nil
	case "settings.logLevel":
		return c.Settings.LogLevel, nil
	default:
		return nil, fmt.Errorf("invalid key: %s", key)
	}
}

// Validate performs configuration validation
func (c *LocalConfig) Validate() error {
	if c == nil || c.Config == nil {
		return fmt.Errorf("config is nil")
	}

	// Validate backup settings
	if err := c.Backup.Validate(); err != nil {
		return fmt.Errorf("backup configuration invalid: %w", err)
	}

	// Validate environment settings
	if err := c.Environment.Validate(); err != nil {
		return fmt.Errorf("environment configuration invalid: %w", err)
	}

	return nil
}

// // Config represents the main configuration structure
// type Config struct {
// 	Version      string              `yaml:"version"`
// 	NixConfig    *NixConfig         `yaml:"nix_config"`
// 	Settings     Settings           `yaml:"settings"`
// 	Environment  EnvironmentSettings `yaml:"environment"`
// 	Project      ProjectConfig       `yaml:"project"`
// }

// // Define supporting types directly instead of aliasing
// type NixConfig struct {
// 	Version     string            `yaml:"version"`
// 	Settings    Settings          `yaml:"settings"`
// 	Shell       ShellConfig       `yaml:"shell"`
// 	Editor      EditorConfig      `yaml:"editor"`
// 	Git         GitConfig         `yaml:"git"`
// 	Packages    PackagesConfig    `yaml:"packages"`
// 	Team        TeamConfig        `yaml:"team"`
// 	Platform    PlatformConfig    `yaml:"platform"`
// 	Development DevelopmentConfig `yaml:"development"`
// }

// Add missing type definitions
// type PackagesConfig struct {
// 	Core         []string            `yaml:"core"`
// 	User         []string            `yaml:"user"`
// 	Team         []string            `yaml:"team"`
// 	PlatformSpecific map[string][]string `yaml:"platformSpecific"`
// }

type TeamConfig struct {
	Enable   bool              `yaml:"enable"`
	Name     string            `yaml:"name"`
	Settings map[string]string `yaml:"settings"`
}

type PlatformConfig struct {
	OS   string `yaml:"os"`
	Arch string `yaml:"arch"`
}

type DevelopmentConfig struct {
	Languages struct {
		Go     struct{ Version string; Packages []string }
		Node   struct{ Version string; Packages []string }
		Python struct{ Version string; Packages []string }
	} `yaml:"languages"`
	Tools []string `yaml:"tools"`
}

func (e EnvironmentSettings) Validate() error {
	if e.Name == "" {
		return fmt.Errorf("environment name cannot be empty")
	}
	if e.Type == "" {
		return fmt.Errorf("environment type cannot be empty")
	}
	return nil
}

func (b BackupSettings) Validate() error {
	if b.MaxBackups < 1 {
		return fmt.Errorf("maxBackups must be at least 1")
	}
	if b.RetentionDays < 1 {
		return fmt.Errorf("retentionDays must be at least 1")
	}
	if b.BackupDir == "" {
		return fmt.Errorf("backupDir cannot be empty")
	}
	return nil
}

// Add these type definitions
type UserConfig struct {
	Version     string            `yaml:"version"`
	Settings    Settings          `yaml:"settings"`
	Environment EnvironmentSettings `yaml:"environment"`
}

type SystemConfig struct {
	Version     string            `yaml:"version"`
	Settings    Settings          `yaml:"settings"`
	Environment EnvironmentSettings `yaml:"environment"`
}

type Previewer interface {
	Preview(config *Config) error
}

// Helper methods for Settings conversion
func SettingsToMap(s types.Settings) map[string]string {
	return map[string]string{
		"autoUpdate":     fmt.Sprintf("%v", s.AutoUpdate),
		"updateInterval": s.UpdateInterval,
		"logLevel":      s.LogLevel,
	}
}

func MapToSettings(m map[string]string) types.Settings {
	autoUpdate := m["autoUpdate"] == "true"
	return types.Settings{
		AutoUpdate:     autoUpdate,
		UpdateInterval: m["updateInterval"],
		LogLevel:       m["logLevel"],
	}
}

// Add these methods to LocalConfig
func (c *LocalConfig) GetNixConfig() *types.NixConfig {
	if c == nil || c.Config == nil {
		return nil
	}
	return c.NixConfig
}

// NixConfigToLocalConfig converts a NixConfig to LocalConfig
func NixConfigToLocalConfig(n *types.NixConfig) *LocalConfig {
	if n == nil {
		return nil
	}
	return &LocalConfig{
		Config: &types.Config{
			NixConfig: n,
		},
	}
}

// LocalConfigToNixConfig converts a LocalConfig to NixConfig
func (c *LocalConfig) ToNixConfig() *types.NixConfig {
	if c == nil || c.Config == nil {
		return nil
	}
	return c.NixConfig
}

// Add these methods to Config
func (c *Config) GetNixConfig() *types.NixConfig {
	if c == nil {
		return nil
	}
	// Convert the local NixConfig to types.NixConfig
	return &types.NixConfig{
		Version:     c.NixConfig.Version,
		Settings:    types.Settings(c.NixConfig.Settings),
		Shell:       types.ShellConfig(c.NixConfig.Shell),
		Editor:      types.EditorConfig(c.NixConfig.Editor),
		Git:         types.GitConfig(c.NixConfig.Git),
		Packages:    types.PackagesConfig(c.NixConfig.Packages),
		Team:        types.TeamConfig(c.NixConfig.Team),
		Platform:    types.PlatformConfig(c.NixConfig.Platform),
		Development: types.DevelopmentConfig(c.NixConfig.Development),
	}
}

func (c *Config) GetValue(key string) (interface{}, error) {
	if c == nil {
		return nil, fmt.Errorf("config is nil")
	}

	switch key {
	case "settings.autoUpdate":
		return c.Settings.AutoUpdate, nil
	case "settings.updateInterval":
		return c.Settings.UpdateInterval, nil
	case "settings.logLevel":
		return c.Settings.LogLevel, nil
	default:
		return nil, fmt.Errorf("invalid key: %s", key)
	}
}

type BackupSettings struct {
	Enabled          bool   `yaml:"enabled"`
	RetentionDays    int    `yaml:"retentionDays"`
	MaxBackups       int    `yaml:"maxBackups"`
	CompressionLevel int    `yaml:"compressionLevel"`
	BackupDir        string `yaml:"backupDir"`
}

// ConfigPreviewer defines the interface for configuration preview generation
type ConfigPreviewer interface {
	GeneratePreview(cfg *types.Config) error
}

// Consolidate duplicate struct declarations

type Config struct {
	NixConfig    types.NixConfig
	Project      types.ProjectConfig
	Packages     types.PackagesConfig
	Team         types.TeamConfig
	Platform     types.PlatformConfig
	Development  types.DevelopmentConfig
	Settings     LocalSettings
	Environment  EnvironmentSettings
}

// Remove duplicate declarations of NixConfig, PackagesConfig, etc.
// Keep single definitions in this file

// Add conversion method if needed
func (c *Config) ToServiceConfig() *types.Config {
	return &types.Config{
		Version:     c.NixConfig.Version,
		NixConfig:   &c.NixConfig,
		Project:     c.Project,
		Packages:    c.Packages,
		Settings:    c.Settings.ToBaseSettings(),
		Environment: c.Environment.ToBaseEnvironmentSettings(),
		Shell:       c.NixConfig.Shell,
		Editor:      c.NixConfig.Editor,
		Git:         c.NixConfig.Git,
	}
}
