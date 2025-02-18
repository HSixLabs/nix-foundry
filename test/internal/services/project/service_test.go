package project

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/types"
	configservice "github.com/shawnkhoffman/nix-foundry/internal/services/config"
	configtypes "github.com/shawnkhoffman/nix-foundry/internal/services/config/types"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
)

type mockConfigService struct {
	cfg              *types.Config
	nixCfg           *types.NixConfig
	compressionLevel int
	maxBackups       int
	retentionDays    int
}

func (m *mockConfigService) LoadConfig(configType configservice.Type, path string) (interface{}, error) {
	return m.cfg, nil
}

func (m *mockConfigService) SaveConfig(cfg *types.Config) error {
	return nil
}

func (m *mockConfigService) GetValue(key string) (interface{}, error) {
	return "", nil
}

func (m *mockConfigService) GetRetentionDays() int {
	return 10
}

func (m *mockConfigService) GetMaxBackups() int {
	return 10
}

func (m *mockConfigService) GetConfig() *types.NixConfig {
	return m.nixCfg
}

func (m *mockConfigService) Save(cfg *types.Config) error {
	m.cfg = cfg
	return nil
}

func (m *mockConfigService) Load() (*types.Config, error) {
	return m.cfg, nil
}

// Add missing methods to implement configtypes.Service interface
func (m *mockConfigService) Apply(config *types.NixConfig, testMode bool) error {
	return nil
}

func (m *mockConfigService) GetConfigDir() string {
	return ""
}

func (m *mockConfigService) GetBackupDir() string {
	return ""
}

func (m *mockConfigService) LoadSection(name string, v interface{}) error {
	return nil
}

func (m *mockConfigService) SaveSection(name string, v interface{}) error {
	return nil
}

func (m *mockConfigService) ApplyFlags(flags map[string]string, force bool) error {
	return nil
}

func (m *mockConfigService) GenerateInitialConfig(shell, editor, gitName, gitEmail string) (*types.Config, error) {
	return nil, nil
}

func (m *mockConfigService) PreviewConfiguration(*types.Config) error {
	return nil
}

func (m *mockConfigService) ConfigExists() bool {
	return true
}

func (m *mockConfigService) Initialize(testMode bool) error {
	return nil
}

// Add CreateBackup method
func (m *mockConfigService) CreateBackup(path string) error {
	return nil
}

// Add RestoreBackup method
func (m *mockConfigService) RestoreBackup(path string) error {
	return nil
}

// Add CreateConfigFromMap method
func (m *mockConfigService) CreateConfigFromMap(configMap map[string]string) *types.NixConfig {
	return &types.NixConfig{
		Version: "1.0",
		Settings: types.Settings{
			LogLevel: configMap["logLevel"],
		},
	}
}

// Add Generate method
func (m *mockConfigService) Generate(defaultEnv string, nixCfg *types.NixConfig) error {
	return nil
}

// Update GenerateEncryptionKey method to match interface
func (m *mockConfigService) GenerateEncryptionKey() error {
	return nil
}

// Add GetCompressionLevel method
func (m *mockConfigService) GetCompressionLevel() int {
	return 6 // Default gzip compression level
}

// Add GetLogger method
func (m *mockConfigService) GetLogger() *logging.Logger {
	return logging.GetLogger()
}

// Add LoadCustomPackages method
func (m *mockConfigService) LoadCustomPackages() ([]string, error) {
	return []string{}, nil
}

// Update LoadProjectWithTeam method to match interface
func (m *mockConfigService) LoadProjectWithTeam(team, projectID string) (*types.ProjectConfig, error) {
	if m.cfg == nil {
		return nil, nil
	}
	return &m.cfg.Project, nil
}

// Update MergeProjectConfigs to properly check for empty config
func (m *mockConfigService) MergeProjectConfigs(projectConfig types.ProjectConfig, teamConfig types.ProjectConfig) types.ProjectConfig {
	// Check if projectConfig is empty by checking its required fields
	if projectConfig.Name == "" && projectConfig.Version == "" {
		return teamConfig
	}
	return projectConfig
}

// Add ReadConfig method to match interface
func (m *mockConfigService) ReadConfig(path string, v interface{}) error {
	return nil
}

