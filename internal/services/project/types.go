package project

import (
	"fmt"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/types"
	configtypes "github.com/shawnkhoffman/nix-foundry/internal/services/config/types"
)

// ProjectConfig represents the project configuration
type ProjectConfig struct {
	Version      string            `yaml:"version"`
	Name         string            `yaml:"name"`
	Environment  string            `yaml:"environment"`
	Dependencies []string          `yaml:"dependencies"`
	Settings     map[string]string `yaml:"settings"`
	Tools        []string          `yaml:"tools"`
	Required     []string          `yaml:"required"`
}

// DependencyList returns the list of project dependencies
func (p *ProjectConfig) DependencyList() []string {
	if p == nil {
		return nil
	}
	return p.Dependencies
}

// Config represents a project configuration
type Config struct {
	Project configtypes.ProjectConfig
}

// Add conversion methods
func NewConfig(base configtypes.Config) *configtypes.Config {
	return &base
}

func (c *Config) ToBaseConfig() types.ProjectConfig {
	return c.Project
}

// Update Service interface to match ProjectService
type Service interface {
	Load() error
	Save() error
	ValidateConflicts(cfg *configtypes.Config) error
	GetProjectConfig() *configtypes.Config
	Import(path string) error
	Export(path string) error
	InitializeProject(name, team string, force bool) error
	UpdateProjectConfig(team string) error
	ImportConfig(path string) error
	ExportConfig(path string) error
	Backup(projectID string) error
	GetConfigDir() string
}

// NewDefaultConfig creates a new default project configuration
func NewDefaultConfig() ProjectConfig {
	return ProjectConfig{
		Version:     "1.0",
		Environment: "development",
		Settings:    make(map[string]string),
		Tools:       []string{},
		Required:    []string{},
	}
}

// Update the validation methods to use the local Config type
func (c *Config) Validate() error {
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

func (c *Config) ValidateVersion() error {
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

func (c *Config) ValidateName() error {
	if c.Project.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(c.Project.Name) > 50 {
		return fmt.Errorf("name exceeds maximum length of 50 characters")
	}
	return nil
}

func (c *Config) ValidateEnvironment() error {
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

func (c *Config) ValidateSettings() error {
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

	if c.Project.Settings["autoUpdate"] == "true" && c.Project.Settings["updateInterval"] != "" {
		if _, err := time.ParseDuration(c.Project.Settings["updateInterval"]); err != nil {
			return fmt.Errorf("invalid update interval format: %s", c.Project.Settings["updateInterval"])
		}
	}

	return nil
}

func (c *Config) ValidateDependencies() error {
	seen := make(map[string]bool)
	for _, dep := range c.Project.Dependencies {
		if seen[dep] {
			return fmt.Errorf("duplicate dependency: %s", dep)
		}
		seen[dep] = true
	}
	return nil
}

// ToProjectConfig converts Config to types.ProjectConfig
func (c *Config) ToProjectConfig() types.ProjectConfig {
	if c == nil {
		return types.ProjectConfig{}
	}
	return types.ProjectConfig{
		Version:      c.Project.Version,
		Name:         c.Project.Name,
		Environment:  c.Project.Environment,
		Dependencies: c.Project.Dependencies,
		Settings:     c.Project.Settings,
	}
}

// FromProjectConfig creates a Config from types.ProjectConfig
func FromProjectConfig(p types.ProjectConfig) *Config {
	return &Config{Project: p}
}
