package project

import (
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	configtypes "github.com/shawnkhoffman/nix-foundry/internal/services/config/types"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
)

// testService extends ProjectService to expose config for testing
type testService struct {
	project.ProjectService
	projectServiceImpl *project.ProjectServiceImpl
	Config             *project.Config
}

func newTestService(configService config.Service, envService environment.Service, packageService packages.Service) *testService {
	impl := project.NewService(configService, envService, packageService).(*project.ProjectServiceImpl)
	return &testService{
		ProjectService:     impl,
		projectServiceImpl: impl,
		Config:             nil,
	}
}

func (s *testService) setConfig(cfg *project.Config) {
	s.Config = cfg
}

func (s *testService) getConfig() *project.Config {
	return s.Config
}

func (s *testService) Backup(projectID string) error {
	// For testing purposes, we can return nil or implement basic backup logic
	return nil
}

func (s *testService) Migrate() error {
	return s.projectServiceImpl.Migrate()
}

func TestMigration(t *testing.T) {
	tmpDir := t.TempDir()
	configService := config.NewService()
	envService := environment.NewService(
		tmpDir,
		configService,
		platform.NewService(),
		false,
		true,
		true,
	)
	packageService := packages.NewService(tmpDir)
	service := newTestService(configService, envService, packageService)

	t.Run("migrate v1.0 to v1.1", func(t *testing.T) {
		testConfig := &project.Config{
			Project: configtypes.ProjectConfig{
				Version:     "1.0.0",
				Name:        "test-project",
				Environment: "development",
				Settings: map[string]string{
					"logLevel":   "info",
					"autoUpdate": "true",
				},
				Dependencies: []string{"git", "docker"},
			},
		}
		service.setConfig(testConfig)

		if err := service.Migrate(); err != nil {
			t.Fatalf("Migration failed: %v", err)
		}

		cfg := service.getConfig()
		if cfg.Project.Version != "1.2" {
			t.Errorf("Expected version 1.2, got %s", cfg.Project.Version)
		}

		// Check migrated fields
		if cfg.Project.Settings["logLevel"] != "info" {
			t.Error("LogLevel not set to default value")
		}

		// Check trimmed dependencies
		expectedDeps := []string{"git", "docker"}
		for i, dep := range cfg.Project.Dependencies {
			if dep != expectedDeps[i] {
				t.Errorf("Dependency not trimmed, expected %q got %q", expectedDeps[i], dep)
			}
		}

		if cfg.Project.Settings["updateInterval"] != "24h" {
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
			Project: configtypes.ProjectConfig{
				Version: "invalid",
			},
		})
		if err := service.Migrate(); err == nil {
			t.Error("Expected error for invalid version")
		}
	})

	t.Run("already latest version", func(t *testing.T) {
		service.setConfig(&project.Config{
			Project: configtypes.ProjectConfig{
				Version: "1.2",
			},
		})
		if err := service.Migrate(); err != nil {
			t.Errorf("Unexpected error for latest version: %v", err)
		}
	})
}
