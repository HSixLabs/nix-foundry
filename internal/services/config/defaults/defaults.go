package defaults

import (
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/types"
)

func New() *types.Config {
    return &types.Config{
        NixConfig: &types.NixConfig{
            Settings: types.Settings{
                LogLevel:       "info",
                AutoUpdate:     true,
                UpdateInterval: "24h",
            },
        },
        Environment: types.EnvironmentSettings{
            Name:       "default",
            Default:    "development",
            AutoSwitch: true,
        },
        Project: types.ProjectConfig{
            Version:     "1.0",
            Environment: "development",
            Settings:    make(map[string]string),
        },
    }
}
