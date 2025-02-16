package environment_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/test/internal/services/environment/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	tmpDir := t.TempDir()

	cfgSvc := config.NewService()
	platformSvc := &mocks.MockPlatformService{}

	svc := environment.NewService(
		tmpDir,
		cfgSvc,
		platformSvc,
	)

	assert.NotNil(t, svc, "Service should not be nil")
}

func TestEnvironmentOperations(t *testing.T) {
	tmpDir := t.TempDir()
	cfgSvc := config.NewService()
	platformSvc := &mocks.MockPlatformService{}

	svc := environment.NewService(
		tmpDir,
		cfgSvc,
		platformSvc,
	)

	t.Run("initialization", func(t *testing.T) {
		err := svc.Initialize(true)
		if err != nil {
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

	t.Run("environment switching", func(t *testing.T) {
		// Create test environment
		testEnv := filepath.Join(tmpDir, "environments", "test-env")
		if err := os.MkdirAll(testEnv, 0755); err != nil {
			t.Fatalf("Failed to create test environment: %v", err)
		}

		err := svc.Switch("test-env", false)
		if err != nil {
			t.Errorf("Failed to switch environment: %v", err)
		}

		current, err := svc.GetCurrentEnvironment()
		if err != nil {
			t.Errorf("Failed to get current environment: %v", err)
		}

		if current != "test-env" {
			t.Errorf("Expected current environment to be test-env, got %s", current)
		}
	})

	t.Run("environment creation", func(t *testing.T) {
		err := svc.CreateEnvironment("new-env", "default")
		if err != nil {
			t.Errorf("Failed to create environment: %v", err)
		}

		newEnvPath := filepath.Join(tmpDir, "environments", "new-env")
		if _, err := os.Stat(newEnvPath); os.IsNotExist(err) {
			t.Error("New environment directory was not created")
		}

		// Check for required files
		requiredFiles := []string{"flake.nix", "home.nix"}
		for _, file := range requiredFiles {
			filePath := filepath.Join(newEnvPath, file)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Errorf("Required file %s was not created", file)
			}
		}
	})
}

func TestCheckPrerequisites(t *testing.T) {
	platformSvc := &mocks.MockPlatformService{}
	platformSvc.On("CheckPrerequisites", true).Return(nil)

	svc := environment.NewService(
		"/test/config/dir",
		config.NewService(),
		platformSvc,
	)

	t.Run("test mode", func(t *testing.T) {
		if err := svc.CheckPrerequisites(true); err != nil {
			t.Errorf("Expected no error in test mode, got %v", err)
		}
	})
}

func TestSetupEnvironmentSymlink(t *testing.T) {
	tmpDir := t.TempDir()
	platformSvc := &mocks.MockPlatformService{}
	platformSvc.On("SetupEnvironmentSymlink").Return(nil)
	platformSvc.On("Initialize", true).Return(nil)

	svc := environment.NewService(
		tmpDir,
		config.NewService(),
		platformSvc,
	)

	// Create required directories
	defaultEnv := filepath.Join(tmpDir, "environments", "default")
	if err := os.MkdirAll(defaultEnv, 0755); err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}

	// Test initialization
	if err := svc.Initialize(true); err != nil {
		t.Fatalf("Failed to initialize: %v", err)
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
}

func TestInitialize(t *testing.T) {
	tmpDir := t.TempDir()
	platformSvc := &mocks.MockPlatformService{}
	platformSvc.On("Initialize", true).Return(nil)

	svc := environment.NewService(
		tmpDir,
		config.NewService(),
		platformSvc,
	)

	t.Run("successful initialization", func(t *testing.T) {
		if err := svc.Initialize(true); err != nil {
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

		// Use mock platform service
		platformSvc := &mocks.MockPlatformService{}
		platformSvc.On("Initialize", true).Return(fmt.Errorf("directory creation failed"))

		svc := environment.NewService(
			readOnlyDir,
			config.NewService(),
			platformSvc,
		)

		err := svc.Initialize(true)
		if err == nil {
			t.Error("Expected error when directory creation fails")
		}
	})
}

// Add new test functions for isolation setup
func TestSetupIsolation(t *testing.T) {
	tmpDir := t.TempDir()
	platformSvc := &mocks.MockPlatformService{}
	platformSvc.On("SetupIsolation", true).Return(nil)

	svc := environment.NewService(
		tmpDir,
		config.NewService(),
		platformSvc,
	)

	t.Run("test mode", func(t *testing.T) {
		if err := svc.SetupIsolation(true); err != nil {
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

	cfgSvc := config.NewService()
	platformSvc := &mocks.MockPlatformService{}

	svc := environment.NewService(
		tmpDir,
		cfgSvc,
		platformSvc,
	)

	t.Run("create new files", func(t *testing.T) {
		if err := svc.InitializeNixFlake(); err != nil {
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

		if err := svc.InitializeNixFlake(); err != nil {
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

	platformSvc := &mocks.MockPlatformService{}
	platformSvc.On("EnableFlakeFeatures").Return(nil)

	svc := environment.NewService(
		"/test/config",
		config.NewService(),
		platformSvc,
	)

	t.Run("enable flake features", func(t *testing.T) {
		if err := svc.EnableFlakeFeatures(); err != nil {
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

func TestGetCurrentEnvironment(t *testing.T) {
	tmpDir := t.TempDir()
	platformSvc := &mocks.MockPlatformService{}
	platformSvc.On("GetCurrentEnvironment").Return("default", nil)

	svc := environment.NewService(
		tmpDir,
		config.NewService(),
		platformSvc,
	)

	env, err := svc.GetCurrentEnvironment()
	if err != nil {
		t.Errorf("GetCurrentEnvironment() error = %v", err)
	}
	if env != "default" {
		t.Errorf("Expected environment 'default', got %s", env)
	}
}
