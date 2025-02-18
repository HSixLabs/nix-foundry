package configtypes

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/types"
)

// Replace local Config with type alias to core types.Config
type (
	Config = types.Config
	ProjectConfig = types.ProjectConfig
	NixConfig = types.NixConfig
	Settings = types.Settings
	EnvironmentSettings = types.EnvironmentSettings
)

// Update validation methods to use core types
func ValidateVersion(cfg *types.Config) error {
	if cfg.Version == "" {
		return fmt.Errorf("version is required")
	}
	return nil
}

func ValidateName(cfg *types.Config) error {
	if cfg.Project.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

// Remove duplicate Service interface declaration from this file

// Add Service interface definition
type Service interface {
	Initialize(testMode bool) error
	Load() (*types.Config, error)
	Save(cfg *types.Config) error
	SaveConfig(cfg *types.Config) error
	Apply(config *types.NixConfig, testMode bool) error
	GetConfig() *types.NixConfig
	GetConfigDir() string
	GetBackupDir() string
	LoadSection(name string, v interface{}) error
	SaveSection(name string, v interface{}) error
	ApplyFlags(flags map[string]string, force bool) error
	GenerateInitialConfig(shell, editor, gitName, gitEmail string) (*types.Config, error)
	PreviewConfiguration(*types.Config) error
	ConfigExists() bool
}
