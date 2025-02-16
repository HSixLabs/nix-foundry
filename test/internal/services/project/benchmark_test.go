package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
)

func BenchmarkProjectOperations(b *testing.B) {
	tmpDir := b.TempDir()
	configDir := filepath.Join(tmpDir, ".nix-foundry")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		b.Fatalf("Failed to create config directory: %v", err)
	}

	configService := config.NewService()
	platformSvc := platform.NewService()
	envService := environment.NewService(
		configService.GetConfigDir(),
		configService,
		platformSvc,
	)
	packageService := packages.NewService(tmpDir)
	projectService := project.NewService(configService, envService, packageService)

	// Sample project config for benchmarks
	projectConfig := &project.Config{
		Version:     "1.0",
		Name:        "benchmark-project",
		Environment: "development",
		Settings: config.Settings{
			LogLevel:   "info",
			AutoUpdate: true,
		},
		Dependencies: []string{"git", "docker", "nodejs"},
	}

	// Validate the config before using it
	if err := projectConfig.Validate(); err != nil {
		b.Fatalf("Invalid project config: %v", err)
	}

	// Initialize the project with the config
	if err := projectService.InitializeProject(projectConfig.Name, "test-team", true); err != nil {
		b.Fatalf("Failed to initialize project: %v", err)
	}

	// Export the config to a temporary file
	configPath := filepath.Join(configDir, "project-config.yaml")
	if err := projectService.ExportConfig(configPath); err != nil {
		b.Fatalf("Failed to export initial project config: %v", err)
	}

	// Import the config from the file
	if err := projectService.ImportConfig(configPath); err != nil {
		b.Fatalf("Failed to import initial project config: %v", err)
	}

	b.Run("import/export cycle", func(b *testing.B) {
		exportPath := filepath.Join(configDir, "benchmark-export.yaml")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Export
			if err := projectService.Export(exportPath); err != nil {
				b.Fatalf("Export failed: %v", err)
			}

			// Import
			if err := projectService.Import(exportPath); err != nil {
				b.Fatalf("Import failed: %v", err)
			}
		}
	})

	b.Run("validation", func(b *testing.B) {
		personalConfig := &config.Config{
			Version: "1.0",
			Settings: config.Settings{
				LogLevel:   "info",
				AutoUpdate: true,
			},
			Environment: config.EnvironmentSettings{
				Default:    "development",
				AutoSwitch: true,
			},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := projectService.ValidateConflicts(personalConfig); err != nil {
				b.Fatalf("ValidateConflicts failed: %v", err)
			}
		}
	})

	b.Run("load/save cycle", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := projectService.Save(); err != nil {
				b.Fatalf("Save failed: %v", err)
			}
			if err := projectService.Load(); err != nil {
				b.Fatalf("Load failed: %v", err)
			}
		}
	})
}

func BenchmarkConfig_Validate(b *testing.B) {
	// Create a valid config that exercises all validation rules
	cfg := &project.Config{
		Version:     "1.2",
		Name:        "test-project",
		Environment: "development",
		Settings: config.Settings{
			LogLevel:       "info",
			AutoUpdate:     true,
			UpdateInterval: "24h",
		},
		Dependencies: []string{"git", "docker"},
	}

	// Validate once before benchmarking to ensure the config is valid
	if err := cfg.Validate(); err != nil {
		b.Fatalf("Initial validation failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := cfg.Validate(); err != nil {
			b.Fatalf("Validation failed: %v", err)
		}
	}
}
