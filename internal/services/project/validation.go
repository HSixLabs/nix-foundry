package project

import (
	"fmt"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
)

func (c *Config) validateVersion() error {
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

func (c *Config) validateName() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(c.Name) > 50 {
		return fmt.Errorf("name exceeds maximum length of 50 characters")
	}
	return nil
}

func (c *Config) validateEnvironment() error {
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

func (c *Config) validateSettings() error {
	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error"}
	found := false
	for _, level := range validLogLevels {
		if c.Settings.LogLevel == level {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("invalid log level: %s", c.Settings.LogLevel)
	}

	// Validate update interval if auto-update is enabled
	if c.Settings.AutoUpdate && c.Settings.UpdateInterval != "" {
		// Parse duration to validate format
		if _, err := time.ParseDuration(c.Settings.UpdateInterval); err != nil {
			return fmt.Errorf("invalid update interval format: %s", c.Settings.UpdateInterval)
		}
	}

	return nil
}

func (c *Config) validateDependencies() error {
	seen := make(map[string]bool)
	for _, dep := range c.Dependencies {
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
	if personalSettings.LogLevel != s.projectConfig.Settings.LogLevel {
		return errors.NewConflictError(
			fmt.Errorf("log level mismatch: personal=%s, project=%s",
				personalSettings.LogLevel, s.projectConfig.Settings.LogLevel),
			"log level settings conflict")
	}

	// Check auto-update settings
	if personalSettings.AutoUpdate != s.projectConfig.Settings.AutoUpdate {
		return errors.NewConflictError(
			fmt.Errorf("auto-update setting mismatch: personal=%v, project=%v",
				personalSettings.AutoUpdate, s.projectConfig.Settings.AutoUpdate),
			"auto-update settings conflict")
	}

	// Check update interval if auto-update is enabled
	if personalSettings.AutoUpdate &&
		personalSettings.UpdateInterval != s.projectConfig.Settings.UpdateInterval {
		return errors.NewConflictError(
			fmt.Errorf("update interval mismatch: personal=%s, project=%s",
				personalSettings.UpdateInterval, s.projectConfig.Settings.UpdateInterval),
			"update interval settings conflict")
	}

	return nil
}
