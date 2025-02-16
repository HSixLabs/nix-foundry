package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"gopkg.in/yaml.v3"
)

// Service defines the interface for project operations
type Service interface {
	Load() error
	Save() error
	ValidateConflicts(personal *config.Config) error
	GetProjectConfig() *ProjectConfig
	Import(path string) error
	Export(path string) error
	InitializeProject(name, team string, force bool) error
	UpdateProjectConfig(team string) error
	ImportConfig(path string) error
	ExportConfig(path string) error
	Backup(projectID string) error
	GetConfigDir() string
}

// ServiceImpl implements the project Service interface
type ServiceImpl struct {
	configService  config.Service
	envService     environment.Service
	packageService packages.Service
	projectConfig  *ProjectConfig
	logger         *logging.Logger
}

// ProjectConfig represents project-specific configuration
type ProjectConfig struct {
	Version      string          `yaml:"version"`
	Name         string          `yaml:"name"`
	Environment  string          `yaml:"environment"`
	Settings     config.Settings `yaml:"settings"`
	Dependencies []string        `yaml:"dependencies"`
}

func NewService(
	cfgSvc config.Service,
	envSvc environment.Service,
	pkgSvc packages.Service,
) Service {
	return &ServiceImpl{
		configService:  cfgSvc,
		envService:     envSvc,
		packageService: pkgSvc,
		logger:         logging.GetLogger(),
	}
}

func (s *ServiceImpl) Load() error {
	s.logger.Info("Loading project configuration")
	var cfg ProjectConfig
	if err := s.configService.LoadSection("project", &cfg); err != nil {
		s.logger.WithError(err).Error("Failed to load project configuration")
		return fmt.Errorf("failed to load project configuration section: %w", err)
	}
	s.projectConfig = &cfg
	s.logger.Debug("Project configuration loaded successfully")
	return nil
}

func (s *ServiceImpl) Save() error {
	s.logger.Info("Saving project configuration")
	if err := s.projectConfig.Validate(); err != nil {
		s.logger.WithError(err).Error("Project configuration validation failed")
		return errors.NewValidationError("project", err, "project configuration validation failed")
	}
	if err := s.configService.Save(); err != nil {
		s.logger.WithError(err).Error("Failed to save project configuration")
		return err
	}
	s.logger.Debug("Project configuration saved successfully")
	return nil
}

func (s *ServiceImpl) ValidateConflicts(personal *config.Config) error {
	if s.projectConfig == nil {
		return errors.NewValidationError("", fmt.Errorf("project config not loaded"), "project configuration must be loaded before validation")
	}

	// Check for environment conflicts
	if personal.Environment.Default != s.projectConfig.Environment {
		return errors.NewConflictError(
			fmt.Errorf("environment mismatch: personal=%s, project=%s",
				personal.Environment.Default, s.projectConfig.Environment),
			"environment settings conflict between personal and project configuration",
		)
	}

	// Check for settings conflicts
	if err := s.validateSettingsConflicts(personal.Settings); err != nil {
		return errors.NewConflictError(err, "settings conflict between personal and project configuration")
	}

	return nil
}

func (s *ServiceImpl) GetProjectConfig() *ProjectConfig {
	return s.projectConfig
}

func (s *ServiceImpl) Import(path string) error {
	s.logger.Info("Importing project configuration", "path", path)

	// Create backup before import
	if err := s.configService.CreateBackup("pre-import-backup"); err != nil {
		s.logger.WithError(err).Error("Failed to create backup before import")
		return errors.NewLoadError(path, err, "failed to create backup before import")
	}

	// Check if path is a directory
	fi, err := os.Stat(path)
	if err != nil {
		s.logger.WithError(err).Error("Failed to access path", "path", path)
		return errors.NewLoadError(path, err, "failed to access configuration path")
	}

	var configPath string
	if fi.IsDir() {
		configPath = filepath.Join(path, "nix-foundry.yaml")
		if _, err := os.Stat(configPath); err != nil {
			return errors.NewLoadError(path, err, "no nix-foundry.yaml found in directory")
		}
	} else {
		configPath = path
	}

	// Load and validate the config
	var cfg ProjectConfig
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

func (s *ServiceImpl) Export(path string) error {
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

func (s *ServiceImpl) InitializeProject(name, team string, force bool) error {
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
	projectCfg := ProjectConfig{
		Version:     "1.0",
		Name:        name,
		Environment: "default",
	}

	// Merge team config if specified
	if team != "" {
		var teamConfig config.Config
		if err := s.configService.LoadSection("team", &teamConfig); err != nil {
			return fmt.Errorf("failed to load team config: %w", err)
		}
		s.projectConfig = s.mergeTeamConfig(projectCfg, &teamConfig)
	}

	// Setup isolated environment
	if err := s.envService.SetupIsolation(true); err != nil {
		return fmt.Errorf("environment setup failed: %w", err)
	}

	// Save project config
	if err := s.configService.Save(); err != nil {
		return fmt.Errorf("failed to save project config: %w", err)
	}

	// Initialize environment packages
	if err := s.packageService.Sync(); err != nil {
		return fmt.Errorf("failed to initialize packages: %w", err)
	}

	return nil
}

func (s *ServiceImpl) mergeTeamConfig(projectCfg ProjectConfig, _ *config.Config) *ProjectConfig {
	// TODO: Implement team config merging logic
	return &projectCfg
}

func (s *ServiceImpl) UpdateProjectConfig(team string) error {
	// Load current configuration
	projectCfg := s.GetProjectConfig()
	if projectCfg == nil {
		return fmt.Errorf("no project configuration loaded")
	}

	// Load team configuration
	var teamCfg config.Config
	if err := s.configService.LoadSection("team", &teamCfg); err != nil {
		return fmt.Errorf("failed to load team config: %w", err)
	}

	// Merge configurations
	merged := mergeConfigs(projectCfg, &teamCfg)

	// Validate merged config
	if err := merged.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	// Persist updated config
	s.projectConfig = merged
	if err := s.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

func mergeConfigs(project *ProjectConfig, _ *config.Config) *ProjectConfig {
	// Implementation of config merging logic
	return project
}

func (s *ServiceImpl) ImportConfig(path string) error {
	// Implementation of ImportConfig method
	return nil
}

func (s *ServiceImpl) ExportConfig(path string) error {
	// Implementation of ExportConfig method
	return nil
}

func (s *ServiceImpl) Backup(projectID string) error {
	// Implementation of Backup method
	return nil
}

func (s *ServiceImpl) GetConfigDir() string {
	return s.configService.GetConfigDir()
}

// TODO: Implement when team permissions system is ready
// func (s *ServiceImpl) validateTeamPermissions(_ *config.Config) error {
// 	return nil
// }
