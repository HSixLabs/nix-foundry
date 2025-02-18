package environment

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/shawnkhoffman/nix-foundry/pkg/progress"
	"go.uber.org/zap"
)

type Service interface {
	Initialize(testMode bool) error
	CheckPrerequisites(testMode bool) error
	SetupIsolation(testMode bool, force bool) error
	InstallBinary() error
	RestoreEnvironment(backupPath string) error
	ValidateRestoredEnvironment(envPath string) error
	Switch(target string, force bool) error
	Rollback(targetTime time.Time, force bool) error
	GetCurrentEnvironment() (string, error)
	CheckHealth() error
	ListEnvironments() []string
	CreateEnvironment(name string, template string) error
	SetupEnvironmentSymlink() error
	EnableFlakeFeatures() error
	InitializeNixFlake() error
	Validate() error
	ApplyConfiguration() error
	AddPackage(pkg string) error
}

type ServiceImpl struct {
	configDir        string
	configService    config.Service
	platformService  platform.Service
	logger           *logging.Logger
	currentEnvPath   string
	environments     map[string]string
	testMode         bool
	isolationEnabled bool
	autoInstall      bool
}

func NewService(
	configDir string,
	configSvc config.Service,
	platformSvc platform.Service,
	testMode bool,
	isolationEnabled bool,
	autoInstall bool,
) Service {
	// Convert configDir to absolute path
	absConfigDir, err := filepath.Abs(configDir)
	if err != nil {
		panic(fmt.Sprintf("failed to resolve config directory: %v", err))
	}

	// Create required directories
	requiredDirs := []string{
		filepath.Join(absConfigDir, "environments", "default"),
		filepath.Join(absConfigDir, "backups"),
		filepath.Join(absConfigDir, "projects"),
	}

	for _, dir := range requiredDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic(fmt.Sprintf("Failed to create environment directory: %v", err))
		}
	}

	return &ServiceImpl{
		configDir:        absConfigDir,
		configService:    configSvc,
		platformService:  platformSvc,
		logger:           logging.GetLogger().WithComponent("environment"),
		currentEnvPath:   filepath.Join(absConfigDir, "environments", "current"),
		environments:     make(map[string]string),
		testMode:         testMode,
		isolationEnabled: isolationEnabled,
		autoInstall:      autoInstall,
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
	s.logger.Debug("Installing nix-foundry binary")

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

func (s *ServiceImpl) CheckHealth() error {
	s.logger.Debug("Checking environment health")

	// Get current environment
	currentEnv, err := s.GetCurrentEnvironment()
	if err != nil {
		return fmt.Errorf("failed to get current environment: %w", err)
	}

	// Check required files
	if err := s.validateEnvironmentStructure(currentEnv); err != nil {
		return fmt.Errorf("environment is UNHEALTHY: %w", err)
	}

	// Check Nix environment
	cmd := exec.Command("nix", "flake", "check")
	cmd.Dir = currentEnv
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("DEGRADED: Nix flake check failed")
	}

	return nil
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

	// Create environment directory structure with explicit permissions
	envPath := filepath.Join(s.configDir, "environments", "default")
	if err := os.MkdirAll(envPath, 0755); err != nil {
		return fmt.Errorf("failed to create environment directory: %w", err)
	}

	// Create both home.nix and flake.nix with full package definitions
	files := map[string]string{
		"home.nix": `{ config, pkgs, ... }:
{
  home.username = "{{.Username}}";
  home.homeDirectory = "{{.HomeDir}}";
  home.stateVersion = "23.11";

  programs.home-manager.enable = true;

  home.packages = with pkgs; [
    # Add your packages here
  ];

  # Add more home-manager configurations here
}`,
		"flake.nix": `{
  description = "Nix Foundry managed environment";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, home-manager }: let
    systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
    forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f system);
  in {
    packages = forAllSystems (system: {
      default = nixpkgs.legacyPackages.${system}.buildEnv {
        name = "nix-foundry-env";
        paths = [];
      };
    });

    defaultPackage = forAllSystems (system: self.packages.${system}.default);

    homeConfigurations = forAllSystems (system: {
      default = home-manager.lib.homeManagerConfiguration {
        pkgs = nixpkgs.legacyPackages.${system};
        modules = [
          ({ config, ... }: {
            home.username = "{{.Username}}";
            home.homeDirectory = "{{.HomeDir}}";
          })
          ./home.nix
        ];
      };
    });
  };
}`,
	}

	// Replace placeholders with actual values
	username := os.Getenv("USER")
	homeDir := os.Getenv("HOME")

	for filename, content := range files {
		content = strings.ReplaceAll(content, "{{.Username}}", username)
		content = strings.ReplaceAll(content, "{{.HomeDir}}", homeDir)
		filePath := filepath.Join(envPath, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}
	}

	return nil
}