// Update Reset method to match interface signature
func (m *mockConfigService) Reset(configType string) error {
	m.cfg = nil
	m.nixCfg = nil
	return nil
}

// Add ResetValue method to match interface
func (m *mockConfigService) ResetValue(key string) error {
	return nil
}

// Add RotateEncryptionKey method to match interface
func (m *mockConfigService) RotateEncryptionKey() error {
	return nil
}

// Add GetEncryptionKey method to match interface
func (m *mockConfigService) GetEncryptionKey() (string, error) {
	return "", nil
}

// Add SaveCustomPackages method to match interface
func (m *mockConfigService) SaveCustomPackages(packages []string) error {
	return nil
}

// Update SetCompressionLevel method to use the field
func (m *mockConfigService) SetCompressionLevel(level int) {
	m.compressionLevel = level
}

// Add SetMaxBackups method to match interface
func (m *mockConfigService) SetMaxBackups(maxBackups int) {
	m.maxBackups = maxBackups
}

// Add SetRetentionDays method to match interface
func (m *mockConfigService) SetRetentionDays(days int) {
	m.retentionDays = days
}

// Add SetValue method to match interface
func (m *mockConfigService) SetValue(key string, value interface{}) error {
	return nil
}

// Add Validate method to match interface
func (m *mockConfigService) Validate() error {
	return nil
}

// Update ValidateConfiguration method to match interface
func (m *mockConfigService) ValidateConfiguration(testMode bool) error {
	return nil
}

// Update WriteConfig method to match interface
func (m *mockConfigService) WriteConfig(path string, v interface{}) error {
	return nil
}

// testServiceImpl extends ServiceImpl for testing
type testServiceImpl struct {
	project.Service
	ConfigService configtypes.Service
	ProjectConfig *configtypes.Config
	Logger        *logging.Logger
}

func newTestServiceImpl(configService configtypes.Service) *testServiceImpl {
	return &testServiceImpl{
		ConfigService: configService,
		Logger:        logging.GetLogger(),
	}
}

func (s *testServiceImpl) Load() error {
	_, err := s.ConfigService.Load()
	return err
}

func (s *testServiceImpl) Save() error {
	cfg, err := s.ConfigService.Load()
	if err != nil {
		return err
	}
	return s.ConfigService.Save(cfg)
}

func (s *testServiceImpl) ValidateConflicts(personal *types.Config) error {
	if s.ProjectConfig == nil {
		return errors.NewValidationError("project", fmt.Errorf("project config is nil"), "project configuration validation failed")
	}
	return nil
}

func (s *testServiceImpl) GetProjectConfig() *project.Config {
	if s.ProjectConfig == nil {
		return nil
	}
	return &project.Config{
		Project: s.ProjectConfig.Project,
	}
}

