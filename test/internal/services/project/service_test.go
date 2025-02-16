package project

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
)

type mockConfigService struct {
	config.Service
	loadSectionErr error
	saveSectionErr error
	projectConfig  project.Config
}

func (m *mockConfigService) LoadSection(name string, v interface{}) error {
	if m.loadSectionErr != nil {
		return m.loadSectionErr
	}
	*(v.(*project.Config)) = m.projectConfig
	return nil
}

func (m *mockConfigService) SaveSection(name string, v interface{}) error {
	if m.saveSectionErr != nil {
		return m.saveSectionErr
	}
	return nil
}

// testServiceImpl extends ServiceImpl for testing
type testServiceImpl struct {
	project.Service
	ConfigService config.Service
	ProjectConfig *project.Config
	Logger        *logging.Logger
}

func newTestServiceImpl(configService config.Service) *testServiceImpl {
	return &testServiceImpl{
		ConfigService: configService,
		Logger:        logging.GetLogger(),
	}
}

func (s *testServiceImpl) Load() error {
	return nil
}

func (s *testServiceImpl) Save() error {
	return nil
}

func (s *testServiceImpl) ValidateConflicts(personal *config.Config) error {
	if s.ProjectConfig == nil {
		return errors.NewValidationError("project", fmt.Errorf("project config is nil"), "project configuration validation failed")
	}
	return nil
}

func (s *testServiceImpl) GetProjectConfig() *project.Config {
	return s.ProjectConfig
}

func TestProjectService(t *testing.T) {
	validProjectConfig := project.Config{
		Version:     "1.0",
		Name:        "test-project",
		Environment: "development",
		Settings: config.Settings{
			LogLevel:   "info",
			AutoUpdate: true,
		},
		Dependencies: []string{"git", "docker"},
	}

	t.Run("load project config", func(t *testing.T) {
		mockConfig := &mockConfigService{
			projectConfig: validProjectConfig,
		}
		tmpDir := t.TempDir()
		envService := environment.NewService(
			tmpDir,
			mockConfig,
			platform.NewService(),
		)
		packageService := packages.NewService(tmpDir)
		service := project.NewService(mockConfig, envService, packageService)

		err := service.Load()
		if err != nil {
			t.Errorf("Load() error = %v, want nil", err)
		}

		cfg := service.GetProjectConfig()
		if cfg.Name != validProjectConfig.Name {
			t.Errorf("GetProjectConfig().Name = %v, want %v", cfg.Name, validProjectConfig.Name)
		}
	})

	t.Run("save project config", func(t *testing.T) {
		mockConfig := &mockConfigService{}
		service := newTestServiceImpl(mockConfig)
		service.ProjectConfig = &validProjectConfig

		err := service.Save()
		if err != nil {
			t.Errorf("Save() error = %v, want nil", err)
		}
	})

	t.Run("validate conflicts", func(t *testing.T) {
		tests := []struct {
			name           string
			projectConfig  *project.Config
			personalConfig *config.Config
			wantErr        bool
		}{
			{
				name: "matching environments",
				projectConfig: &project.Config{
					Environment: "development",
					Settings: config.Settings{
						LogLevel: "info",
					},
				},
				personalConfig: &config.Config{
					Environment: config.EnvironmentSettings{
						Default: "development",
					},
				},
				wantErr: false,
			},
			{
				name: "mismatched environments",
				projectConfig: &project.Config{
					Environment: "production",
					Settings: config.Settings{
						LogLevel: "info",
					},
				},
				personalConfig: &config.Config{
					Environment: config.EnvironmentSettings{
						Default: "development",
					},
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				service := newTestServiceImpl(nil)
				service.ProjectConfig = tt.projectConfig

				err := service.ValidateConflicts(tt.personalConfig)
				if (err != nil) != tt.wantErr {
					t.Errorf("ValidateConflicts() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("nil project config", func(t *testing.T) {
		service := newTestServiceImpl(&mockConfigService{})
		service.ProjectConfig = nil

		err := service.ValidateConflicts(&config.Config{})
		if err == nil {
			t.Error("ValidateConflicts() with nil project config should return error")
		}
	})
}

func TestProjectImportExport(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("successful import", func(t *testing.T) {
		// Create a test config file
		configPath := filepath.Join(tmpDir, "test-config.yaml")
		testConfig := `
version: "1.0"
name: "test-project"
environment: "development"
settings:
  logLevel: "info"
  autoUpdate: true
dependencies:
  - "git"
`
		if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
			t.Fatalf("Failed to create test config: %v", err)
		}

		mockConfig := &mockConfigService{
			saveSectionErr: nil,
		}
		service := newTestServiceImpl(mockConfig)
		service.ProjectConfig = &project.Config{}

		if err := service.Import(configPath); err != nil {
			t.Errorf("Import() error = %v, want nil", err)
		}
	})

	t.Run("failed import - invalid file", func(t *testing.T) {
		service := newTestServiceImpl(&mockConfigService{})
		service.ProjectConfig = nil

		if err := service.Import("nonexistent.yaml"); err == nil {
			t.Error("Import() should fail with nonexistent file")
		}
	})

	t.Run("successful export", func(t *testing.T) {
		exportPath := filepath.Join(tmpDir, "export-config.yaml")
		service := newTestServiceImpl(&mockConfigService{})
		service.ProjectConfig = &project.Config{
			Version:     "1.0",
			Name:        "test-project",
			Environment: "development",
			Settings: config.Settings{
				LogLevel:   "info",
				AutoUpdate: true,
			},
			Dependencies: []string{"git"},
		}

		if err := service.Export(exportPath); err != nil {
			t.Errorf("Export() failed: %v", err)
		}

		// Verify exported file exists and is readable
		if _, err := os.Stat(exportPath); err != nil {
			t.Errorf("Export() did not create file: %v", err)
		}

		// Import the exported file to verify structure
		newService := newTestServiceImpl(&mockConfigService{})
		newService.ProjectConfig = nil
		if err := newService.Import(exportPath); err != nil {
			t.Errorf("Failed to import exported file: %v", err)
		}
	})

	t.Run("failed export - no config", func(t *testing.T) {
		service := newTestServiceImpl(&mockConfigService{})
		service.ProjectConfig = nil

		if err := service.Export(filepath.Join(tmpDir, "should-fail.yaml")); err == nil {
			t.Error("Export() should fail with nil projectConfig")
		}
	})
}
