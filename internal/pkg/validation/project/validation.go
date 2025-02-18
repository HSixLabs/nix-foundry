package project

import (
	"errors"
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/types"
	configtypes "github.com/shawnkhoffman/nix-foundry/internal/services/config/types"
)

// ValidateConfig validates a project configuration
func ValidateConfig(cfg *configtypes.Config) error {
	if err := validateVersion(cfg.Project.Version); err != nil {
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

func Validate(cfg *configtypes.Config) error {
	if err := validateVersion(cfg.Project.Version); err != nil {
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

func ValidateProjectConfig(cfg *types.ProjectConfig) error {
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

func validateName(cfg *configtypes.Config) error {
	if cfg.Project.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(cfg.Project.Name) > 50 {
		return fmt.Errorf("name exceeds maximum length of 50 characters")
	}
	return nil
}

func validateEnvironment(cfg *configtypes.Config) error {
	if cfg.Project.Environment == "" {
		return fmt.Errorf("environment is required")
	}
	validEnvs := []string{"development", "staging", "production"}
	for _, env := range validEnvs {
		if cfg.Project.Environment == env {
			return nil
		}
	}
	return fmt.Errorf("invalid environment: %s", cfg.Project.Environment)
}

func validateSettings(cfg *configtypes.Config) error {
	logLevel, exists := cfg.Project.Settings["logLevel"]
	if !exists || logLevel == "" {
		return fmt.Errorf("log level is required")
	}

	validLevels := []string{"debug", "info", "warn", "error"}
	validLevel := false
	for _, level := range validLevels {
		if logLevel == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		return fmt.Errorf("invalid log level: %s", logLevel)
	}
	return nil
}

func validateDependencies(cfg *configtypes.Config) error {
	seen := make(map[string]bool)
	for _, dep := range cfg.Project.Dependencies {
		if seen[dep] {
			return fmt.Errorf("duplicate dependency: %s", dep)
		}
		seen[dep] = true
	}
	return nil
}

func validateVersion(version string) error {
	if version == "" {
		return errors.New("version is required")
	}

	// if !semver.IsValid(version) {
	// 	return fmt.Errorf("invalid semantic version format: %s", version)
	// }
	return nil
}

// ... other validation functions
