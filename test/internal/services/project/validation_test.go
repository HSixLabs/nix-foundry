package project

import (
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/validation/project"
	configtypes "github.com/shawnkhoffman/nix-foundry/internal/services/config/types"
)

// TestConfig_Validate tests the Config.Validate method
func TestConfig_Validate(t *testing.T) {
	svcConfig := &configtypes.ProjectConfig{
		Version:     "1.0",
		Name:        "test",
		Environment: "development",
		Settings: map[string]string{
			"LogLevel":   "info",
			"AutoUpdate": "true",
		},
	}

	err := project.ValidateConfig(&configtypes.Config{
		Project: *svcConfig,
	})
	if err != nil {
		t.Errorf("Expected valid config to pass validation: %v", err)
	}
}

// Test cases for config validation
func TestConfig_ValidateSettings_EmptyInterval(t *testing.T) {
	cfg := &configtypes.Config{
		Project: configtypes.ProjectConfig{
			Version:     "1.2",
			Name:        "test-project",
			Environment: "development",
		},
	}

	err := project.ValidateConfig(cfg)
	if err != nil {
		t.Errorf("Expected no error for empty update interval, got %v", err)
	}
}

func TestConfig_ValidateDependencies_Empty(t *testing.T) {
	cfg := &configtypes.Config{
		Project: configtypes.ProjectConfig{
			Version:      "1.2",
			Name:         "test-project",
			Environment:  "development",
			Dependencies: []string{},
		},
	}

	err := project.ValidateConfig(cfg)
	if err != nil {
		t.Errorf("Expected no error for empty dependencies, got %v", err)
	}
}

func TestValidateConfig(t *testing.T) {
	cfg := &configtypes.Config{
		Project: configtypes.ProjectConfig{
			Version:     "1.0",
			Name:        "test-project",
			Environment: "development",
		},
	}

	err := project.ValidateConfig(cfg)
	if err != nil {
		t.Errorf("Expected no error for valid config, got: %v", err)
	}
}
