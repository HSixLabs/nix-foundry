package project

import (
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/types"
	servicetypes "github.com/shawnkhoffman/nix-foundry/internal/services/config/types"
)

func ToServiceConfig(pkg *types.ProjectConfig) *servicetypes.ProjectConfig {
	return &servicetypes.ProjectConfig{
		Version:      pkg.Version,
		Name:         pkg.Name,
		Environment:  pkg.Environment,
		Settings:     pkg.Settings,
		Dependencies: pkg.Dependencies,
	}
}

func ToPkgConfig(service *servicetypes.ProjectConfig) *types.ProjectConfig {
	return &types.ProjectConfig{
		Version:      service.Version,
		Name:         service.Name,
		Environment:  service.Environment,
		Settings:     service.Settings,
		Dependencies: service.Dependencies,
	}
}

func convertSettings(settings map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range settings {
		result[k] = v
	}
	return result
}
