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
	"gopkg.in/yaml.v3"
)

func TestProjectIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	homeDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	t.Setenv("HOME", homeDir)
	defer func() {
		os.Setenv("HOME", originalHome)
	}()

	configDir := filepath.Join(homeDir, ".nix-foundry")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	t.Run("full configuration workflow", func(t *testing.T) {
		// Initialize services
		configService := config.NewService()
		envService := environment.NewService(configDir, configService, validation.NewService(), platform.NewService())
		packageService := packages.NewService(configDir)
		projectService := NewService(configService, envService, packageService)

		// Create test project config
		projectConfig := &ProjectConfig{
			Version:     "1.0",
			Name:        "test-project",
			Environment: "development",
			Settings: config.Settings{
				LogLevel:   "info",
				AutoUpdate: true,
			},
			Dependencies: []string{"git"},
		}

		// Test Export
		exportPath := filepath.Join(configDir, "export-test.yaml")
		service := projectService.(*ServiceImpl)
		service.projectConfig = projectConfig
		if err := service.Export(exportPath); err != nil {
			t.Fatalf("Export failed: %v", err)
		}

		// Test Import
		newService := NewService(configService, envService, packageService)
		if err := newService.Import(exportPath); err != nil {
			t.Fatalf("Import failed: %v", err)
		}

		// Verify imported config
		importedConfig := newService.GetProjectConfig()
		if importedConfig.Name != projectConfig.Name {
			t.Errorf("Import/Export mismatch: got name %s, want %s", importedConfig.Name, projectConfig.Name)
		}

		// Test backup creation during import
		backupDir := filepath.Join(configDir, "backups")
		files, err := os.ReadDir(backupDir)
		if err != nil {
			t.Fatalf("Failed to read backup directory: %v", err)
		}
		if len(files) == 0 {
			t.Error("No backup file created during import")
		}

		// Test conflict validation with personal config
		personalConfig := &config.Config{
			Version: "1.0",
			Settings: config.Settings{
				LogLevel:   "info",
				AutoUpdate: true,
			},
			Environment: config.EnvironmentSettings{
				Default: "development",
				AutoSwitch: true,
			},
		}

		if err := newService.ValidateConflicts(personalConfig); err != nil {
			t.Errorf("ValidateConflicts failed: %v", err)
		}

		// Test invalid import path
		if err := newService.Import("nonexistent.yaml"); err == nil {
			t.Error("Import should fail with nonexistent file")
		}

		// Test export with nil config
		nilService := NewService(configService, envService, packageService)
		if err := nilService.Export("should-fail.yaml"); err == nil {
			t.Error("Export should fail with nil config")
		}
	})

	t.Run("import/export edge cases", func(t *testing.T) {
		configService := config.NewService()
		envService := environment.NewService(configDir, configService, validation.NewService(), platform.NewService())
		packageService := packages.NewService(configDir)
		projectService := NewService(configService, envService, packageService)

		// Test importing from directory
		testDir := filepath.Join(configDir, "test-dir")
		if err := os.MkdirAll(testDir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Test importing from directory without config file
		if err := projectService.Import(testDir); err == nil {
			t.Error("Import should fail when directory has no config file")
		}

		// Create invalid config file
		invalidPath := filepath.Join(configDir, "invalid.yaml")
		if err := os.WriteFile(invalidPath, []byte("invalid: yaml: content"), 0644); err != nil {
			t.Fatalf("Failed to create invalid config file: %v", err)
		}

		// Test importing invalid config
		if err := projectService.Import(invalidPath); err == nil {
			t.Error("Import should fail with invalid config file")
		}

		// Test exporting to invalid path
		if err := projectService.Export("/invalid/path/config.yaml"); err == nil {
			t.Error("Export should fail with invalid path")
		}

		// Test backup creation during import
		validConfig := &ProjectConfig{
			Version:     "1.2",
			Name:        "test-project",
			Environment: "development",
			Settings: config.Settings{
				LogLevel:   "info",
				AutoUpdate: true,
			},
		}

		validPath := filepath.Join(configDir, "valid.yaml")
		data, err := yaml.Marshal(validConfig)
		if err != nil {
			t.Fatalf("Failed to marshal valid config: %v", err)
		}

		if err := os.WriteFile(validPath, data, 0644); err != nil {
			t.Fatalf("Failed to write valid config file: %v", err)
		}

		if err := projectService.Import(validPath); err != nil {
			t.Fatalf("Import failed with valid config: %v", err)
		}

		// Verify backup was created
		backupFiles, err := os.ReadDir(filepath.Join(configDir, "backups"))
		if err != nil {
			t.Fatalf("Failed to read backup directory: %v", err)
		}
		if len(backupFiles) == 0 {
			t.Error("No backup file created during import")
		}
	})
}
