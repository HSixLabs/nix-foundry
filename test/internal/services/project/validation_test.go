package project

import (
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
)

// TestConfig_Validate tests the Config.Validate method
func TestConfig_Validate(t *testing.T) {
	validConfig := &project.Config{
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

	t.Run("valid config", func(t *testing.T) {
		if err := validConfig.Validate(); err != nil {
			t.Errorf("Expected valid config to pass validation: %v", err)
		}
	})

	tests := []struct {
		name       string
		mutateFunc func(*project.Config)
		wantErrMsg string
	}{
		{
			name: "empty version",
			mutateFunc: func(c *project.Config) {
				c.Version = ""
			},
			wantErrMsg: "invalid version: version is required",
		},
		{
			name: "invalid version",
			mutateFunc: func(c *project.Config) {
				c.Version = "2.0"
			},
			wantErrMsg: "invalid version: unsupported version: 2.0",
		},
		{
			name: "empty name",
			mutateFunc: func(c *project.Config) {
				c.Name = ""
			},
			wantErrMsg: "invalid name: name is required",
		},
		{
			name: "name too long",
			mutateFunc: func(c *project.Config) {
				c.Name = "this-is-a-very-long-project-name-that-exceeds-fifty-characters-limit"
			},
			wantErrMsg: "invalid name: name exceeds maximum length of 50 characters",
		},
		{
			name: "invalid environment",
			mutateFunc: func(c *project.Config) {
				c.Environment = "test"
			},
			wantErrMsg: "invalid environment: invalid environment: test",
		},
		{
			name: "invalid log level",
			mutateFunc: func(c *project.Config) {
				c.Settings.LogLevel = "trace"
			},
			wantErrMsg: "invalid settings: invalid log level: trace",
		},
		{
			name: "invalid update interval",
			mutateFunc: func(c *project.Config) {
				c.Settings.AutoUpdate = true
				c.Settings.UpdateInterval = "invalid"
			},
			wantErrMsg: "invalid settings: invalid update interval format: invalid",
		},
		{
			name: "duplicate dependency",
			mutateFunc: func(c *project.Config) {
				c.Dependencies = []string{"git", "git"}
			},
			wantErrMsg: "invalid dependencies: duplicate dependency: git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := *validConfig // Create a copy
			tt.mutateFunc(&cfg)

			err := cfg.Validate()
			if err == nil {
				t.Error("Expected validation error, got nil")
				return
			}
			if err.Error() != tt.wantErrMsg {
				t.Errorf("Expected error message %q, got %q", tt.wantErrMsg, err.Error())
			}
		})
	}
}

// Rename testService to validationTestService
type validationTestService struct {
	*project.ServiceImpl
	Config *project.Config
}

// Rename newTestService to newValidationTestService and update its implementation
func newValidationTestService(configService config.Service) *validationTestService {
	return &validationTestService{
		ServiceImpl: project.NewService(configService, nil, nil).(*project.ServiceImpl),
	}
}

func (s *validationTestService) SetConfig(cfg *project.Config) {
	s.Config = cfg
}

// Update the test to use the renamed test service
func TestServiceImpl_ValidateSettingsConflicts_NilConfig(t *testing.T) {
	configService := &mockConfigService{} // Use the existing mock
	service := newValidationTestService(configService)

	settings := config.Settings{
		LogLevel:       "info",
		AutoUpdate:     true,
		UpdateInterval: "24h",
	}

	err := service.ValidateConflicts(&config.Config{Settings: settings})
	if err == nil {
		t.Error("Expected error for nil config, got nil")
		return
	}
}

// Update the remaining tests to use the public Validate method
func TestConfig_ValidateSettings_EmptyInterval(t *testing.T) {
	cfg := &project.Config{
		Version:     "1.2", // Add required fields
		Name:        "test-project",
		Environment: "development",
		Settings: config.Settings{
			LogLevel:   "info",
			AutoUpdate: true,
			// UpdateInterval intentionally left empty
		},
	}

	err := cfg.Validate() // Use the public Validate method
	if err != nil {
		t.Errorf("Expected no error for empty update interval, got %v", err)
	}
}

func TestConfig_ValidateDependencies_Empty(t *testing.T) {
	cfg := &project.Config{
		Version:     "1.2",
		Name:        "test-project",
		Environment: "development",
		Settings: config.Settings{
			LogLevel:   "info",
			AutoUpdate: true,
		},
		Dependencies: []string{}, // Empty dependencies should be valid
	}

	err := cfg.Validate() // Use the public Validate method
	if err != nil {
		t.Errorf("Expected no error for empty dependencies, got %v", err)
	}
}

func TestValidateConfig(t *testing.T) {
	cfg := &project.Config{
		Version:     "1.0",
		Name:        "test-project",
		Environment: "development",
		Settings: config.Settings{ // Add required settings
			LogLevel:   "info",
			AutoUpdate: false,
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}
}
