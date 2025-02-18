package config

import (
	"os"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	configtypes "github.com/shawnkhoffman/nix-foundry/internal/services/config/types"
)

func BenchmarkConfigOperations(b *testing.B) {
	tmpDir := b.TempDir()

	// Set environment variable for config path
	os.Setenv("NIX_FOUNDRY_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("NIX_FOUNDRY_CONFIG_DIR")

	// Create a new service instance
	service := config.NewService()

	// Create a test config using configtypes
	testConfig := &configtypes.Config{
		Project: configtypes.ProjectConfig{
			Version:     "1.0",
			Name:        "test-project",
			Environment: "development",
			Settings: map[string]string{
				"logLevel":   "info",
				"autoUpdate": "true",
			},
		},
	}

	// Initial save for subsequent operations
	if err := service.Save(testConfig); err != nil {
		b.Fatalf("Failed to save initial config: %v", err)
	}

	b.Run("Load", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := service.Load()
			if err != nil {
				b.Fatalf("Load failed: %v", err)
			}
		}
	})

	b.Run("Save", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := service.Save(testConfig); err != nil {
				b.Fatalf("Save failed: %v", err)
			}
		}
	})

	b.Run("GetValue", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := service.GetValue("backup.maxBackups"); err != nil {
				b.Fatalf("GetValue failed: %v", err)
			}
		}
	})

	b.Run("SetValue", func(b *testing.B) {
		values := []string{"10", "20", "30", "40", "50"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			value := values[i%len(values)]
			if err := service.SetValue("backup.maxBackups", value); err != nil {
				b.Fatalf("SetValue failed: %v", err)
			}
		}
	})

	b.Run("Validate", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := service.Validate(); err != nil {
				b.Fatalf("Validate failed: %v", err)
			}
		}
	})

	b.Run("LoadSection", func(b *testing.B) {
		var settings config.Settings // Use the proper type from the config package
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := service.LoadSection("settings", &settings); err != nil {
				b.Fatalf("LoadSection failed: %v", err)
			}
		}
	})

	b.Run("CompleteLifecycle", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Load config
			_, err := service.Load()
			if err != nil {
				b.Fatalf("Load failed: %v", err)
			}

			// Modify value
			if err := service.SetValue("backup.maxBackups", "15"); err != nil {
				b.Fatalf("SetValue failed: %v", err)
			}

			// Save changes
			if err := service.Save(testConfig); err != nil {
				b.Fatalf("Save failed: %v", err)
			}
		}
	})
}
