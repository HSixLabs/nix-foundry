package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
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
		envService := environment.NewService(
			configDir,
			configService,
			platform.NewService(),
		)
		packageService := packages.NewService(configDir)
		projectService := project.NewService(configService, envService, packageService)

		// Create test project config
		projectConfig := &project.Config{
			Version:     "1.0",
			Name:        "test-project",
			Environment: "development",
			Settings: config.Settings{
				LogLevel:   "info",
				AutoUpdate: true,
			},
			Dependencies: []string{"git"},
		}

		// Validate the config before using it
		if err := projectConfig.Validate(); err != nil {
			t.Fatalf("Invalid project config: %v", err)
		}

		// Test Export
		exportPath := filepath.Join(configDir, "export-test.yaml")

		// Initialize the project with the config
		if err := projectService.InitializeProject(projectConfig.Name, "test-team", true); err != nil {
			t.Fatalf("Failed to initialize project: %v", err)
		}

		// Export the initialized config
		if err := projectService.Export(exportPath); err != nil {
			t.Fatalf("Export failed: %v", err)
		}

		// Test Import
		newService := project.NewService(configService, envService, packageService)
		if err := newService.Import(exportPath); err != nil {
			t.Fatalf("Import failed: %v", err)
		}

		// Verify imported config matches original config
		importedConfig := newService.GetProjectConfig()
		if importedConfig.Version != projectConfig.Version {
			t.Errorf("Version mismatch: got %s, want %s", importedConfig.Version, projectConfig.Version)
		}
		if importedConfig.Environment != projectConfig.Environment {
			t.Errorf("Environment mismatch: got %s, want %s", importedConfig.Environment, projectConfig.Environment)
		}
		if len(importedConfig.Dependencies) != len(projectConfig.Dependencies) {
			t.Errorf("Dependencies length mismatch: got %d, want %d",
				len(importedConfig.Dependencies), len(projectConfig.Dependencies))
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
				Default:    "development",
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
		nilService := project.NewService(configService, envService, packageService)
		if err := nilService.Export("should-fail.yaml"); err == nil {
			t.Error("Export should fail with nil config")
		}
	})

	t.Run("import/export edge cases", func(t *testing.T) {
		configService := config.NewService()
		envService := environment.NewService(
			configDir,
			configService,
			platform.NewService(),
		)
		packageService := packages.NewService(configDir)
		projectService := project.NewService(configService, envService, packageService)

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
		validConfig := &project.Config{
			Version:     "1.2",
			Name:        "test-project",
			Environment: "development",
			Settings: config.Settings{
				LogLevel:   "info",
				AutoUpdate: true,
			},
		}

		validPath := filepath.Join(configDir, "valid.yaml")
		data, marshalErr := json.Marshal(validConfig)
		if marshalErr != nil {
			t.Fatalf("Failed to marshal test config: %v", marshalErr)
		}

		writeErr := os.WriteFile(validPath, data, 0644)
		if writeErr != nil {
			t.Fatalf("Failed to write test config: %v", writeErr)
		}

		importErr := projectService.Import(validPath)
		if importErr != nil {
			t.Fatalf("Import should succeed with valid config: %v", importErr)
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
