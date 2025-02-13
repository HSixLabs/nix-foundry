package config

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

type UserConfig struct {
	Extends  string         `yaml:"extends,omitempty"`
	Shell    string         `yaml:"shell"`
	Editors  []EditorConfig `yaml:"editors,omitempty"`
	Packages []string       `yaml:"packages,omitempty"`
	Git      GitConfig      `yaml:"git"`
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
		} `yaml:"go,omitempty"`
		Node struct {
			Version  string   `yaml:"version,omitempty"`
			Packages []string `yaml:"packages,omitempty"`
		} `yaml:"node,omitempty"`
		Python struct {
			Version  string   `yaml:"version,omitempty"`
			Packages []string `yaml:"packages,omitempty"`
		} `yaml:"python,omitempty"`
	} `yaml:"languages,omitempty"`
}
