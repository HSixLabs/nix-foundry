package environment

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/validation"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
)

func TestNewService(t *testing.T) {
	configDir := "/test/config/dir"
	service := NewService(
		configDir,
		config.NewService(),
		validation.NewService(),
		platform.NewService(),
	).(*ServiceImpl)

	if service.configDir != configDir {
		t.Errorf("Expected configDir to be %s, got %s", configDir, service.configDir)
	}

	if service.logger == nil {
		t.Error("Expected logger to be initialized")
	}
}

func TestCheckPrerequisites(t *testing.T) {
	service := NewService(
		"/test/config/dir",
		config.NewService(),
		validation.NewService(),
		platform.NewService(),
	)

	t.Run("test mode", func(t *testing.T) {
		if err := service.CheckPrerequisites(true); err != nil {
			t.Errorf("Expected no error in test mode, got %v", err)
		}
	})
}

func TestSetupEnvironmentSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewService(
		tmpDir,
		config.NewService(),
		validation.NewService(),
		platform.NewService(),
	).(*ServiceImpl)

	// Create required directories
	defaultEnv := filepath.Join(tmpDir, "environments", "default")
	if err := os.MkdirAll(defaultEnv, 0755); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}

	// Test symlink creation
	if err := service.setupEnvironmentSymlink(); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Verify symlink
	currentEnv := filepath.Join(tmpDir, "environments", "current")
	fi, err := os.Lstat(currentEnv)
	if err != nil {
		t.Fatalf("Failed to stat symlink: %v", err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		t.Error("Expected a symlink to be created")
	}

	// Test recreating symlink (should handle existing symlink)
	if err := service.setupEnvironmentSymlink(); err != nil {
		t.Errorf("Failed to recreate symlink: %v", err)
	}
}

func TestInitialize(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewService(
		tmpDir,
		config.NewService(),
		validation.NewService(),
		platform.NewService(),
	)

	t.Run("successful initialization", func(t *testing.T) {
		if err := service.Initialize(true); err != nil {
			t.Errorf("Expected successful initialization, got error: %v", err)
		}

		// Verify directory structure
		expectedDirs := []string{
			tmpDir,
			filepath.Join(tmpDir, "environments"),
			filepath.Join(tmpDir, "environments", "default"),
			filepath.Join(tmpDir, "backups"),
			filepath.Join(tmpDir, "logs"),
		}

		for _, dir := range expectedDirs {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				t.Errorf("Expected directory %s to exist", dir)
			}
		}
	})

	t.Run("directory creation failure", func(t *testing.T) {
		// Create a file where we expect a directory
		readOnlyDir := t.TempDir()
		blockingFile := filepath.Join(readOnlyDir, "environments")
		if err := os.WriteFile(blockingFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create blocking file: %v", err)
		}

		service := NewService(
			readOnlyDir,
			config.NewService(),
			validation.NewService(),
			platform.NewService(),
		)
		err := service.Initialize(true)
		if err == nil {
			t.Error("Expected error when directory creation fails")
		}
	})
}

// Add new test functions for isolation setup
func TestSetupIsolation(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewService(
		tmpDir,
		config.NewService(),
		validation.NewService(),
		platform.NewService(),
	)

	t.Run("test mode", func(t *testing.T) {
		if err := service.(*ServiceImpl).SetupIsolation(true); err != nil {
			t.Errorf("Expected no error in test mode, got %v", err)
		}

		// Verify test files were created
		files := []string{
			filepath.Join(tmpDir, "flake.nix"),
			filepath.Join(tmpDir, "home.nix"),
		}
		for _, file := range files {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				t.Errorf("Expected file %s to exist", file)
			}
		}
	})
}

func TestInitializeNixFlake(t *testing.T) {
	tmpDir := t.TempDir()
	service := NewService(
		tmpDir,
		config.NewService(),
		validation.NewService(),
		platform.NewService(),
	).(*ServiceImpl)

	t.Run("create new files", func(t *testing.T) {
		if err := service.initializeNixFlake(); err != nil {
			t.Fatalf("Failed to initialize nix flake: %v", err)
		}

		// Verify files were created
		files := []string{
			filepath.Join(tmpDir, "flake.nix"),
			filepath.Join(tmpDir, "home.nix"),
		}
		for _, file := range files {
			content, err := os.ReadFile(file)
			if err != nil {
				t.Errorf("Failed to read file %s: %v", file, err)
				continue
			}
			if len(content) == 0 {
				t.Errorf("Expected non-empty content in %s", file)
			}
		}
	})

	t.Run("existing files", func(t *testing.T) {
		// Test that existing files are not overwritten
		flakeContent := "existing flake config"
		homeContent := "existing home config"

		if err := os.WriteFile(filepath.Join(tmpDir, "flake.nix"), []byte(flakeContent), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, "home.nix"), []byte(homeContent), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		if err := service.initializeNixFlake(); err != nil {
			t.Fatalf("Failed to initialize with existing files: %v", err)
		}

		// Verify files were not modified
		content, err := os.ReadFile(filepath.Join(tmpDir, "flake.nix"))
		if err != nil || string(content) != flakeContent {
			t.Error("Existing flake.nix was modified")
		}

		content, err = os.ReadFile(filepath.Join(tmpDir, "home.nix"))
		if err != nil || string(content) != homeContent {
			t.Error("Existing home.nix was modified")
		}
	})
}

func TestEnableFlakeFeatures(t *testing.T) {
	tmpHome := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	service := NewService(
		"/test/config",
		config.NewService(),
		validation.NewService(),
		platform.NewService(),
	).(*ServiceImpl)

	t.Run("enable flake features", func(t *testing.T) {
		if err := service.enableFlakeFeatures(); err != nil {
			t.Fatalf("Failed to enable flake features: %v", err)
		}

		// Verify nix.conf was created with correct content
		content, err := os.ReadFile(filepath.Join(tmpHome, ".config", "nix", "nix.conf"))
		if err != nil {
			t.Fatalf("Failed to read nix.conf: %v", err)
		}

		expected := "experimental-features = nix-command flakes"
		if string(content) != expected {
			t.Errorf("Expected nix.conf content to be %q, got %q", expected, string(content))
		}
	})
}
