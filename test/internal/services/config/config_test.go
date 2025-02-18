package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/types"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
)

var testConfig = &types.Config{
	Version: "1.0",
	NixConfig: &types.NixConfig{
		Version: "1.0",
	},
	Settings: types.Settings{
		AutoUpdate: true,
		LogLevel:   "info",
	},
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  types.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: types.Config{
				Version: "1.0",
				Settings: types.Settings{
					LogLevel:       "info",
					AutoUpdate:     true,
					UpdateInterval: "24h",
				},
				Backup: types.BackupSettings{
					MaxBackups:    5,
					RetentionDays: 30,
					BackupDir:     "~/backups",
				},
				Environment: types.EnvironmentSettings{
					Default:    "development",
					AutoSwitch: true,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			config: types.Config{
				Version: "1.0",
				Settings: types.Settings{
					LogLevel: "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "empty version",
			config: types.Config{
				Settings: types.Settings{
					LogLevel: "info",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid backup settings",
			config: types.Config{
				Version: "1.0",
				Settings: types.Settings{
					LogLevel: "info",
				},
				Backup: types.BackupSettings{
					MaxBackups:    0,
					RetentionDays: 30,
					BackupDir:     "~/backups",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid environment",
			config: types.Config{
				Version: "1.0",
				Settings: types.Settings{
					LogLevel: "info",
				},
				Environment: types.EnvironmentSettings{
					Default: "invalid",
				},
			},
			wantErr: true,
		},
		// Add more test cases...
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

func TestServiceImpl(t *testing.T) {
	tmpDir := t.TempDir()

	// Set environment variable for config path
	os.Setenv("NIX_FOUNDRY_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("NIX_FOUNDRY_CONFIG_DIR")

	// Create a new service instance
	service := config.NewService()

	t.Run("set and get value", func(t *testing.T) {
		if err := service.SetValue("settings.logLevel", "debug"); err != nil {
			t.Fatalf("SetValue() error = %v", err)
		}

		value, err := service.GetValue("settings.logLevel")
		if err != nil {
			t.Fatalf("GetValue() error = %v", err)
		}

		if value != "debug" {
			t.Errorf("GetValue() = %v, want %v", value, "debug")
		}
	})

	t.Run("reset configuration", func(t *testing.T) {
		if err := service.SetValue("settings.logLevel", "debug"); err != nil {
			t.Fatalf("SetValue() error = %v", err)
		}

		if err := service.Reset(""); err != nil {
			t.Fatalf("Reset() error = %v", err)
		}

		value, err := service.GetValue("settings.logLevel")
		if err != nil {
			t.Fatalf("GetValue() error = %v", err)
		}
		if value != "info" {
			t.Errorf("After reset, GetValue() = %v, want %v", value, "info")
		}
	})

	t.Run("reset specific section", func(t *testing.T) {
		if err := service.SetValue("backup.maxBackups", "20"); err != nil {
			t.Fatalf("SetValue() error = %v", err)
		}

		if err := service.Reset("backup"); err != nil {
			t.Fatalf("Reset(backup) error = %v", err)
		}

		value, err := service.GetValue("backup.maxBackups")
		if err != nil {
			t.Fatalf("GetValue() error = %v", err)
		}
		if value != 10 {
			t.Errorf("After reset, GetValue() = %v, want %v", value, 10)
		}
	})

	t.Run("save and load", func(t *testing.T) {
		if err := service.SetValue("settings.autoUpdate", "false"); err != nil {
			t.Fatalf("SetValue() error = %v", err)
		}

		if err := service.Save(testConfig); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		newService := config.NewService()

		_, err := newService.Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		val, err := newService.GetValue("settings.autoUpdate")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if val != "false" {
			t.Errorf("After load, GetValue() = %v, want %v", val, "false")
		}
	})

	t.Run("invalid path", func(t *testing.T) {
		_, err := service.GetValue("invalid.path")
		if err == nil {
			t.Error("GetValue() with invalid path should return error")
		}

		err = service.SetValue("invalid.path", "value")
		if err == nil {
			t.Error("SetValue() with invalid path should return error")
		}
	})

	t.Run("type conversion errors", func(t *testing.T) {
		// Try to set string to int field
		err := service.SetValue("backup.maxBackups", "not-a-number")
		if err == nil {
			t.Error("SetValue() with invalid type conversion should return error")
		}

		// Try to set string to bool field
		err = service.SetValue("settings.autoUpdate", "not-a-bool")
		if err == nil {
			t.Error("SetValue() with invalid bool conversion should return error")
		}
	})

	t.Run("file operations", func(t *testing.T) {
		// Test loading non-existent file
		nonExistentService := config.NewService()
		_, err := nonExistentService.Load()
		if err == nil {
			t.Error("Load() with non-existent file should return error")
		}

		// Test saving to read-only directory
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		mkdirErr := os.MkdirAll(readOnlyDir, 0444)
		if mkdirErr != nil {
			t.Fatalf("Failed to create read-only directory: %v", mkdirErr)
		}
		readOnlyService := config.NewService()
		err = readOnlyService.Save(testConfig)
		if err == nil {
			t.Error("expected error for readonly service")
		}
	})

	t.Run("validate after load", func(t *testing.T) {
		// Create invalid config file
		invalidConfig := `
version: "1.0"
settings:
  logLevel: "invalid-level"
  autoUpdate: true
`
		invalidConfigPath := filepath.Join(tmpDir, "invalid.yaml")
		if err := os.WriteFile(invalidConfigPath, []byte(invalidConfig), 0644); err != nil {
			t.Fatalf("Failed to write invalid config: %v", err)
		}

		invalidService := config.NewService()
		_, err := invalidService.Load()
		if err == nil {
			t.Error("Load() with invalid config should return validation error")
		}
	})

	t.Run("section operations", func(t *testing.T) {
		var settings types.Settings
		err := service.LoadSection("settings", &settings)
		if err != nil {
			t.Fatalf("LoadSection() error = %v", err)
		}
		if settings.LogLevel != "info" {
			t.Errorf("LoadSection() settings = %v, want LogLevel=info", settings)
		}

		// Test loading non-existent section
		var dummy struct{}
		err = service.LoadSection("nonexistent", &dummy)
		if err == nil {
			t.Error("LoadSection() with non-existent section should return error")
		}
	})

	// Add more test cases...
}

func TestGetValue(t *testing.T) {
	service := config.NewService()

	// Set up test config
	if err := service.Save(testConfig); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	val, err := service.GetValue("settings.autoUpdate")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}

	if val != "true" {
		t.Errorf("Expected 'true', got %v", val)
	}
}
