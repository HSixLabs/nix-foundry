package project

import (
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
)

// Update the Settings struct to match config.Settings
type Settings struct {
	LogLevel       string `yaml:"logLevel"`
	AutoUpdate     bool   `yaml:"autoUpdate"`
	UpdateInterval string `yaml:"updateInterval"`
}

func TestProjectConfig_Validate(t *testing.T) {
	validConfig := &ProjectConfig{
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
		mutateFunc func(*ProjectConfig)
		wantErrMsg string
	}{
		{
			name: "empty version",
			mutateFunc: func(c *ProjectConfig) {
				c.Version = ""
			},
			wantErrMsg: "invalid version: version is required",
		},
		{
			name: "invalid version",
			mutateFunc: func(c *ProjectConfig) {
				c.Version = "2.0"
			},
			wantErrMsg: "invalid version: unsupported version: 2.0",
		},
		{
			name: "empty name",
			mutateFunc: func(c *ProjectConfig) {
				c.Name = ""
			},
			wantErrMsg: "invalid name: name is required",
		},
		{
			name: "name too long",
			mutateFunc: func(c *ProjectConfig) {
				c.Name = "this-is-a-very-long-project-name-that-exceeds-fifty-characters-limit"
			},
			wantErrMsg: "invalid name: name exceeds maximum length of 50 characters",
		},
		{
			name: "invalid environment",
			mutateFunc: func(c *ProjectConfig) {
				c.Environment = "test"
			},
			wantErrMsg: "invalid environment: invalid environment: test",
		},
		{
			name: "invalid log level",
			mutateFunc: func(c *ProjectConfig) {
				c.Settings.LogLevel = "trace"
			},
			wantErrMsg: "invalid settings: invalid log level: trace",
		},
		{
			name: "invalid update interval",
			mutateFunc: func(c *ProjectConfig) {
				c.Settings.AutoUpdate = true
				c.Settings.UpdateInterval = "invalid"
			},
			wantErrMsg: "invalid settings: invalid update interval format: invalid",
		},
		{
			name: "duplicate dependency",
			mutateFunc: func(c *ProjectConfig) {
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

func TestServiceImpl_ValidateSettingsConflicts(t *testing.T) {
	service := &ServiceImpl{
		projectConfig: &ProjectConfig{
			Settings: config.Settings{
				LogLevel:       "info",
				AutoUpdate:     true,
				UpdateInterval: "24h",
			},
		},
	}

	tests := []struct {
		name       string
		settings   config.Settings
		wantErrMsg string
	}{
		{
			name: "no conflicts",
			settings: config.Settings{
				LogLevel:       "info",
				AutoUpdate:     true,
				UpdateInterval: "24h",
			},
			wantErrMsg: "",
		},
		{
			name: "log level conflict",
			settings: config.Settings{
				LogLevel:       "debug",
				AutoUpdate:     true,
				UpdateInterval: "24h",
			},
			wantErrMsg: "log level settings conflict",
		},
		{
			name: "auto-update conflict",
			settings: config.Settings{
				LogLevel:       "info",
				AutoUpdate:     false,
				UpdateInterval: "24h",
			},
			wantErrMsg: "auto-update settings conflict",
		},
		{
			name: "update interval conflict",
			settings: config.Settings{
				LogLevel:       "info",
				AutoUpdate:     true,
				UpdateInterval: "12h",
			},
			wantErrMsg: "update interval settings conflict",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateSettingsConflicts(tt.settings)
			if tt.wantErrMsg == "" {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				return
			}
			if err == nil {
				t.Error("Expected error, got nil")
				return
			}
			if err.Error() != tt.wantErrMsg {
				t.Errorf("Expected error message %q, got %q", tt.wantErrMsg, err.Error())
			}
		})
	}
}

func TestServiceImpl_ValidateSettingsConflicts_NilConfig(t *testing.T) {
	service := &ServiceImpl{
		projectConfig: nil,
		logger:        logging.GetLogger(),
	}

	settings := config.Settings{
		LogLevel:       "info",
		AutoUpdate:     true,
		UpdateInterval: "24h",
	}

	err := service.validateSettingsConflicts(settings)
	if err == nil {
		t.Error("Expected error for nil config, got nil")
		return
	}

	if _, ok := err.(*errors.ConfigError); !ok {
		t.Errorf("Expected ConfigError, got %T", err)
	}
}

func TestProjectConfig_ValidateSettings_EmptyInterval(t *testing.T) {
	cfg := &ProjectConfig{
		Settings: config.Settings{
			LogLevel:   "info",
			AutoUpdate: true,
			// UpdateInterval intentionally left empty
		},
	}

	err := cfg.validateSettings()
	if err != nil {
		t.Errorf("Expected no error for empty update interval, got %v", err)
	}
}

func TestProjectConfig_ValidateDependencies_Empty(t *testing.T) {
	cfg := &ProjectConfig{
		Version:     "1.2",
		Name:        "test-project",
		Environment: "development",
		Settings: config.Settings{
			LogLevel:   "info",
			AutoUpdate: true,
		},
		Dependencies: []string{}, // Empty dependencies should be valid
	}

	if err := cfg.validateDependencies(); err != nil {
		t.Errorf("Expected no error for empty dependencies, got %v", err)
	}
}
