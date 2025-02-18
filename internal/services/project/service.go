package project

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/types"
	configservice "github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"gopkg.in/yaml.v3"
)

// ProjectService defines the interface for project operations
type ProjectService interface {
	Load() error
	Save() error
	ValidateConflicts(cfg *types.Config) error
	GetProjectConfig() *types.Config
	Import(path string) error
	Export(path string) error
	InitializeProject(name, team string, force bool) error
	UpdateProjectConfig(team string) error
	ImportConfig(path string) error
	ExportConfig(path string) error
	Backup(projectID string) error
	GetConfigDir() string
}

// ProjectServiceImpl implements the project Service interface
type ProjectServiceImpl struct {
	configService  configservice.Service
	envService     environment.Service
	packageService packages.Service
	projectConfig  *types.Config
	logger         *logging.Logger
}

// NewService creates a new project service instance
func NewService(
	configService configservice.Service,
	envService environment.Service,
	pkgService packages.Service,
) ProjectService {
	return &ProjectServiceImpl{
		configService:  configService,
		envService:     envService,
		packageService: pkgService,
		logger:         logging.GetLogger(),
	}
}

func (s *ProjectServiceImpl) Load() error {
	cfg, err := s.configService.Load()
	if err != nil {
		return err
	}
	s.projectConfig = cfg
	return nil
}

func (s *ProjectServiceImpl) Save() error {
	if s.projectConfig == nil {
		return fmt.Errorf("no configuration to save")
	}
	return s.configService.Save(s.projectConfig)
}

func (s *ProjectServiceImpl) ValidateConflicts(cfg *types.Config) error {
	if s.projectConfig == nil {
		return fmt.Errorf("no project configuration loaded")
	}

	if cfg == nil {
		return fmt.Errorf("personal config is nil")
	}

	// Add your validation logic here
	// Example:
	// Compare configurations and check for conflicts
	// between personal.Project and s.projectConfig

	return nil
}

func (s *ProjectServiceImpl) GetProjectConfig() *types.Config {
	return s.projectConfig
}

func (s *ProjectServiceImpl) Import(path string) error {
	s.logger.Info("Importing project configuration", "path", path)

	// Create backup before import
	if err := s.configService.CreateBackup("pre-import-backup"); err != nil {
		s.logger.WithError(err).Error("Failed to create backup before import")
		return errors.NewLoadError(path, err, "failed to create backup before import")
	}

	// Check if path is a directory
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.NewNotFoundError(err, "configuration file not found")
		}
		s.logger.WithError(err).Error("Failed to access path", "path", path)
		return errors.NewLoadError(path, err, "failed to access configuration path")
	}

	var configPath string
	if fi.IsDir() {
		configPath = filepath.Join(path, "nix-foundry.yaml")
		if _, statErr := os.Stat(configPath); statErr != nil {
			return errors.NewLoadError(path, statErr, "no nix-foundry.yaml found in directory")
		}
	} else {
		configPath = path
	}

	// Load and validate the config
	var cfg types.Config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return errors.NewLoadError(configPath, err, "failed to read config file")
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return errors.NewLoadError(configPath, err, "failed to parse config file")
	}

	// Validate the imported config
	if err := cfg.Validate(); err != nil {
		return errors.NewValidationError(configPath, err, "imported configuration is invalid")
	}

	s.projectConfig = &cfg
	return s.Save()
}

func (s *ProjectServiceImpl) Export(path string) error {
	s.logger.Info("Exporting project configuration", "path", path)

	if s.projectConfig == nil {
		return errors.NewValidationError(path, fmt.Errorf("no configuration loaded"), "cannot export nil configuration")
	}

	// Validate before export
	if err := s.projectConfig.Validate(); err != nil {
		return errors.NewValidationError(path, err, "cannot export invalid configuration")
	}

	data, err := yaml.Marshal(s.projectConfig)
	if err != nil {
		return errors.NewLoadError(path, err, "failed to marshal configuration")
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return errors.NewLoadError(path, err, "failed to write configuration file")
	}

	return nil
}

func (s *ProjectServiceImpl) InitializeProject(name, team string, force bool) error {
	// Validate environment first
	if err := s.envService.CheckPrerequisites(false); err != nil {
		return fmt.Errorf("environment validation failed: %w", err)
	}

	s.logger.Info("Initializing project", "name", name, "team", team)

	// Check existing config
	if _, err := os.Stat(".nix-foundry.yaml"); err == nil && !force {
		return fmt.Errorf("project configuration already exists, use --force to overwrite")
	}

	// Create base config
	s.projectConfig = &types.Config{
		LastUpdated: time.Now(),
		Project: types.ProjectConfig{
			Version:      "1.0",
			Name:         name,
			Environment:  "default",
			Settings:     make(map[string]string),
			Dependencies: []string{},
		},
		Settings: types.Settings{
			AutoUpdate:     false,
			UpdateInterval: "24h",
			LogLevel:       "info",
		},
	}

	// Merge team config if specified
	if team != "" {
		var teamConfig configservice.Config
		if err := s.configService.LoadSection("team", &teamConfig); err != nil {
			return fmt.Errorf("failed to load team config: %w", err)
		}
		s.projectConfig = s.mergeTeamConfig(*s.projectConfig, &teamConfig)
	}

	// Setup isolated environment
	if err := s.envService.SetupIsolation(true, false); err != nil {
		return fmt.Errorf("environment setup failed: %w", err)
	}

	// Save project config
	if err := s.configService.Save(s.projectConfig); err != nil {
		return fmt.Errorf("failed to save project config: %w", err)
	}

	// Initialize environment packages
	if err := s.packageService.Sync(); err != nil {
		return fmt.Errorf("failed to initialize packages: %w", err)
	}

	return nil
}

