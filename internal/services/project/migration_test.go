package project

import (
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/validation"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
)

func TestMigration(t *testing.T) {
	tmpDir := t.TempDir()
	configService := config.NewService()
	envService := environment.NewService(tmpDir, configService, validation.NewService(), platform.NewService())
	packageService := packages.NewService(tmpDir)
	service := NewService(configService, envService, packageService).(*ServiceImpl)

	t.Run("migrate v1.0 to v1.1", func(t *testing.T) {
		service.projectConfig = &ProjectConfig{
			Version: "1.0",
			Name:    "test-project",
			Settings: config.Settings{
				AutoUpdate: true,
			},
			Dependencies: []string{" git ", "  docker  "},
		}

		if err := service.Migrate(); err != nil {
			t.Fatalf("Migration failed: %v", err)
		}

		if service.projectConfig.Version != "1.2" {
			t.Errorf("Expected version 1.2, got %s", service.projectConfig.Version)
		}

		// Check migrated fields
		if service.projectConfig.Settings.LogLevel != "info" {
			t.Error("LogLevel not set to default value")
		}

		// Check trimmed dependencies
		expectedDeps := []string{"git", "docker"}
		for i, dep := range service.projectConfig.Dependencies {
			if dep != expectedDeps[i] {
				t.Errorf("Dependency not trimmed, expected %q got %q", expectedDeps[i], dep)
			}
		}

		if service.projectConfig.Settings.UpdateInterval != "24h" {
			t.Error("UpdateInterval not set for AutoUpdate enabled config")
		}
	})

	t.Run("nil config", func(t *testing.T) {
		service.projectConfig = nil
		if err := service.Migrate(); err == nil {
			t.Error("Expected error for nil config")
		}
	})

	t.Run("invalid version", func(t *testing.T) {
		service.projectConfig = &ProjectConfig{
			Version: "invalid",
		}
		if err := service.Migrate(); err == nil {
			t.Error("Expected error for invalid version")
		}
	})

	t.Run("already latest version", func(t *testing.T) {
		service.projectConfig = &ProjectConfig{
			Version: "1.2",
		}
		if err := service.Migrate(); err != nil {
			t.Errorf("Unexpected error for latest version: %v", err)
		}
	})
}
