package schema

import "time"

type Config struct {
	Version  string    `yaml:"version" validate:"required"`
	Kind     string    `yaml:"kind" validate:"required,oneof=project team user"`
	Metadata Metadata  `yaml:"metadata" validate:"required"`
	Base     string    `yaml:"base,omitempty"`
	Modules  []Module  `yaml:"modules,omitempty"`
	Nix      NixConfig `yaml:"nix" validate:"required"`
	Settings Settings  `yaml:"settings" validate:"required"`
}

type Metadata struct {
	Name        string `yaml:"name" validate:"required"`
	Description string `yaml:"description,omitempty"`
}

type Module struct {
	Name     string         `yaml:"name" validate:"required"`
	Type     string         `yaml:"type" validate:"required"`
	Config   map[string]any `yaml:"config,omitempty"`
	Packages []string       `yaml:"packages,omitempty"`
}

type NixConfig struct {
	Channels []Channel `yaml:"channels,omitempty"`
	Packages struct {
		Core     []string `yaml:"core" validate:"required"`
		Optional []string `yaml:"optional,omitempty"`
	} `yaml:"packages" validate:"required"`
}

type Channel struct {
	Name string `yaml:"name" validate:"required"`
	URL  string `yaml:"url" validate:"required"`
}

type Settings struct {
	AutoUpdate     bool          `yaml:"autoUpdate" validate:"boolean"`
	UpdateInterval time.Duration `yaml:"updateInterval" validate:"required"`
	LogLevel       string        `yaml:"logLevel" validate:"oneof=info debug warn"`
}