func (s *ProjectServiceImpl) mergeTeamConfig(projectCfg types.Config, _ *configservice.Config) *types.Config {
	// TODO: Implement team config merging logic
	return &projectCfg
}

func (s *ProjectServiceImpl) UpdateProjectConfig(team string) error {
	// Load current configuration
	projectCfg := s.GetProjectConfig()
	if projectCfg == nil {
		return fmt.Errorf("no project configuration loaded")
	}

	// Validate environment and packages
	if err := s.ValidateEnvironment(); err != nil {
		return fmt.Errorf("environment validation failed: %w", err)
	}

	if err := s.ValidatePackages(); err != nil {
		return fmt.Errorf("package validation failed: %w", err)
	}

	// Load team configuration
	var teamCfg configservice.Config
	if err := s.configService.LoadSection("team", &teamCfg); err != nil {
		return fmt.Errorf("failed to load team config: %w", err)
	}

	// Merge configurations - dereference the pointer for mergeConfigs
	merged := mergeConfigs(*projectCfg, &teamCfg)

	// Validate merged config
	if err := validateConfig(&merged); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Persist updated config
	s.projectConfig = &merged
	if err := s.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// Update mergeConfigs to handle the correct types
func mergeConfigs(project types.Config, team *configservice.Config) types.Config {
	result := project

	if team != nil {
		// Merge settings properly using struct fields
		if !result.Settings.AutoUpdate {
			result.Settings.AutoUpdate = team.Settings.AutoUpdate
		}
		if result.Settings.UpdateInterval == "" {
			result.Settings.UpdateInterval = team.Settings.UpdateInterval
		}
		if result.Settings.LogLevel == "" {
			result.Settings.LogLevel = team.Settings.LogLevel
		}

		// Merge other team-specific fields
		if result.Project.Environment == "" {
			result.Project.Environment = team.Project.Environment
		}

		// Merge dependencies
		existingDeps := make(map[string]bool)
		for _, dep := range result.Project.Dependencies {
			existingDeps[dep] = true
		}

		// Add team dependencies that don't already exist
		for _, dep := range team.Project.Dependencies {
			if !existingDeps[dep] {
				result.Project.Dependencies = append(result.Project.Dependencies, dep)
			}
		}

		// Merge other fields as needed
		if result.Project.Version == "" {
			result.Project.Version = team.Project.Version
		}
	}

	return result
}

func (s *ProjectServiceImpl) ImportConfig(path string) error {
	// Implementation of ImportConfig method
	return nil
}

func (s *ProjectServiceImpl) ExportConfig(path string) error {
	// Implementation of ExportConfig method
	return nil
}

func (s *ProjectServiceImpl) Backup(projectID string) error {
	// Implementation of Backup method
	return nil
}

func (s *ProjectServiceImpl) GetConfigDir() string {
	return s.configService.GetConfigDir()
}

// TODO: Implement when team permissions system is ready
// func (s *ProjectServiceImpl) validateTeamPermissions(_ *config.Config) error {
// 	return nil
// }

// Remove the method definition on types.Config and create a standalone function
func validateConfig(c *types.Config) error {
	if err := c.Project.ValidateVersion(); err != nil {
		return fmt.Errorf("version validation failed: %w", err)
	}

	if err := c.Project.ValidateName(); err != nil {
		return fmt.Errorf("name validation failed: %w", err)
	}

	if err := c.Project.ValidateEnvironment(); err != nil {
		return fmt.Errorf("environment validation failed: %w", err)
	}

	if err := c.Project.ValidateSettings(); err != nil {
		return fmt.Errorf("settings validation failed: %w", err)
	}

	if err := c.Project.ValidateDependencies(); err != nil {
		return fmt.Errorf("dependencies validation failed: %w", err)
	}

	return nil
}

func (s *ProjectServiceImpl) ValidateEnvironment() error {
	if s.envService == nil {
		return fmt.Errorf("environment service not initialized")
	}
	return s.envService.Validate()
}

func (s *ProjectServiceImpl) ValidatePackages() error {
	if s.packageService == nil {
		return fmt.Errorf("package service not initialized")
	}
	return s.packageService.Validate()
}

// Update the helper method to work with Settings struct
func (s *ProjectServiceImpl) validateSettingsConflicts(settings map[string]string) error {
	if s.projectConfig == nil {
		return fmt.Errorf("project config not loaded")
	}

	// Compare critical settings using struct fields
	if settings["environment"] != s.projectConfig.Project.Environment {
		return fmt.Errorf("setting 'environment' mismatch: project=%s, personal=%s",
			s.projectConfig.Project.Environment, settings["environment"])
	}

	if settings["logLevel"] != s.projectConfig.Settings.LogLevel {
		return fmt.Errorf("setting 'logLevel' mismatch: project=%s, personal=%s",
			s.projectConfig.Settings.LogLevel, settings["logLevel"])
	}

	// Convert autoUpdate to string for comparison
	projectAutoUpdate := fmt.Sprintf("%v", s.projectConfig.Settings.AutoUpdate)
	if settings["autoUpdate"] != projectAutoUpdate {
		return fmt.Errorf("setting 'autoUpdate' mismatch: project=%s, personal=%s",
			projectAutoUpdate, settings["autoUpdate"])
	}

	return nil
}

// Make sure ProjectServiceImpl implements both interfaces
var (
	_ Service = (*ProjectServiceImpl)(nil)
	_ ProjectService = (*ProjectServiceImpl)(nil)
)
