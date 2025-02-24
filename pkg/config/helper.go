/*
Package config provides configuration management functionality for Nix Foundry.
It handles loading, saving, and applying configuration settings across different
scopes including user, team, and project configurations.
*/
package config

import "github.com/shawnkhoffman/nix-foundry/pkg/filesystem"

/*
GetConfigService creates and returns a new configuration service instance.
It initializes the service with a default OS filesystem implementation,
enabling configuration operations on the local system.
*/
func GetConfigService() *Service {
	fs := filesystem.NewOSFileSystem()
	return NewService(fs)
}
