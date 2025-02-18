package project

import "github.com/shawnkhoffman/nix-foundry/internal/services/config"

// SettingsToMap converts a config.Settings struct to a map[string]string
func SettingsToMap(s config.Settings) map[string]string {
	return map[string]string{
		"autoUpdate":     boolToString(s.AutoUpdate),
		"updateInterval": s.UpdateInterval,
		"logLevel":       s.LogLevel,
	}
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