func (s *ServiceImpl) SetupEnvironmentSymlink() error {
	s.logger.Info("Setting up environment symlink")

	currentLink := filepath.Join(s.configDir, "environments", "current")
	defaultEnv := filepath.Join(s.configDir, "environments", "default")
	tempLink := currentLink + ".tmp"

	// Ensure default environment exists and is valid
	if _, err := os.Stat(defaultEnv); os.IsNotExist(err) {
		return fmt.Errorf("default environment missing: %w", err)
	}

	// Clean up any existing temp link (ignore errors)
	_ = os.Remove(tempLink)

	// Create temporary symlink first
	if err := os.Symlink(defaultEnv, tempLink); err != nil {
		return fmt.Errorf("failed to create temporary symlink: %w", err)
	}

	// Atomically replace current symlink
	if err := os.Rename(tempLink, currentLink); err != nil {
		_ = os.Remove(tempLink) // Clean up on failure
		return fmt.Errorf("failed to commit symlink: %w", err)
	}

	// Final verification that symlink points to correct target
	target, err := filepath.EvalSymlinks(currentLink)
	if err != nil {
		return fmt.Errorf("symlink verification failed: %w", err)
	}

	resolvedTarget, _ := filepath.EvalSymlinks(defaultEnv)
	if target != resolvedTarget {
		return fmt.Errorf("symlink target mismatch: expected %s, got %s", resolvedTarget, target)
	}

	s.logger.Info("Environment symlink successfully created",
		"symlink", currentLink,
		"target", resolvedTarget)
	return nil
}

func (s *ServiceImpl) Initialize(testMode bool) error {
	s.logger.Debug("Initializing environment service", zap.String("component", "environment"))

	// First create core directory structure
	if err := s.createDirectoryStructure(); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Enable flake features before any nix operations
	if err := s.EnableFlakeFeatures(); err != nil {
		return fmt.Errorf("failed to enable flake features: %w", err)
	}

	// Initialize flake first
	if err := s.InitializeNixFlake(); err != nil {
		return fmt.Errorf("failed to initialize Nix flake: %w", err)
	}

	// Then setup isolation with proper path resolution
	if s.isolationEnabled {
		envPath, err := s.getEnvironmentPath("default")
		if err != nil {
			return fmt.Errorf("isolation setup failed: %w", err)
		}
		if err := s.initializeFlake(envPath, false); err != nil {
			return fmt.Errorf("isolation setup failed: %w", err)
		}
	}

	return nil
}

// Add retry mechanism
func retry(attempts int, delay time.Duration, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		if i > 0 {
			time.Sleep(delay)
		}
		if err = fn(); err == nil {
			return nil
		}
	}
	return fmt.Errorf("after %d attempts: %w", attempts, err)
}

func (s *ServiceImpl) ApplyConfiguration() error {
	s.logger.Debug("Applying environment configuration", zap.String("component", "environment"))

	envPath, err := s.getEnvironmentPath("default")
	if err != nil {
		return fmt.Errorf("failed to get environment path: %w", err)
	}

	// Correct system architecture mapping
	systemArch := runtime.GOARCH
	if systemArch == "arm64" {
		systemArch = "aarch64"
	}
	systemOS := strings.ToLower(runtime.GOOS)
	nixSystem := fmt.Sprintf("%s-%s", systemArch, systemOS)

	// Use a simpler attribute path
	cmd := exec.Command("nix", "build", "--show-trace", fmt.Sprintf(".#defaultPackage.%s", nixSystem))
	cmd.Dir = envPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build environment: %w", err)
	}

	s.logger.Info("Successfully applied configuration")
	return nil
}

func (s *ServiceImpl) AddPackage(pkg string) error {
	s.logger.Debug("Starting package addition", "package", pkg)

	envPath, err := s.getEnvironmentPath("default")
	if err != nil {
		return fmt.Errorf("failed to get environment path: %w", err)
	}

	homeNixPath := filepath.Join(envPath, "home.nix")
	content, err := os.ReadFile(homeNixPath)
	if err != nil {
		return fmt.Errorf("failed to read home.nix: %w", err)
	}

	// Find the packages array and append the new package
	updatedContent := regexp.MustCompile(`home\.packages = with pkgs; \[\n([\s\S]*?)\n\s*\];`).
		ReplaceAllString(string(content),
			"home.packages = with pkgs; [\n$1    ${pkg}\n  ];")

	if err := os.WriteFile(homeNixPath, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("failed to update home.nix: %w", err)
	}

	s.logger.Debug("Updated home.nix configuration")
	return nil
}
