package types

// NewDefaultProjectConfig creates a default project configuration
func NewDefaultProjectConfig() ProjectConfig {
    return ProjectConfig{
        Version:      "1.0",
        Name:         "default",
        Environment:  "development",
        Settings:     make(map[string]string),
        Dependencies: []string{},
    }
}
