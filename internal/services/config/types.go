package config

import (
	"fmt"
	"time"
)

// Config represents the main configuration structure
type Config struct {
	LastUpdated time.Time           `yaml:"lastUpdated"`
	Version     string              `yaml:"version"`
	Settings    Settings            `yaml:"settings"`
	Backup      BackupSettings      `yaml:"backup"`
	Environment EnvironmentSettings `yaml:"environment"`
	Shell       ShellConfig         `yaml:"shell"`
	Editor      EditorConfig        `yaml:"editor"`
	Git         GitConfig           `yaml:"git"`
}

// EnvironmentSettings contains environment-related configuration
type EnvironmentSettings struct {
	Default    string `yaml:"default"`
	AutoSwitch bool   `yaml:"autoSwitch"`
}

// BackupSettings contains backup-related configuration
type BackupSettings struct {
	MaxBackups     int      `yaml:"maxBackups"`
	RetentionDays  int      `yaml:"retentionDays"`
	BackupDir      string   `yaml:"backupDir"`
	ExcludePattern []string `yaml:"excludePattern"`
}

// ShellConfig represents shell configuration
type ShellConfig struct {
	Type    string   `yaml:"type"`
	Plugins []string `yaml:"plugins,omitempty"`
}

// EditorConfig represents editor configuration
type EditorConfig struct {
	Type    string   `yaml:"type"`
	Plugins []string `yaml:"plugins,omitempty"`
}

// GitConfig represents git configuration
type GitConfig struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}

// Settings struct definition
type Settings struct {
	AutoUpdate     bool   `yaml:"autoUpdate"`
	UpdateInterval string `yaml:"updateInterval"`
	LogLevel       string `yaml:"logLevel"`
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

func (s Settings) Validate() error {
	if s.AutoUpdate && s.UpdateInterval == "" {
		return fmt.Errorf("update interval required when auto-update is enabled")
	}

	return nil
}

func (e EnvironmentSettings) Validate() error {
	if e.Default == "" {
		return fmt.Errorf("default environment cannot be empty")
	}

	validEnvs := []string{"development", "staging", "production"}
	for _, env := range validEnvs {
		if e.Default == env {
			return nil
		}
	}
	return fmt.Errorf("invalid default environment: %s", e.Default)
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
