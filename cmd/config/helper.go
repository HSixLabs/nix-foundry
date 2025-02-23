// Package config provides configuration management commands for Nix Foundry.
package config

import (
	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/shawnkhoffman/nix-foundry/service/config"
)

// getConfigService returns a new configuration service.
func getConfigService() *config.Service {
	fs := filesystem.NewOSFileSystem()
	return config.NewService(fs)
}
