package config

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

// ProjectConfig represents project-specific configuration
type ProjectConfig struct {
	BaseConfig `yaml:",inline"`
	Required   []string `yaml:"required,omitempty"`
	Tools      struct {
		Go     []string `yaml:"go,omitempty"`
		Node   []string `yaml:"node,omitempty"`
		Python []string `yaml:"python,omitempty"`
	} `yaml:"tools,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
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
