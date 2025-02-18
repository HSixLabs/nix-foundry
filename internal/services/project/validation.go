package project

import (
	"fmt"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	configtypes "github.com/shawnkhoffman/nix-foundry/internal/services/config/types"
)

func Validate(cfg *configtypes.Config) error {
	if cfg.Project.Version == "" {
		return fmt.Errorf("version is required")
	}
	if cfg.Project.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(cfg.Project.Name) > 50 {
		return fmt.Errorf("name exceeds maximum length of 50 characters")
	}
	if cfg.Project.Environment == "" {
		return fmt.Errorf("environment is required")
	}
	validEnvs := []string{"development", "staging", "production"}
	valid := false
	for _, env := range validEnvs {
		if cfg.Project.Environment == env {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid environment: %s", cfg.Project.Environment)
	}
	return nil
}

func (c *Config) validateVersion() error {
	if c.Project.Version == "" {
		return fmt.Errorf("version is required")
	}
	validVersions := []string{"1.0", "1.1", "1.2"}
	for _, v := range validVersions {
		if c.Project.Version == v {
			return nil
		}
	}
	return fmt.Errorf("unsupported version: %s", c.Project.Version)
}

func (c *Config) validateName() error {
	if c.Project.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(c.Project.Name) > 50 {
		return fmt.Errorf("name exceeds maximum length of 50 characters")
	}
	return nil
}

func (c *Config) validateEnvironment() error {
	if c.Project.Environment == "" {
		return fmt.Errorf("environment is required")
	}
	validEnvs := []string{"development", "staging", "production"}
	for _, env := range validEnvs {
		if c.Project.Environment == env {
			return nil
		}
	}
	return fmt.Errorf("invalid environment: %s", c.Project.Environment)
}

func (c *Config) validateSettings() error {
	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error"}
	found := false
	for _, level := range validLogLevels {
		if c.Project.Settings["logLevel"] == level {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("invalid log level: %s", c.Project.Settings["logLevel"])
	}

	// Validate update interval if auto-update is enabled
	if c.Project.Settings["autoUpdate"] == "true" && c.Project.Settings["updateInterval"] != "" {
		// Parse duration to validate format
		if _, err := time.ParseDuration(c.Project.Settings["updateInterval"]); err != nil {
			return fmt.Errorf("invalid update interval format: %s", c.Project.Settings["updateInterval"])
		}
	}

	return nil
}

func (c *Config) validateDependencies() error {
	seen := make(map[string]bool)
	for _, dep := range c.Project.Dependencies {
		if dep == "" {
			return fmt.Errorf("empty dependency name")
		}
		if seen[dep] {
			return fmt.Errorf("duplicate dependency: %s", dep)
		}
		seen[dep] = true
	}
	return nil
}

func (s *ServiceImpl) validateSettingsConflicts(personalSettings config.Settings) error {
	if s.projectConfig == nil {
		return errors.NewValidationError("", fmt.Errorf("no configuration loaded"), "cannot validate conflicts with nil configuration")
	}

	// Check log level conflicts
	logLevel := s.projectConfig.Settings.LogLevel
	personalLogLevel := personalSettings.LogLevel
	if personalLogLevel != logLevel {
		return errors.NewConflictError(
			"settings.log_level",
			fmt.Errorf("log level mismatch: personal=%s, project=%s",
				personalLogLevel, logLevel),
			"log level settings conflict",
			"Align log level settings between personal and project configurations",
		)
	}

	// Check auto-update settings
	autoUpdate := s.projectConfig.Settings.AutoUpdate
	personalAutoUpdate := personalSettings.AutoUpdate
	if personalAutoUpdate != autoUpdate {
		return errors.NewConflictError(
			"settings.auto_update",
			fmt.Errorf("auto-update setting mismatch: personal=%v, project=%v",
				personalAutoUpdate, autoUpdate),
			"auto-update settings conflict",
			"Ensure auto-update settings match between personal and project configurations",
		)
	}

	// Check update interval if auto-update is enabled
	personalUpdateInterval := personalSettings.UpdateInterval
	projectUpdateInterval := s.projectConfig.Settings.UpdateInterval
	if personalUpdateInterval != projectUpdateInterval {
		return errors.NewConflictError(
			"settings.update_interval",
			fmt.Errorf("update interval mismatch: personal=%s, project=%s",
				personalUpdateInterval, projectUpdateInterval),
			"update interval settings conflict",
			"Synchronize update intervals between personal and project configurations",
		)
	}

	return nil
}

func validateVersion(cfg *Config) error {
	if cfg.Project.Version == "" {
		return fmt.Errorf("version is required")
	}
	validVersions := []string{"1.0", "1.1", "1.2"}
	for _, v := range validVersions {
		if cfg.Project.Version == v {
			return nil
		}
	}
	return fmt.Errorf("unsupported version: %s", cfg.Project.Version)
}