func TestProjectService(t *testing.T) {
	validProjectConfig := types.Config{
		Project: types.ProjectConfig{
			Version:     "1.0",
			Name:        "test-project",
			Environment: "development",
			Settings: map[string]string{
				"LogLevel":   "info",
				"AutoUpdate": "true",
			},
		},
	}

	t.Run("load project config", func(t *testing.T) {
		mockConfig := &mockConfigService{
			cfg: &types.Config{
				Version: "1.0",
				Project: types.ProjectConfig{
					Name: "test",
				},
			},
			nixCfg: &types.NixConfig{
				Version: "1.0",
			},
		}
		tmpDir := t.TempDir()
		envService := environment.NewService(
			tmpDir,
			mockConfig,
			platform.NewService(),
			true,
			true,
			true,
		)
		packageService := packages.NewService(tmpDir)
		service := project.NewService(mockConfig, envService, packageService)

		err := service.Load()
		if err != nil {
			t.Errorf("Load() error = %v, want nil", err)
		}

		cfg := service.GetProjectConfig()
		if cfg.Project.Name != validProjectConfig.Project.Name {
			t.Errorf("GetProjectConfig().Project.Name = %v, want %v", cfg.Project.Name, validProjectConfig.Project.Name)
		}
	})

	t.Run("save project config", func(t *testing.T) {
		mockConfig := &mockConfigService{}
		service := newTestServiceImpl(mockConfig)
		service.ProjectConfig = &configtypes.Config{
			Project: validProjectConfig.Project,
		}

		err := service.Save()
		if err != nil {
			t.Errorf("Save() error = %v, want nil", err)
		}
	})

	t.Run("validate conflicts", func(t *testing.T) {
		tests := []struct {
			name           string
			projectConfig  *project.Config
			personalConfig *types.Config
			wantErr        bool
		}{
			{
				name: "matching environments",
				projectConfig: project.FromProjectConfig(types.ProjectConfig{
					Environment: "development",
					Settings: map[string]string{
						"LogLevel": "info",
					},
				}),
				personalConfig: &types.Config{
					Environment: types.EnvironmentSettings{
						Default: "development",
					},
				},
				wantErr: false,
			},
			{
				name: "mismatched environments",
				projectConfig: project.FromProjectConfig(types.ProjectConfig{
					Environment: "production",
					Settings: map[string]string{
						"LogLevel": "info",
					},
				}),
				personalConfig: &types.Config{
					Environment: types.EnvironmentSettings{
						Default: "development",
					},
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				service := newTestServiceImpl(nil)
				service.ProjectConfig = &configtypes.Config{
					Project: tt.projectConfig.Project,
				}

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

		err := service.ValidateConflicts(&types.Config{})
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

		mockConfig := &mockConfigService{}
		service := newTestServiceImpl(mockConfig)
		service.ProjectConfig = &configtypes.Config{
			Project: configtypes.ProjectConfig{
				Version:     "1.0",
				Name:        "test-project",
				Environment: "development",
				Settings: map[string]string{
					"LogLevel":   "info",
					"AutoUpdate": "true",
				},
				Dependencies: []string{"git"},
			},
		}

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
		service.ProjectConfig = &configtypes.Config{
			Project: configtypes.ProjectConfig{
				Version:     "1.0",
				Name:        "test-project",
				Environment: "development",
				Settings: map[string]string{
					"LogLevel":   "info",
					"AutoUpdate": "true",
				},
				Dependencies: []string{"git"},
			},
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

func TestProjectConfig(t *testing.T) {
	cfg := &types.Config{
		Environment: types.EnvironmentSettings{
			Default:    "development",
			AutoSwitch: true,
		},
	}

	// Test environment settings validation
	service := newTestServiceImpl(&mockConfigService{})
	service.ProjectConfig = &configtypes.Config{
		Project: configtypes.ProjectConfig{
			Environment: "development",
		},
	}

	err := service.ValidateConflicts(cfg)
	if err != nil {
		t.Errorf("ValidateConflicts() error = %v, want nil", err)
	}

	// Test mismatched environment
	cfg.Environment.Default = "production"
	err = service.ValidateConflicts(cfg)
	if err == nil {
		t.Error("ValidateConflicts() error = nil, want error for mismatched environment")
	}
}

func TestConfigValidation(t *testing.T) {
	cfg := &configtypes.Config{
		Project: configtypes.ProjectConfig{
			Version:     "1.0",
			Name:        "test-project",
			Environment: "development",
			Settings:    make(map[string]string),
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("Expected valid config to pass validation: %v", err)
	}
}

func TestService_ValidateSettingsConflicts_NilConfig(t *testing.T) {
	mockConfig := &mockConfigService{
		cfg: &configtypes.Config{
			Project: configtypes.ProjectConfig{
				Version: "1.2",
			},
		},
	}
	service := newTestServiceImpl(mockConfig)

	err := service.ValidateConflicts(nil)
	if err == nil {
		t.Error("Expected error for nil config")
	}
}

func TestService_CreateProject(t *testing.T) {
	validConfig := &configtypes.Config{
		Project: configtypes.ProjectConfig{
			Version: "1.0.0",
			Name:    "test",
		},
	}

	mockConfig := &mockConfigService{cfg: validConfig}
	service := newTestServiceImpl(mockConfig)

	err := service.InitializeProject("test", "team", false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestProjectConfig_ValidateFields(t *testing.T) {
	validConfig := &configtypes.Config{
		Project: configtypes.ProjectConfig{
			Version: "1.2",
			Name:    "test-project",
		},
	}

	invalidConfig := &configtypes.Config{
		Project: configtypes.ProjectConfig{
			Version: "1.2",
			Name:    "",
		},
	}

	tests := []struct {
		name    string
		config  *configtypes.Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  validConfig,
			wantErr: false,
		},
		{
			name:    "invalid config",
			config:  invalidConfig,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
