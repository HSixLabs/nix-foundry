package environment

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/templates"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/testutils"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/validation"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/shawnkhoffman/nix-foundry/pkg/progress"
)

type Service interface {
	Initialize(testMode bool) error
	CheckPrerequisites(testMode bool) error
	SetupIsolation(testMode bool) error
	InstallBinary() error
	RestoreEnvironment(backupPath string) error
	ValidateRestoredEnvironment(envPath string) error
	Switch(target string, force bool) error
	Rollback(targetTime time.Time, force bool) error
	GetCurrentEnvironment() string
	CheckHealth() string
	ListEnvironments() []string
	CreateEnvironment(name string, template string) error
}

type ServiceImpl struct {
	configDir      string
	logger         *logging.Logger
	configSvc      config.Service
	validator      validation.Service
	platformSvc    platform.Service
	currentEnvPath string
	environments   map[string]string
}

func NewService(
	configDir string,
	cfgSvc config.Service,
	validator validation.Service,
	platform platform.Service,
) Service {
	return &ServiceImpl{
		configDir:      configDir,
		configSvc:      cfgSvc,
		validator:      validator,
		platformSvc:    platform,
		currentEnvPath: filepath.Join(configDir, "environments", "current"),
		environments:   loadEnvironments(),
		logger:         logging.GetLogger(),
	}
}

func (s *ServiceImpl) Initialize(testMode bool) error {
	s.logger.Info("Initializing environment")

	// Check prerequisites first
	if err := s.CheckPrerequisites(testMode); err != nil {
		return err
	}

	// Setup platform-specific requirements
	if err := s.platformSvc.SetupPlatform(testMode); err != nil {
		return err
	}

	// Enable flake features
	if !testMode {
		if err := s.enableFlakeFeatures(); err != nil {
			return err
		}
	}

	// Create required directories
	dirs := []string{
		s.configDir,
		filepath.Join(s.configDir, "environments"),
		filepath.Join(s.configDir, "environments", "default"),
		filepath.Join(s.configDir, "backups"),
		filepath.Join(s.configDir, "logs"),
	}

	for _, dir := range dirs {
		s.logger.Debug("Creating directory", "path", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.NewLoadError(dir, err, "failed to create directory")
		}
	}

	// Setup environment symlink
	if err := s.setupEnvironmentSymlink(); err != nil {
		return err
	}

	if testMode {
		s.logger.Debug("Creating test configuration files")
		files := map[string]string{
			filepath.Join(s.configDir, "flake.nix"): testutils.MockFlakeContent,
			filepath.Join(s.configDir, "home.nix"):  templates.HomeManagerTemplate,
		}
		for file, content := range files {
			if err := os.WriteFile(file, []byte(content), 0644); err != nil {
				return errors.NewLoadError(file, err, "failed to create test file")
			}
		}
		return nil
	}

	return s.SetupIsolation(testMode)
}

func (s *ServiceImpl) CheckPrerequisites(testMode bool) error {
	if testMode {
		s.logger.Debug("Skipping prerequisite checks in test mode")
		return nil
	}

	s.logger.Debug("Checking prerequisites")
	prerequisites := []string{"nix", "home-manager"}

	for _, prereq := range prerequisites {
		if _, err := exec.LookPath(prereq); err != nil {
			if prereq == "nix" {
				return errors.NewValidationError("", err,
					"nix is not installed\ntry running 'curl -L https://nixos.org/nix/install | sh' manually")
			}
			return errors.NewValidationError("", err, fmt.Sprintf("%s is not installed", prereq))
		}
	}

	if _, err := exec.LookPath("nix-channel"); err != nil {
		return errors.NewValidationError("", err,
			"nix setup failed: try running 'curl -L https://nixos.org/nix/install | sh' manually")
	}

	return nil
}

