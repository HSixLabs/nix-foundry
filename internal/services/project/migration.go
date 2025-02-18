package project

import (
	"fmt"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	config "github.com/shawnkhoffman/nix-foundry/internal/services/config"
	configtypes "github.com/shawnkhoffman/nix-foundry/internal/services/config/types"
)

// MigrationFunc defines a function that migrates configuration from one version to another
type MigrationFunc func(*configtypes.Config) error

// migrationMap stores available migrations between versions
var migrationMap = map[string]MigrationFunc{
	"1.0->1.1": migrateV1_0ToV1_1,
	"1.1->1.2": migrateV1_1ToV1_2,
}

// ServiceImpl implements the Service interface
type ServiceImpl struct {
	configService  config.Service
	projectConfig  *configtypes.Config
	logger         *logging.Logger
}

// Migrate updates the configuration to the latest version
func (s *ServiceImpl) Migrate() error {
	if s.projectConfig == nil {
		return errors.NewValidationError("", fmt.Errorf("no configuration loaded"), "cannot migrate nil configuration")
	}

	currentVersion := s.projectConfig.Version
	latestVersion := "1.2" // Should be maintained with each release

	if currentVersion == latestVersion {
		s.logger.Debug("Configuration already at latest version")
		return nil
	}

	s.logger.Info("Starting configuration migration",
		"from_version", currentVersion,
		"to_version", latestVersion)

	// Add this section where appropriate
	if s.projectConfig.Project.Version == "1.0" {
		legacyConfig := &LegacyConfig{
			Name: s.projectConfig.Project.Name,
			// ... other fields from ProjectConfig ...
		}

		newConfig := migrateLegacyConfig(legacyConfig)

		s.projectConfig = newConfig
	}

	// Apply migrations in sequence
	for currentVersion != latestVersion {
		nextVersion := getNextVersion(currentVersion)
		migrationKey := fmt.Sprintf("%s->%s", currentVersion, nextVersion)

		migration, exists := migrationMap[migrationKey]
		if !exists {
			return fmt.Errorf("no migration path from %s to %s", currentVersion, nextVersion)
		}

		s.logger.Debug("Applying migration", "from", currentVersion, "to", nextVersion)
		if err := migration(s.projectConfig); err != nil {
			return fmt.Errorf("migration from %s to %s failed: %w", currentVersion, nextVersion, err)
		}

		currentVersion = nextVersion
		s.projectConfig.Version = currentVersion
	}

	return s.Save()
}

// Example migration function
func migrateV1_0ToV1_1(cfg *configtypes.Config) error {
	if cfg.Project.Dependencies == nil {
		cfg.Project.Dependencies = []string{}
	}
	// Add new fields with default values
	if cfg.Settings.LogLevel == "" {
		cfg.Settings.LogLevel = "info"
	}

	// Convert legacy fields
	for i, dep := range cfg.Project.Dependencies {
		cfg.Project.Dependencies[i] = strings.TrimSpace(dep)
	}

	return nil
}

func migrateV1_1ToV1_2(cfg *configtypes.Config) error {
	// Example: Add new required settings
	if cfg.Settings.AutoUpdate {
		cfg.Settings.UpdateInterval = "24h"
	}
	return nil
}

// getNextVersion determines the next version in the migration path
func getNextVersion(current string) string {
	switch current {
	case "1.0":
		return "1.1"
	case "1.1":
		return "1.2"
	default:
		return current
	}
}

// Use local Config type
func migrateLegacyConfig(legacy *LegacyConfig) *configtypes.Config {
	return &configtypes.Config{
		Project: configtypes.ProjectConfig{
			Version: legacy.Version,
			Name:    legacy.Name,
			// ... other fields ...
		},
	}
}

// Add this method to ServiceImpl
func (s *ServiceImpl) Save() error {
	if s.projectConfig == nil {
		return fmt.Errorf("no configuration to save")
	}
	return s.configService.Save(s.projectConfig)
}

// Add LegacyConfig definition and fix conversions
type LegacyConfig struct {
	Version string
	Name    string
	// ... other legacy fields
}

func ConvertFromLegacy(legacy *LegacyConfig) *configtypes.Config {
	return &configtypes.Config{
		Project: configtypes.ProjectConfig{
			Version: legacy.Version,
			Name:    legacy.Name,
			// ... other fields
		},
	}
}

// Add this method to ProjectServiceImpl
func (s *ProjectServiceImpl) Migrate() error {
	if s.projectConfig == nil {
		return fmt.Errorf("no configuration to migrate")
	}

	currentVersion := s.projectConfig.Project.Version
	if currentVersion == "" {
		return fmt.Errorf("invalid version: version is empty")
	}

	// Check if already at latest version
	if currentVersion == "1.2" {
		s.logger.Debug("Configuration already at latest version")
		return nil
	}

	// Apply migrations in sequence
	migrations := []string{"1.0->1.1", "1.1->1.2"}
	for _, migration := range migrations {
		if migrationFunc, ok := migrationMap[migration]; ok {
			if err := migrationFunc(s.projectConfig); err != nil {
				return fmt.Errorf("migration %s failed: %w", migration, err)
			}
			s.logger.Debug("Applied migration", "from", currentVersion, "to", migration)
		}
	}

	return s.Save()
}
