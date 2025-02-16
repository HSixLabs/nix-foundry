package environment

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
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
	GetCurrentEnvironment() (string, error)
	CheckHealth() string
	ListEnvironments() []string
	CreateEnvironment(name string, template string) error
	SetupEnvironmentSymlink() error
	EnableFlakeFeatures() error
	InitializeNixFlake() error
}

type ServiceImpl struct {
	configDir       string
	configService   config.Service
	platformService platform.Service
	logger          *logging.Logger
	currentEnvPath  string
	environments    map[string]string
}

func NewService(
	configDir string,
	cfgSvc config.Service,
	platformSvc platform.Service,
) Service {
	return &ServiceImpl{
		configDir:       configDir,
		configService:   cfgSvc,
		platformService: platformSvc,
		logger:          logging.GetLogger(),
		currentEnvPath:  filepath.Join(configDir, "environments", "current"),
		environments:    make(map[string]string), // Direct initialization
	}
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
	if err := s.platformService.ValidateBackup(backupPath); err != nil {
		return err
	}

	// Restore files using platform service
	return s.platformService.RestoreFromBackup(backupPath, s.currentEnvPath)
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
	s.logger.Info("Switching environment", "target", target, "force", force)

	// Validate target environment exists
	targetPath := filepath.Join(s.configDir, "environments", target)
	if _, err := os.Stat(targetPath); err != nil {
		return errors.NewValidationError(targetPath, err, "target environment not found")
	}

	// Get current environment for comparison
	currentEnv, err := s.GetCurrentEnvironment()
	if err != nil {
		return fmt.Errorf("failed to get current environment: %w", err)
	}

	// Skip if already on target environment
	if currentEnv == targetPath {
		s.logger.Info("Already on target environment", "target", target)
		return nil
	}

	// Check for conflicts if not forcing
	if !force {
		if err := s.checkSwitchConflicts(currentEnv); err != nil {
			return fmt.Errorf("switch conflicts detected: %w", err)
		}
	}

	// Create backup before switch
	backupPath := currentEnv + ".pre-switch"
	if err := s.createBackup(currentEnv, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Perform the switch
	if err := s.performSwitch(targetPath); err != nil {
		// Attempt to restore from backup on failure
		if restoreErr := s.restoreFromBackup(backupPath, currentEnv); restoreErr != nil {
			return fmt.Errorf("switch failed: %v, restore failed: %v", err, restoreErr)
		}
		return fmt.Errorf("failed to switch environment: %w", err)
	}

	// Clean up backup
	if err := os.RemoveAll(backupPath); err != nil {
		s.logger.Warn("Failed to clean up backup", "path", backupPath, "error", err)
	}

	s.logger.Info("Successfully switched environment", "target", target)
	return nil
}

func (s *ServiceImpl) ListEnvironments() []string {
	s.logger.Debug("Listing environments")
	var environments []string

	// Read environments directory
	envsDir := filepath.Join(s.configDir, "environments")
	entries, err := os.ReadDir(envsDir)
	if err != nil {
		s.logger.Error("Failed to read environments directory", "error", err)
		return environments
	}

	// Filter valid environments
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "current" {
			continue
		}

		envPath := filepath.Join(envsDir, entry.Name())
		if err := s.validateEnvironmentStructure(envPath); err != nil {
			s.logger.Warn("Invalid environment found", "path", envPath, "error", err)
			continue
		}

		environments = append(environments, entry.Name())
	}

	return environments
}

func (s *ServiceImpl) CheckHealth() string {
	s.logger.Debug("Checking environment health")

	// Get current environment
	currentEnv, err := s.GetCurrentEnvironment()
	if err != nil {
		return "ERROR: " + err.Error()
	}

	// Check required files
	if err := s.validateEnvironmentStructure(currentEnv); err != nil {
		return "UNHEALTHY: " + err.Error()
	}

	// Check Nix environment
	cmd := exec.Command("nix", "flake", "check")
	cmd.Dir = currentEnv
	if err := cmd.Run(); err != nil {
		return "DEGRADED: Nix flake check failed"
	}

	return "HEALTHY"
}

func (s *ServiceImpl) GetCurrentEnvironment() (string, error) {
	env, err := s.getActiveEnvironment()
	if err != nil {
		return "", fmt.Errorf("failed to get active environment: %w", err)
	}
	return env, nil
}

func (s *ServiceImpl) CreateEnvironment(name string, template string) error {
	s.logger.Info("Creating environment", "name", name, "template", template)

	// Validate environment name
	if name == "" {
		return errors.NewValidationError("", fmt.Errorf("empty name"), "environment name cannot be empty")
	}
	if name == "current" {
		return errors.NewValidationError("", fmt.Errorf("reserved name"), "cannot use 'current' as environment name")
	}

	// Check if environment already exists
	envPath := filepath.Join(s.configDir, "environments", name)
	if _, err := os.Stat(envPath); err == nil {
		return errors.NewValidationError(envPath, fmt.Errorf("environment exists"), "environment already exists")
	}

	// Create environment using template
	if err := s.setupEnvironment(name); err != nil {
		return fmt.Errorf("failed to setup environment: %w", err)
	}

	s.logger.Info("Environment created successfully", "name", name)
	return nil
}