func (s *ServiceImpl) InstallBinary() error {
	s.logger.Info("Installing nix-foundry binary")

	spin := progress.NewSpinner("Installing nix-foundry...")
	spin.Start()
	defer spin.Stop()

	binDir := filepath.Join(os.Getenv("HOME"), ".local", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		spin.Fail("Failed to create bin directory")
		return errors.NewLoadError(binDir, err, "failed to create bin directory")
	}

	executable, err := os.Executable()
	if err != nil {
		spin.Fail("Failed to get executable path")
		return errors.NewLoadError("", err, "failed to get executable path")
	}

	symlink := filepath.Join(binDir, "nix-foundry")
	if err := os.Symlink(executable, symlink); err != nil && !os.IsExist(err) {
		spin.Fail("Failed to create symlink")
		return errors.NewLoadError(symlink, err, "failed to create symlink")
	}

	spin.Success("nix-foundry installed")
	s.logger.Info(fmt.Sprintf("ℹ️  Add %s to your PATH to use nix-foundry from anywhere", binDir))

	return nil
}

func (s *ServiceImpl) getActiveEnvironment() (string, error) {
	s.logger.Debug("Getting active environment")

	// First check for a current symlink
	currentEnv := filepath.Join(s.configDir, "environments", "current")
	if target, err := os.Readlink(currentEnv); err == nil {
		// Return the absolute path of the target
		if filepath.IsAbs(target) {
			return target, nil
		}
		// Convert relative symlink target to absolute path
		return filepath.Join(filepath.Dir(currentEnv), target), nil
	}

	// Fall back to default environment
	defaultEnv := filepath.Join(s.configDir, "environments", "default")
	if _, err := os.Stat(defaultEnv); err != nil {
		return "", errors.NewValidationError(defaultEnv, err, "no active environment found")
	}

	// Return the absolute path
	absPath, err := filepath.Abs(defaultEnv)
	if err != nil {
		return "", errors.NewLoadError(defaultEnv, err, "failed to get absolute path")
	}
	return absPath, nil
}

func (s *ServiceImpl) RestoreEnvironment(backupPath string) error {
	s.logger.Info("Restoring environment from backup", "path", backupPath)

	// Validate backup using platform service instead of validator
	if err := s.platformSvc.ValidateBackup(backupPath); err != nil {
		return err
	}

	// Restore files using platform service
	return s.platformSvc.RestoreFromBackup(backupPath, s.currentEnvPath)
}

func (s *ServiceImpl) ValidateRestoredEnvironment(envPath string) error {
	requiredFiles := []string{"flake.nix", "home.nix"}

	for _, file := range requiredFiles {
		if _, err := os.Stat(filepath.Join(envPath, file)); err != nil {
			s.logger.Error("Missing required file", "file", file)
			return fmt.Errorf("missing required file: %s", file)
		}
	}
	return nil
}

func (s *ServiceImpl) Switch(target string, force bool) error {
	tempLink := filepath.Join(s.configDir, "environments", ".tmp-current")

	// Use struct field instead of local variable
	if err := os.Symlink(s.environments[target], tempLink); err != nil {
		return fmt.Errorf("switch failed: %w", err)
	}

	return os.Rename(tempLink, s.currentEnvPath)
}

func (s *ServiceImpl) enableFlakeFeatures() error {
	s.logger.Info("Enabling Nix flake features")

	cmd := exec.Command("nix", "env", "config", "--experimental-features", "nix-command flakes")
	if output, err := cmd.CombinedOutput(); err != nil {
		s.logger.Error("Failed to enable flake features", "output", string(output))
		return errors.NewOperationError("nix config", err, map[string]interface{}{
			"context": "failed to enable flake features",
		})
	}
	return nil
}

func loadEnvironments() map[string]string {
	return map[string]string{
		"default": filepath.Join("environments", "default"),
		"stable":  filepath.Join("environments", "releases", "stable"),
		"dev":     filepath.Join("environments", "development"),
	}
}

func (s *ServiceImpl) GetCurrentEnvironment() string {
	return filepath.Base(s.currentEnvPath)
}

func (s *ServiceImpl) CheckHealth() string {
	return "healthy" // Simplified health check
}

func (s *ServiceImpl) ListEnvironments() []string {
	environments := make([]string, 0, len(s.environments))
	for env := range s.environments {
		environments = append(environments, env)
	}
	return environments
}

func (s *ServiceImpl) CreateEnvironment(name string, template string) error {
	// Implementation of CreateEnvironment method
	return nil // Placeholder return, actual implementation needed
}
