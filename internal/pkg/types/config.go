package types

import (
	"fmt"
	"time"
)

// Config represents the main configuration structure
type Config struct {
	LastUpdated time.Time           `yaml:"lastUpdated"`
	Version     string              `yaml:"version"`
	NixConfig   *NixConfig          `yaml:"nixConfig"`
	Settings    Settings            `yaml:"settings"`
	Backup      BackupSettings      `yaml:"backup"`
	Environment EnvironmentSettings `yaml:"environment"`
	Shell       ShellConfig         `yaml:"shell"`
	Editor      EditorConfig        `yaml:"editor"`
	Git         GitConfig           `yaml:"git"`
	Project     ProjectConfig       `yaml:"project"`
	Packages    PackagesConfig      `yaml:"packages"`
	Dependencies []string          `yaml:"dependencies"`
}

// NewDefaultConfig creates a new default configuration
func NewDefaultConfig() *Config {
	return &Config{
		LastUpdated: time.Now(),
		Version:     "1.0",
		NixConfig:   &NixConfig{},
		Settings:    Settings{},
		Project:     NewDefaultProjectConfig(),
	}
}

// Other type definitions...
type Settings struct {
	AutoUpdate     bool   `yaml:"autoUpdate"`
	UpdateInterval string `yaml:"updateInterval"`
	LogLevel       string `yaml:"logLevel"`
}

type NixConfig struct {
	Version     string            `yaml:"version"`
	Settings    Settings          `yaml:"settings"`
	Shell       ShellConfig       `yaml:"shell"`
	Editor      EditorConfig      `yaml:"editor"`
	Git         GitConfig         `yaml:"git"`
	Packages    PackagesConfig    `yaml:"packages"`
	Team        TeamConfig        `yaml:"team"`
	Platform    PlatformConfig    `yaml:"platform"`
	Development DevelopmentConfig `yaml:"development"`
}

// Define all other missing types
type EditorConfig struct {
	Type        string `yaml:"type"`
	ConfigPath  string `yaml:"configPath"`
	PackageName string `yaml:"packageName"`
}

type GitConfig struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}

// ... other struct definitions moved from config/types.go

// Replace undefined types with actual definitions
type BackupSettings struct {
	MaxBackups     int      `yaml:"maxBackups"`
	RetentionDays  int      `yaml:"retentionDays"`
	BackupDir      string   `yaml:"backupDir"`
	ExcludePattern []string `yaml:"excludePattern"`
	Frequency      string   `yaml:"frequency"`
	Location       string   `yaml:"location"`
	Enabled          bool   `yaml:"enabled"`
	CompressionLevel int    `yaml:"compressionLevel"`
}

type EnvironmentSettings struct {
	Name       string `yaml:"name"`
	Value      string `yaml:"value"`
	Default    string `yaml:"default"`
	AutoSwitch bool   `yaml:"autoSwitch"`
	Type       string `yaml:"type"`
}

// Add other missing type definitions

// Add missing type definitions
type ShellConfig struct {
	Type     string `yaml:"type"`
	InitFile string `yaml:"initFile"`
}

type PackagesConfig struct {
	Core []string `yaml:"core"`
	User []string `yaml:"user"`
	Team []string `yaml:"team"`
}

type TeamConfig struct {
	Members []string `yaml:"members"`
}

type PlatformConfig struct {
	OS   string `yaml:"os"`
	Arch string `yaml:"arch"`
}

type DevelopmentConfig struct {
	Debug bool `yaml:"debug"`
}

type UserConfig struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}

type SystemConfig struct {
	Admin bool `yaml:"admin"`
}

// Add these method implementations
func (s Settings) Validate() error {
	if s.AutoUpdate && s.UpdateInterval == "" {
		return fmt.Errorf("update interval required when auto-update is enabled")
	}
	return nil
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

func (c *Config) Validate() error {
	if c.Version == "" {
		return fmt.Errorf("version cannot be empty")
	}
	if err := c.Settings.Validate(); err != nil {
		return fmt.Errorf("settings validation failed: %w", err)
	}
	if err := c.Backup.Validate(); err != nil {
		return fmt.Errorf("backup validation failed: %w", err)
	}
	if err := c.Environment.Validate(); err != nil {
		return fmt.Errorf("environment validation failed: %w", err)
	}
	return nil
}

// Add these validation methods to the ProjectConfig type
func (c *ProjectConfig) Validate() error {
	if err := c.ValidateVersion(); err != nil {
		return fmt.Errorf("version validation failed: %w", err)
	}
	if err := c.ValidateName(); err != nil {
		return fmt.Errorf("name validation failed: %w", err)
	}
	if err := c.ValidateEnvironment(); err != nil {
		return fmt.Errorf("environment validation failed: %w", err)
	}
	if err := c.ValidateSettings(); err != nil {
		return fmt.Errorf("settings validation failed: %w", err)
	}
	if err := c.ValidateDependencies(); err != nil {
		return fmt.Errorf("dependencies validation failed: %w", err)
	}
	return nil
}

func (c *ProjectConfig) ValidateVersion() error {
	if c.Version == "" {
		return fmt.Errorf("version is required")
	}
	validVersions := []string{"1.0", "1.1", "1.2"}
	for _, v := range validVersions {
		if c.Version == v {
			return nil
		}
	}
	return fmt.Errorf("unsupported version: %s", c.Version)
}

func (c *ProjectConfig) ValidateName() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(c.Name) > 50 {
		return fmt.Errorf("name exceeds maximum length of 50 characters")
	}
	return nil
}

func (c *ProjectConfig) ValidateEnvironment() error {
	if c.Environment == "" {
		return fmt.Errorf("environment is required")
	}
	validEnvs := []string{"development", "staging", "production"}
	for _, env := range validEnvs {
		if c.Environment == env {
			return nil
		}
	}
	return fmt.Errorf("invalid environment: %s", c.Environment)
}

func (c *ProjectConfig) ValidateSettings() error {
	validLogLevels := []string{"debug", "info", "warn", "error"}
	found := false
	for _, level := range validLogLevels {
		if c.Settings["logLevel"] == level {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("invalid log level: %s", c.Settings["logLevel"])
	}

	if c.Settings["autoUpdate"] == "true" && c.Settings["updateInterval"] != "" {
		if _, err := time.ParseDuration(c.Settings["updateInterval"]); err != nil {
			return fmt.Errorf("invalid update interval format: %s", c.Settings["updateInterval"])
		}
	}
	return nil
}

func (c *ProjectConfig) ValidateDependencies() error {
	seen := make(map[string]bool)
	for _, dep := range c.Dependencies {
		if seen[dep] {
			return fmt.Errorf("duplicate dependency: %s", dep)
		}
		seen[dep] = true
	}
	return nil
}

// Add these methods
func (c *Config) GetNixConfig() *NixConfig {
	if c == nil {
		return nil
	}
	return c.NixConfig
}

// func (c *Config) GetValue(key string) (interface{}, error) {
// 	// Implementation
// }

// Centralized configuration types
type ProjectConfig struct {
	Name         string
	Version      string
	Environment  string
	Dependencies []string
	Settings     map[string]string
	Tools        []string
	Required     []string
}
