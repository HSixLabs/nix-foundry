package config

import (
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/types"
)

func NewDefaultConfig() *types.Config {
	return &types.Config{
		Version: "1.0",
		Project: types.ProjectConfig{
			Environment: "development",
			Settings:    make(map[string]string),
		},
		Settings: types.Settings{
			AutoUpdate:     true,
			UpdateInterval: "24h",
			LogLevel:       "info",
		},
	}
}
