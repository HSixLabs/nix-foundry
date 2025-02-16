package config

import (
	"path/filepath"
	"testing"
)

func BenchmarkConfigOperations(b *testing.B) {
	tmpDir := b.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	service := &ServiceImpl{
		path:   configPath,
		config: NewDefaultConfig(),
	}

	// Initial save for subsequent operations
	if err := service.Save(); err != nil {
		b.Fatalf("Failed to save initial config: %v", err)
	}

	b.Run("Load", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := service.Load(); err != nil {
				b.Fatalf("Load failed: %v", err)
			}
		}
	})

	b.Run("Save", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := service.Save(); err != nil {
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
			if err := service.config.Validate(); err != nil {
				b.Fatalf("Validate failed: %v", err)
			}
		}
	})

	b.Run("LoadSection", func(b *testing.B) {
		var settings Settings
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
			if err := service.Load(); err != nil {
				b.Fatalf("Load failed: %v", err)
			}

			// Modify value
			if err := service.SetValue("backup.maxBackups", "15"); err != nil {
				b.Fatalf("SetValue failed: %v", err)
			}

			// Save changes
			if err := service.Save(); err != nil {
				b.Fatalf("Save failed: %v", err)
			}
		}
	})
}
