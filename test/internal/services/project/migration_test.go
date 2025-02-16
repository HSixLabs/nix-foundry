package project

import (
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
)

// testService extends ServiceImpl to expose config for testing
type testService struct {
	*project.ServiceImpl
	Config *project.Config
}

func newTestService(configService config.Service, envService environment.Service, packageService packages.Service) *testService {
	return &testService{
		ServiceImpl: project.NewService(configService, envService, packageService).(*project.ServiceImpl),
	}
}

func (s *testService) setConfig(cfg *project.Config) {
	s.Config = cfg
}

func (s *testService) getConfig() *project.Config {
	return s.Config
}

func TestMigration(t *testing.T) {
	tmpDir := t.TempDir()
	configService := config.NewService()
	envService := environment.NewService(
		tmpDir,
		configService,
		platform.NewService(),
	)
	packageService := packages.NewService(tmpDir)
	service := newTestService(configService, envService, packageService)

	t.Run("migrate v1.0 to v1.1", func(t *testing.T) {
		service.setConfig(&project.Config{
			Version: "1.0",
			Name:    "test-project",
			Settings: config.Settings{
				AutoUpdate: true,
			},
			Dependencies: []string{" git ", "  docker  "},
		})

		if err := service.Migrate(); err != nil {
			t.Fatalf("Migration failed: %v", err)
		}

		cfg := service.getConfig()
		if cfg.Version != "1.2" {
			t.Errorf("Expected version 1.2, got %s", cfg.Version)
		}

		// Check migrated fields
		if cfg.Settings.LogLevel != "info" {
			t.Error("LogLevel not set to default value")
		}

		// Check trimmed dependencies
		expectedDeps := []string{"git", "docker"}
		for i, dep := range cfg.Dependencies {
			if dep != expectedDeps[i] {
				t.Errorf("Dependency not trimmed, expected %q got %q", expectedDeps[i], dep)
			}
		}

		if cfg.Settings.UpdateInterval != "24h" {
			t.Error("UpdateInterval not set for AutoUpdate enabled config")
		}
	})

	t.Run("nil config", func(t *testing.T) {
		service.setConfig(nil)
		if err := service.Migrate(); err == nil {
			t.Error("Expected error for nil config")
		}
	})

	t.Run("invalid version", func(t *testing.T) {
		service.setConfig(&project.Config{
			Version: "invalid",
		})
		if err := service.Migrate(); err == nil {
			t.Error("Expected error for invalid version")
		}
	})

	t.Run("already latest version", func(t *testing.T) {
		service.setConfig(&project.Config{
			Version: "1.2",
		})
		if err := service.Migrate(); err != nil {
			t.Errorf("Unexpected error for latest version: %v", err)
		}
	})
}
