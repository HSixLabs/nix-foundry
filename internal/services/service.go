package services

import (
	"fmt"
	"os"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
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

	projectCfg := config.ProjectConfig{
		BaseConfig: config.BaseConfig{
			Type:    config.ProjectConfigType,
			Version: "1.0",
			Name:    projectName,
		},
		Required: []string{"git"},
	}

	if teamName != "" {
		teamConfig, err := s.configManager.LoadConfig(config.TeamConfigType, teamName)
		if err != nil {
			return fmt.Errorf("failed to load team configuration: %w", err)
		}
		teamProjectConfig, ok := teamConfig.(config.ProjectConfig)
		if !ok {
			return fmt.Errorf("invalid team configuration type")
		}
		projectCfg = s.configManager.MergeProjectConfigs(projectCfg, teamProjectConfig)
	}

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
func (s *PackageService) ListCustomPackages() ([]string, error) {
	return s.configManager.LoadCustomPackages()
}

// AddPackages adds new packages to the configuration
func (s *PackageService) AddPackages(packages []string) error {
	existing, err := s.configManager.LoadCustomPackages()
	if err != nil {
		return fmt.Errorf("failed to load existing packages: %w", err)
	}

	// Create a map to deduplicate packages
	packageMap := make(map[string]bool)
	for _, pkg := range existing {
		packageMap[pkg] = true
	}
	for _, pkg := range packages {
		packageMap[pkg] = true
	}

	// Convert back to slice
	var finalPackages []string
	for pkg := range packageMap {
		finalPackages = append(finalPackages, pkg)
	}

	if err := s.configManager.SaveCustomPackages(finalPackages); err != nil {
		return fmt.Errorf("failed to save packages: %w", err)
	}

	return nil
}

// RemovePackages removes packages from the configuration
func (s *PackageService) RemovePackages(packages []string) error {
	existing, err := s.configManager.LoadCustomPackages()
	if err != nil {
		return fmt.Errorf("failed to load existing packages: %w", err)
	}

	// Create a map of packages to remove
	toRemove := make(map[string]bool)
	for _, pkg := range packages {
		toRemove[pkg] = true
	}

	// Filter out removed packages
	var remaining []string
	for _, pkg := range existing {
		if !toRemove[pkg] {
			remaining = append(remaining, pkg)
		}
	}

	if err := s.configManager.SaveCustomPackages(remaining); err != nil {
		return fmt.Errorf("failed to save packages: %w", err)
	}

	return nil
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
