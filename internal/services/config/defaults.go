package config

import "time"

// NewDefaultConfig creates a new configuration with default values
func NewDefaultConfig() *Config {
	return &Config{
		LastUpdated: time.Now(),
		Version:     "1.0",
		Settings: Settings{
			AutoUpdate: true,
			LogLevel:   "info",
		},
		Backup: BackupSettings{
			MaxBackups:     10,
			RetentionDays:  30,
			BackupDir:      "~/.nix-foundry/backups",
			ExcludePattern: []string{".git", "node_modules"},
		},
		Environment: EnvironmentSettings{
			Default:    "development",
			AutoSwitch: true,
		},
	}
}