func (s *ServiceImpl) SwitchEnvironment(name string) error {
	// Get current active environment first
	currentEnv, err := s.getActiveEnvironment()
	if err != nil {
		s.logger.Warn("Failed to get active environment", "error", err)
	} else if currentEnv == name {
		s.logger.Info("Already in requested environment", "environment", name)
		return nil
	}

	// Check if environment exists
	if _, err := s.getEnvironmentPath(name); err != nil {
		return fmt.Errorf("environment does not exist: %w", err)
	}

	// Update symlink
	currentEnv = filepath.Join(s.configDir, "environments", "current")
	targetEnv := filepath.Join(s.configDir, "environments", name)

	// Remove existing symlink
	if err := os.Remove(currentEnv); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing symlink: %w", err)
	}

	// Create new symlink
	if err := os.Symlink(targetEnv, currentEnv); err != nil {
		return fmt.Errorf("failed to create environment symlink: %w", err)
	}

	s.logger.Info("Switched environment", "from", currentEnv, "to", name)
	return nil
}

func (s *ServiceImpl) Rollback(targetTime time.Time, force bool) error {
	s.logger.Info("Rolling back environment", "target_time", targetTime)

	// Get backup directory path
	backupDir := filepath.Join(s.configDir, "backups")
	targetBackup := filepath.Join(backupDir, targetTime.Format("20060102-150405"))

	// Check if backup exists
	if _, err := os.Stat(targetBackup); err != nil {
		return errors.NewLoadError(targetBackup, err, "backup not found")
	}

	// If not forcing, check for conflicts
	if !force {
		if err := s.checkRollbackConflicts(targetBackup); err != nil {
			return fmt.Errorf("rollback conflicts detected: %w", err)
		}
	}

	// Create a backup of current environment before rollback
	currentEnv := filepath.Join(s.configDir, "environments", "current")
	backupPath := currentEnv + ".pre-rollback"
	if err := os.Rename(currentEnv, backupPath); err != nil {
		return errors.NewLoadError(currentEnv, err, "failed to backup current environment")
	}

	// Attempt to restore from backup
	if err := s.RestoreEnvironment(targetBackup); err != nil {
		// If restore fails, attempt to restore the pre-rollback backup
		if rollbackErr := os.Rename(backupPath, currentEnv); rollbackErr != nil {
			return fmt.Errorf("rollback failed: %v, recovery failed: %v", err, rollbackErr)
		}
		return fmt.Errorf("failed to restore from backup: %w", err)
	}

	// Clean up pre-rollback backup
	if err := os.RemoveAll(backupPath); err != nil {
		s.logger.Warn("Failed to clean up pre-rollback backup", "path", backupPath, "error", err)
	}

	return nil
}

func (s *ServiceImpl) EnableFlakeFeatures() error {
	s.logger.Info("Enabling Nix flake features")

	// Create .config/nix directory if it doesn't exist
	nixConfigDir := filepath.Join(os.Getenv("HOME"), ".config", "nix")
	if err := os.MkdirAll(nixConfigDir, 0755); err != nil {
		return errors.NewLoadError(nixConfigDir, err, "failed to create nix config directory")
	}

	// Write nix.conf with flake features enabled
	nixConfPath := filepath.Join(nixConfigDir, "nix.conf")
	content := "experimental-features = nix-command flakes"
	if err := os.WriteFile(nixConfPath, []byte(content), 0644); err != nil {
		return errors.NewLoadError(nixConfPath, err, "failed to write nix.conf")
	}

	return nil
}

func (s *ServiceImpl) InitializeNixFlake() error {
	s.logger.Info("Initializing Nix flake environment")

	defaultEnvPath := filepath.Join(s.configDir, "environments", "default")

	// Enable flakes in Nix configuration
	cmd := exec.Command("nix-env", "--set-flag", "experimental-features", "nix-command flakes")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to enable flake features: %s: %w", string(output), err)
	}

	// Initialize flake
	cmd = exec.Command("nix", "flake", "init")
	cmd.Dir = defaultEnvPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to initialize flake: %s: %w", string(output), err)
	}

	// Update flake lock file
	cmd = exec.Command("nix", "flake", "update")
	cmd.Dir = defaultEnvPath
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to update flake lock: %s: %w", string(output), err)
	}

	return nil
}

func (s *ServiceImpl) SetupEnvironmentSymlink() error {
	s.logger.Info("Setting up environment symlink")

	currentLink := filepath.Join(s.configDir, "environments", "current")
	defaultEnv := filepath.Join(s.configDir, "environments", "default")

	// Remove existing symlink if it exists
	if _, err := os.Lstat(currentLink); err == nil {
		if err := os.Remove(currentLink); err != nil {
			return fmt.Errorf("failed to remove existing symlink: %w", err)
		}
	}

	// Create new symlink pointing to default environment
	if err := os.Symlink(defaultEnv, currentLink); err != nil {
		return fmt.Errorf("failed to create environment symlink: %w", err)
	}

	return nil
}
