package services

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/types"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
)

// ConfigurationService handles business logic for configuration management
type ConfigurationService struct {
	configManager config.Service
}

// NewConfigurationService creates a new configuration service
func NewConfigurationService() (*ConfigurationService, error) {
	manager := config.NewService()

	return &ConfigurationService{
		configManager: manager,
	}, nil
}

// InitializeProject sets up a new project configuration
func (s *ConfigurationService) InitializeProject(projectName, teamName string, force bool) error {
	if _, err := os.Stat(".nix-foundry.yaml"); err == nil && !force {
		return fmt.Errorf("project configuration already exists. Use --force to override")
	}

	// Create base project config using types.ProjectConfig
	projectCfg := types.ProjectConfig{
		Version: "1.0",
		Name:    projectName,
		Required: []string{"git"},
	}

	if teamName != "" {
		teamConfig, err := s.configManager.LoadConfig(config.TeamType, teamName)
		if err != nil {
			return fmt.Errorf("failed to load team configuration: %w", err)
		}

		// Convert team config to correct type
		teamProjectConfig, ok := teamConfig.(types.ProjectConfig)
		if !ok {
			return fmt.Errorf("invalid team configuration type")
		}

		// Merge configs using the correct types
		projectCfg = s.configManager.MergeProjectConfigs(projectCfg, teamProjectConfig)
	}

	// Write the config using the merged project config
	if err := s.configManager.WriteConfig(".nix-foundry.yaml", projectCfg); err != nil {
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	return nil
}

// UpdateProjectConfig updates the project configuration with team settings
func (s *ConfigurationService) UpdateProjectConfig(teamName string) error {
	projectConfig, err := s.configManager.LoadProjectWithTeam("", teamName)
	if err != nil {
		return fmt.Errorf("failed to load project configuration: %w", err)
	}

	if err := s.configManager.WriteConfig(".nix-foundry.yaml", projectConfig); err != nil {
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	return nil
}

// PackageService handles package management operations
type PackageService struct {
	configManager config.Service
}

// NewPackageService creates a new package service
func NewPackageService() (*PackageService, error) {
	manager := config.NewService()
	return &PackageService{
		configManager: manager,
	}, nil
}

// ListCustomPackages returns the list of custom packages
func (s *PackageService) ListCustomPackages(ctx context.Context) ([]string, error) {
	cfg, err := s.configManager.Load()
	if err != nil {
		return nil, err
	}
	return append(cfg.Packages.User, cfg.Packages.Team...), nil
}

// AddPackages adds new packages to the configuration
func (s *PackageService) AddPackages(ctx context.Context, packages []string) error {
	cfg, err := s.configManager.Load()
	if err != nil {
		return err
	}

	existing := make(map[string]bool)
	for _, pkg := range cfg.Packages.User {
		existing[pkg] = true
	}

	var added []string
	for _, pkg := range packages {
		if !existing[pkg] {
			cfg.Packages.User = append(cfg.Packages.User, pkg)
			added = append(added, pkg)
		}
	}

	if len(added) > 0 {
		// Create a new config with only the necessary fields
		configToSave := &types.Config{
			LastUpdated: time.Now(),
			Packages:    cfg.Packages,
			Project:     cfg.Project,
			NixConfig:   cfg.NixConfig,
			Settings:    cfg.Settings,
			Environment: cfg.Environment,
		}
		return s.configManager.Save(configToSave)
	}
	return nil
}

// RemovePackages removes packages from the configuration
func (s *PackageService) RemovePackages(ctx context.Context, packages []string) error {
	cfg, err := s.configManager.Load()
	if err != nil {
		return err
	}

	packageMap := make(map[string]bool)
	for _, pkg := range packages {
		packageMap[pkg] = true
	}

	var filtered []string
	for _, pkg := range cfg.Packages.User {
		if !packageMap[pkg] {
			filtered = append(filtered, pkg)
		}
	}

	cfg.Packages.User = filtered

	// Create a new config with only the necessary fields
	configToSave := &types.Config{
		LastUpdated: time.Now(),
		Packages:    cfg.Packages,
		Project:     cfg.Project,
		NixConfig:   cfg.NixConfig,
		Settings:    cfg.Settings,
		Environment: cfg.Environment,
	}
	return s.configManager.Save(configToSave)
}

// Add service-layer functionality
// This would include business logic that coordinates between packages

// Add to ConfigurationService
func (s *ConfigurationService) ApplyConfiguration(testMode bool) error {
	var nixConfig config.NixConfig
	if err := s.configManager.ReadConfig("config.yaml", &nixConfig); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := s.configManager.Apply(&nixConfig, testMode); err != nil {
		return fmt.Errorf("failed to apply configuration: %w", err)
	}

	return nil
}

// Remove the standalone functions and add them as methods to ConfigurationService
func (s *ConfigurationService) LoadProjectConfig(name string) (*project.Config, error) {
	cfg, err := s.configManager.LoadConfig(config.ProjectType, name)
	if err != nil {
		return nil, fmt.Errorf("failed to load project config: %w", err)
	}

	projectCfg, ok := cfg.(*project.Config)
	if !ok {
		return nil, fmt.Errorf("invalid configuration type")
	}

	return projectCfg, nil
}

func (s *ConfigurationService) LoadTeamConfig(name string) (*config.TeamConfig, error) {
	cfg, err := s.configManager.LoadConfig(config.TeamType, name)
	if err != nil {
		return nil, fmt.Errorf("failed to load team config: %w", err)
	}

	teamCfg, ok := cfg.(*config.TeamConfig)
	if !ok {
		return nil, fmt.Errorf("invalid team configuration type")
	}

	return teamCfg, nil
}

type ServiceImpl struct {
	configManager config.Service
	// ... other fields
}

// Update the conversion helper to properly map all fields
func convertConfig(cfg *config.Config) *types.Config {
	if cfg == nil {
		return nil
	}

	return &types.Config{
		LastUpdated: time.Now(),
		NixConfig:   &types.NixConfig{
			Version:     cfg.NixConfig.Version,
			Settings:    cfg.NixConfig.Settings,
			Shell:       cfg.NixConfig.Shell,
			Editor:      cfg.NixConfig.Editor,
			Git:         cfg.NixConfig.Git,
			Packages:    cfg.NixConfig.Packages,
			Team:        cfg.NixConfig.Team,
			Platform:    cfg.NixConfig.Platform,
			Development: cfg.NixConfig.Development,
		},
		Project:     cfg.Project,
		Packages:    cfg.NixConfig.Packages,
		Settings:    types.Settings{
			AutoUpdate:     cfg.NixConfig.Settings.AutoUpdate,
			UpdateInterval: cfg.NixConfig.Settings.UpdateInterval,
			LogLevel:       cfg.NixConfig.Settings.LogLevel,
		},
		Environment: types.EnvironmentSettings{
			// Map the environment settings fields as needed
			// Add the appropriate fields based on your EnvironmentSettings struct
		},
		Shell:       cfg.NixConfig.Shell,
		Editor:      cfg.NixConfig.Editor,
		Git:         cfg.NixConfig.Git,
		Dependencies: cfg.Project.Dependencies,
	}
}

// Update the ServiceImpl methods to use the proper types
func (s *ServiceImpl) Save(cfg *config.Config) error {
	converted := convertConfig(cfg)
	return s.configManager.Save(converted)
}

func (s *ServiceImpl) SaveConfig(cfg *config.Config) error {
	converted := convertConfig(cfg)
	return s.configManager.Save(converted)
}
