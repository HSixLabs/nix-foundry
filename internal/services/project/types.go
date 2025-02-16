package project

import "github.com/shawnkhoffman/nix-foundry/internal/services/config"

// Config represents a project configuration
type Config struct {
	Version      string          `yaml:"version"`
	Name         string          `yaml:"name"`
	Environment  string          `yaml:"environment"`
	Settings     config.Settings `yaml:"settings"`
	Dependencies []string        `yaml:"dependencies"`
}

// LegacyConfig represents an old project configuration format
type LegacyConfig struct {
	Version      string   `yaml:"version"`
	Name         string   `yaml:"name"`
	Environment  string   `yaml:"environment"`
	Dependencies []string `yaml:"dependencies"`
}
