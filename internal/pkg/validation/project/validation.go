package project

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
)

// ValidateConfig validates a project configuration
func ValidateConfig(cfg *project.Config) error {
	if err := validateVersion(cfg); err != nil {
		return fmt.Errorf("invalid version: %w", err)
	}

	if err := validateName(cfg); err != nil {
		return fmt.Errorf("invalid name: %w", err)
	}

	if err := validateEnvironment(cfg); err != nil {
		return fmt.Errorf("invalid environment: %w", err)
	}

	if err := validateSettings(cfg); err != nil {
		return fmt.Errorf("invalid settings: %w", err)
	}

	if err := validateDependencies(cfg); err != nil {
		return fmt.Errorf("invalid dependencies: %w", err)
	}

	return nil
}

func validateVersion(cfg *project.Config) error {
	if cfg.Version == "" {
		return fmt.Errorf("version is required")
	}
	validVersions := []string{"1.0", "1.1", "1.2"}
	for _, v := range validVersions {
		if cfg.Version == v {
			return nil
		}
	}
	return fmt.Errorf("unsupported version: %s", cfg.Version)
}

func validateName(cfg *project.Config) error {
	if cfg.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(cfg.Name) > 50 {
		return fmt.Errorf("name exceeds maximum length of 50 characters")
	}
	return nil
}

func validateEnvironment(cfg *project.Config) error {
	if cfg.Environment == "" {
		return fmt.Errorf("environment is required")
	}
	validEnvs := []string{"development", "staging", "production"}
	for _, env := range validEnvs {
		if cfg.Environment == env {
			return nil
		}
	}
	return fmt.Errorf("invalid environment: %s", cfg.Environment)
}

func validateSettings(cfg *project.Config) error {
	if cfg.Settings.LogLevel == "" {
		return fmt.Errorf("log level is required")
	}
	validLevels := []string{"debug", "info", "warn", "error"}
	validLevel := false
	for _, level := range validLevels {
		if cfg.Settings.LogLevel == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		return fmt.Errorf("invalid log level: %s", cfg.Settings.LogLevel)
	}
	return nil
}

func validateDependencies(cfg *project.Config) error {
	seen := make(map[string]bool)
	for _, dep := range cfg.Dependencies {
		if seen[dep] {
			return fmt.Errorf("duplicate dependency: %s", dep)
		}
		seen[dep] = true
	}
	return nil
}

// ... other validation functions
