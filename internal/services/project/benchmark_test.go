package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/validation"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
)

func BenchmarkProjectOperations(b *testing.B) {
	tmpDir := b.TempDir()
	configDir := filepath.Join(tmpDir, ".nix-foundry")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		b.Fatalf("Failed to create config directory: %v", err)
	}

	configService := config.NewService()
	envService := environment.NewService(tmpDir, configService, validation.NewService(), platform.NewService())
	packageService := packages.NewService(tmpDir)
	projectService := NewService(configService, envService, packageService)

	// Sample project config for benchmarks
	projectConfig := &ProjectConfig{
		Version:     "1.0",
		Name:        "benchmark-project",
		Environment: "development",
		Settings: config.Settings{
			LogLevel:   "info",
			AutoUpdate: true,
		},
		Dependencies: []string{"git", "docker", "nodejs"},
	}

	b.Run("import/export cycle", func(b *testing.B) {
		exportPath := filepath.Join(configDir, "benchmark-export.yaml")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Setup
			service := projectService.(*ServiceImpl)
			service.projectConfig = projectConfig

			// Export
			if err := service.Export(exportPath); err != nil {
				b.Fatalf("Export failed: %v", err)
			}

			// Import
			newService := NewService(configService, envService, packageService)
			if err := newService.Import(exportPath); err != nil {
				b.Fatalf("Import failed: %v", err)
			}
		}
	})

	b.Run("validation", func(b *testing.B) {
		service := projectService.(*ServiceImpl)
		service.projectConfig = projectConfig

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
			if err := service.ValidateConflicts(personalConfig); err != nil {
				b.Fatalf("ValidateConflicts failed: %v", err)
			}
		}
	})

	b.Run("load/save cycle", func(b *testing.B) {
		service := projectService.(*ServiceImpl)
		service.projectConfig = projectConfig

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := service.Save(); err != nil {
				b.Fatalf("Save failed: %v", err)
			}
			if err := service.Load(); err != nil {
				b.Fatalf("Load failed: %v", err)
			}
		}
	})
}
