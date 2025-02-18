package config

import (
	"fmt"
)

// Single source for all type declarations
type Type string

const (
	PersonalConfigType Type = "personal"
	ProjectConfigType  Type = "project"
	TeamConfigType     Type = "team"
)

// BaseConfig provides common configuration fields
type BaseConfig struct {
	Type    Type   `yaml:"type"`
	Version string `yaml:"version"`
	Name    string `yaml:"name,omitempty"`
}

// Validator defines the interface for configuration validation
type Validator interface {
	Validate() error
}

// ProjectConfig represents project-specific configuration
type ProjectConfig struct {
	BaseConfig `yaml:",inline"`
	Version    string              `yaml:"version"`
	Name       string              `yaml:"name"`
	Environment string              `yaml:"environment"`
	Dependencies []string            `yaml:"dependencies"`
	Required     []string           `yaml:"required,omitempty"`
	Settings     map[string]string  `yaml:"settings,omitempty"`
	Tools        []string           `yaml:"tools,omitempty"`
}

// Validate implements the Validator interface
func (p *ProjectConfig) Validate() error {
	if err := p.validateVersion(); err != nil {
		return fmt.Errorf("version validation failed: %w", err)
	}
	if err := p.validateName(); err != nil {
		return fmt.Errorf("name validation failed: %w", err)
	}
	if err := p.validateEnvironment(); err != nil {
		return fmt.Errorf("environment validation failed: %w", err)
	}
	if err := p.validateSettings(); err != nil {
		return fmt.Errorf("settings validation failed: %w", err)
	}
	if err := p.validateDependencies(); err != nil {
		return fmt.Errorf("dependencies validation failed: %w", err)
	}
	return nil
}

// ConfigPaths stores standard configuration paths
type Paths struct {
	Personal string
	Project  string
	Team     string
	Current  string
}

type EditorConfig struct {
	Type       string                 `yaml:"type"`
	Extensions []string               `yaml:"extensions,omitempty"`
	Settings   map[string]interface{} `yaml:"settings,omitempty"`
}

type GitConfig struct {
	Enable bool `yaml:"enable"`
	User   struct {
		Name  string `yaml:"name"`
		Email string `yaml:"email"`
	} `yaml:"user,omitempty"`
	Config map[string]string `yaml:"config,omitempty"`
}

type ShellConfig struct {
	Type    string   `yaml:"type"`
	Plugins []string `yaml:"plugins,omitempty"`
}

type PackagesConfig struct {
	Additional       []string            `yaml:"additional"`
	PlatformSpecific map[string][]string `yaml:"platformSpecific"`
	Development      []string            `yaml:"development"`
	Team             map[string][]string `yaml:"team"`
}

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
		Go struct {
			Version  string   `yaml:"version,omitempty"`
			Packages []string `yaml:"packages,omitempty"`
		} `yaml:"go"`
		Node struct {
			Version  string   `yaml:"version"`
			Packages []string `yaml:"packages,omitempty"`
		} `yaml:"node"`
		Python struct {
			Version  string   `yaml:"version"`
			Packages []string `yaml:"packages,omitempty"`
		} `yaml:"python"`
	} `yaml:"languages"`
	Tools []string `yaml:"tools,omitempty"`
}

// NixConfig represents the internal configuration structure
type NixConfig struct {
	Version     string            `yaml:"version"`
	Shell       ShellConfig       `yaml:"shell"`
	Editor      EditorConfig      `yaml:"editor"`
	Git         GitConfig         `yaml:"git"`
	Packages    PackagesConfig    `yaml:"packages"`
	Team        TeamConfig        `yaml:"team"`
	Platform    PlatformConfig    `yaml:"platform"`
	Development DevelopmentConfig `yaml:"development"`
}
