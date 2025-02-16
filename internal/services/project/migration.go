package project

import (
	"fmt"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
)

// MigrationFunc defines a function that migrates configuration from one version to another
type MigrationFunc func(*ProjectConfig) error

// migrationMap stores available migrations between versions
var migrationMap = map[string]MigrationFunc{
	"1.0->1.1": migrateV1_0ToV1_1,
	"1.1->1.2": migrateV1_1ToV1_2,
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
func migrateV1_0ToV1_1(cfg *ProjectConfig) error {
	// Add new fields with default values
	if cfg.Settings.LogLevel == "" {
		cfg.Settings.LogLevel = "info"
	}

	// Convert legacy fields
	for i, dep := range cfg.Dependencies {
		cfg.Dependencies[i] = strings.TrimSpace(dep)
	}

	return nil
}

func migrateV1_1ToV1_2(cfg *ProjectConfig) error {
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
