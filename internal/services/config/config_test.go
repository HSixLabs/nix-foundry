package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Version: "1.0",
				Settings: Settings{
					LogLevel:       "info",
					AutoUpdate:     true,
					UpdateInterval: "24h",
				},
				Backup: BackupSettings{
					MaxBackups:    5,
					RetentionDays: 30,
					BackupDir:     "~/backups",
				},
				Environment: EnvironmentSettings{
					Default:    "development",
					AutoSwitch: true,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			config: Config{
				Version: "1.0",
				Settings: Settings{
					LogLevel: "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "empty version",
			config: Config{
				Settings: Settings{
					LogLevel: "info",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid backup settings",
			config: Config{
				Version: "1.0",
				Settings: Settings{
					LogLevel: "info",
				},
				Backup: BackupSettings{
					MaxBackups:    0,
					RetentionDays: 30,
					BackupDir:     "~/backups",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid environment",
			config: Config{
				Version: "1.0",
				Settings: Settings{
					LogLevel: "info",
				},
				Environment: EnvironmentSettings{
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
	configPath := filepath.Join(tmpDir, "config.yaml")

	service := &ServiceImpl{
		path:   configPath,
		config: NewDefaultConfig(),
	}

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

		if err := service.Save(); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		newService := &ServiceImpl{
			path: configPath,
		}

		if err := newService.Load(); err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		value, err := newService.GetValue("settings.autoUpdate")
		if err != nil {
			t.Fatalf("GetValue() error = %v", err)
		}
		if value != false {
			t.Errorf("After load, GetValue() = %v, want %v", value, false)
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
		nonExistentService := &ServiceImpl{
			path: filepath.Join(tmpDir, "nonexistent.yaml"),
		}
		err := nonExistentService.Load()
		if err == nil {
			t.Error("Load() with non-existent file should return error")
		}

		// Test saving to read-only directory
		readOnlyDir := filepath.Join(tmpDir, "readonly")
		if err := os.MkdirAll(readOnlyDir, 0444); err != nil {
			t.Fatalf("Failed to create read-only directory: %v", err)
		}
		readOnlyService := &ServiceImpl{
			path:   filepath.Join(readOnlyDir, "config.yaml"),
			config: NewDefaultConfig(),
		}
		err = readOnlyService.Save()
		if err == nil {
			t.Error("Save() to read-only directory should return error")
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

		invalidService := &ServiceImpl{
			path: invalidConfigPath,
		}
		err := invalidService.Load()
		if err == nil {
			t.Error("Load() with invalid config should return validation error")
		}
	})

	t.Run("section operations", func(t *testing.T) {
		var settings Settings
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
