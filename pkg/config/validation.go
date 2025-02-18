package config

// DELETE THESE DUPLICATE METHODS
// func (v *Validator) ValidateConfig() error {...}
// func (v *Validator) validateShell() error {...}
// func (v *Validator) validateEditor() error {...}
// func (v *Validator) validatePackages() error {...}
// func (v *Validator) validateGit() error {...}
// func (v *Validator) validateTeam() error {...}

import (
	"fmt"
	"time"
)

func (p *ProjectConfig) validateVersion() error {
	if p.Version == "" {
		return fmt.Errorf("version is required")
	}
	validVersions := []string{"1.0", "1.1", "1.2"}
	for _, v := range validVersions {
		if p.Version == v {
			return nil
		}
	}
	return fmt.Errorf("unsupported version: %s", p.Version)
}

func (p *ProjectConfig) validateName() error {
	if p.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(p.Name) > 50 {
		return fmt.Errorf("name exceeds maximum length of 50 characters")
	}
	return nil
}

func (p *ProjectConfig) validateEnvironment() error {
	if p.Environment == "" {
		return fmt.Errorf("environment is required")
	}
	validEnvs := []string{"development", "staging", "production"}
	valid := false
	for _, env := range validEnvs {
		if p.Environment == env {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid environment: %s", p.Environment)
	}
	return nil
}

func (p *ProjectConfig) validateSettings() error {
	if p.Settings == nil {
		return fmt.Errorf("settings cannot be nil")
	}

	// Validate log level if set
	if logLevel, ok := p.Settings["logLevel"]; ok {
		validLevels := []string{"debug", "info", "warn", "error"}
		valid := false
		for _, level := range validLevels {
			if logLevel == level {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid log level: %s", logLevel)
		}
	}

	// Validate update interval if auto-update is enabled
	if autoUpdate, ok := p.Settings["autoUpdate"]; ok && autoUpdate == "true" {
		if interval, ok := p.Settings["updateInterval"]; ok {
			if _, err := time.ParseDuration(interval); err != nil {
				return fmt.Errorf("invalid update interval format: %s", interval)
			}
		}
	}

	return nil
}

func (p *ProjectConfig) validateDependencies() error {
	seen := make(map[string]bool)
	for _, dep := range p.Dependencies {
		if dep == "" {
			return fmt.Errorf("empty dependency not allowed")
		}
		if seen[dep] {
			return fmt.Errorf("duplicate dependency: %s", dep)
		}
		seen[dep] = true
	}
	return nil
}
