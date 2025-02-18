package config

import (
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
)

type Config struct {
	Shell     ShellConfig  `yaml:"shell"`
	Editor    EditorConfig `yaml:"editor"`
	Git       GitConfig    `yaml:"git"`
	Packages  PackageConfig `yaml:"packages"`
	Project   project.Config `yaml:"project"`
}

type PackageConfig struct {
	Core []string `yaml:"core"`
	User []string `yaml:"user"`
	Team []string `yaml:"team"`
}

type ShellConfig struct {
	Type     string `yaml:"type"`
	InitFile string `yaml:"initFile"`
}

type EditorConfig struct {
	Default string `yaml:"default"`
	// ... other fields ...
}

type GitConfig struct {
	UserName  string `yaml:"userName"`
	UserEmail string `yaml:"userEmail"`
}
