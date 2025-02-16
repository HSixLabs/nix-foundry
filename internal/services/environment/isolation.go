package environment

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/types"
)

// SetupIsolation initializes an isolated environment structure
func (s *ServiceImpl) SetupIsolation(testMode bool) error {
	if testMode {
		s.logger.Debug("Skipping isolation setup in test mode")
		return nil
	}

	// Validate core dependencies first
	if err := s.validateCoreDependencies(); err != nil {
		return fmt.Errorf("dependency check failed: %w", err)
	}

	// Load existing environments
	s.environments = s.loadExistingEnvironments()

	// Create required directory structure
	if err := s.createDirectoryStructure(); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Initialize default environment if it doesn't exist
	if _, exists := s.environments["default"]; !exists {
		if err := s.setupEnvironment("default"); err != nil {
			return fmt.Errorf("failed to setup default environment: %w", err)
		}
	}

	// Setup environment symlink
	if err := s.setupEnvironmentSymlink(); err != nil {
		return fmt.Errorf("failed to setup environment symlink: %w", err)
	}

	// Initialize Nix flake
	if err := s.initializeNixFlake(); err != nil {
		return fmt.Errorf("failed to initialize Nix flake: %w", err)
	}

	return nil
}

// validateEnvironmentStructure checks if an environment has the required structure
func (s *ServiceImpl) validateEnvironmentStructure(envPath string) error {
	s.logger.Debug("Validating environment structure", "path", envPath)

	requiredFiles := []string{"flake.nix", "home.nix"}
	for _, file := range requiredFiles {
		if _, err := os.Stat(filepath.Join(envPath, file)); err != nil {
			s.logger.Error("Missing required file", "file", file)
			return fmt.Errorf("missing required file: %s", file)
		}
	}

	// Verify directory permissions
	info, err := os.Stat(envPath)
	if err != nil {
		return fmt.Errorf("failed to access environment directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("environment path is not a directory: %s", envPath)
	}
	if info.Mode().Perm()&0200 == 0 {
		return fmt.Errorf("environment directory is not writable: %s", envPath)
	}

	return nil
}

// loadExistingEnvironments scans for and loads existing environment configurations
func (s *ServiceImpl) loadExistingEnvironments() map[string]string {
	environments := make(map[string]string)
	envsDir := filepath.Join(s.configDir, "environments")

	entries, err := os.ReadDir(envsDir)
	if err != nil {
		if !os.IsNotExist(err) {
			s.logger.Warn("Failed to read environments directory", "error", err)
		}
		return environments
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "current" {
			continue
		}

		envPath := filepath.Join(envsDir, entry.Name())
		if err := s.validateEnvironmentStructure(envPath); err != nil {
			s.logger.Warn("Invalid environment found", "path", envPath, "error", err)
			continue
		}

		environments[entry.Name()] = envPath
	}

	return environments
}

// createDirectoryStructure creates the required directory hierarchy
func (s *ServiceImpl) createDirectoryStructure() error {
	dirs := []string{
		filepath.Join(s.configDir, "environments"),
		filepath.Join(s.configDir, "environments", "default"),
		filepath.Join(s.configDir, "storage"),
		filepath.Join(s.configDir, "cache"),
		filepath.Join(s.configDir, "backups"),
	}

	for _, dir := range dirs {
		if fi, err := os.Stat(dir); err == nil {
			if !fi.IsDir() {
				return fmt.Errorf("path exists but is not a directory: %s", dir)
			}
			continue
		}

		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// validateCoreDependencies checks if required tools are available
func (s *ServiceImpl) validateCoreDependencies() error {
	requiredTools := map[string]string{
		"nix":          "Nix package manager",
		"home-manager": "Home Manager",
		"git":          "Git version control",
	}

	for cmd, desc := range requiredTools {
		if _, err := exec.LookPath(cmd); err != nil {
			return fmt.Errorf("%s not found: %w", desc, err)
		}
	}
	return nil
}

// setupEnvironment creates a new environment with the given name
func (s *ServiceImpl) setupEnvironment(name string) error {
	s.logger.Info("Setting up environment", "name", name)
	envPath := filepath.Join(s.configDir, "environments", name)

	// Create environment directory
	if err := os.MkdirAll(envPath, 0755); err != nil {
		return fmt.Errorf("failed to create environment directory: %w", err)
	}

	// Create basic environment files
	files := map[string]string{
		"flake.nix": `{
  description = "Nix Foundry managed environment";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, home-manager, ... }: {
    defaultPackage.x86_64-linux = self.packages.x86_64-linux.default;
    defaultPackage.x86_64-darwin = self.packages.x86_64-darwin.default;

    packages.x86_64-linux.default = nixpkgs.legacyPackages.x86_64-linux.buildEnv {
      name = "nix-foundry-env";
      paths = [];
    };

    packages.x86_64-darwin.default = nixpkgs.legacyPackages.x86_64-darwin.buildEnv {
      name = "nix-foundry-env";
      paths = [];
    };
  };
}`,
		"home.nix": `{ config, pkgs, ... }:

{
  home.username = builtins.getEnv "USER";
  home.homeDirectory = builtins.getEnv "HOME";
  home.stateVersion = "23.11";

  programs.home-manager.enable = true;

  home.packages = with pkgs; [
    # Add your packages here
  ];

  # Add more home-manager configurations here
}`,
	}

	for filename, content := range files {
		filePath := filepath.Join(envPath, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}
	}

	// Initialize git repository for version control
	cmd := exec.Command("git", "init")
	cmd.Dir = envPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	s.environments[name] = envPath
	return nil
}

// setupEnvironmentSymlink creates the current environment symlink
func (s *ServiceImpl) setupEnvironmentSymlink() error {
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

// initializeNixFlake sets up the Nix flake environment
func (s *ServiceImpl) initializeNixFlake() error {
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

// getEnvironmentPath returns the full path to an environment
func (s *ServiceImpl) getEnvironmentPath(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("environment name cannot be empty")
	}

	path := filepath.Join(s.configDir, "environments", name)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("environment %s does not exist", name)
		}
		return "", fmt.Errorf("failed to access environment %s: %w", name, err)
	}

	return path, nil
}

func (s *ServiceImpl) teardownEnvironment(env string) error {
	envPath, err := s.getEnvironmentPath(env)
	if err != nil {
		return fmt.Errorf("environment teardown failed: %w", err)
	}

	// Remove environment directory
	if err := os.RemoveAll(envPath); err != nil {
		return fmt.Errorf("failed to remove environment directory: %w", err)
	}

	// Update environments list
	delete(s.environments, env)
	s.logger.Info("Torn down environment", "environment", env)

	return nil
}

func (s *ServiceImpl) Cleanup(env string) error {
	if env == "default" {
		return fmt.Errorf("cannot remove default environment")
	}
	return s.teardownEnvironment(env)
}

// Update Initialize method to simplify error handling
func (s *ServiceImpl) Initialize(testMode bool) error {
	s.logger.Info("Initializing environment")

	// Check prerequisites first
	if err := s.CheckPrerequisites(testMode); err != nil {
		return err
	}

	// Setup platform-specific requirements
	if err := s.platformService.SetupPlatform(testMode); err != nil {
		return err
	}

	// Simplify error handling by returning directly
	return s.createDirectoryStructure()
}

func (s *ServiceImpl) IsolateEnvironment(name string, _ *types.CommonConfig) error {
	s.logger.Info("Isolating environment", "name", name)

	// Get environment path
	envPath, err := s.getEnvironmentPath(name)
	if err != nil {
		return fmt.Errorf("failed to get environment path: %w", err)
	}

	// Create isolated directory structure
	isolatedPath := filepath.Join(envPath, "isolated")
	if err := os.MkdirAll(isolatedPath, 0755); err != nil {
		return fmt.Errorf("failed to create isolated directory: %w", err)
	}

	username := os.Getenv("USER")
	homeDir := os.Getenv("HOME")
	nixVersion := "nixos-unstable" // Default to unstable channel
	homeManagerVersion := "master" // Default to latest home-manager

	// Create environment-specific Nix configuration
	files := map[string]string{
		"flake.nix": fmt.Sprintf(`{
  description = "Isolated Nix environment for %s";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/%s";
    home-manager = {
      url = "github:nix-community/home-manager/%s";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, home-manager, ... }: {
    defaultPackage.x86_64-linux = self.packages.x86_64-linux.default;
    defaultPackage.x86_64-darwin = self.packages.x86_64-darwin.default;

    packages.x86_64-linux.default = nixpkgs.legacyPackages.x86_64-linux.buildEnv {
      name = "%s-isolated";
      paths = [];
    };

    packages.x86_64-darwin.default = nixpkgs.legacyPackages.x86_64-darwin.buildEnv {
      name = "%s-isolated";
      paths = [];
    };
  };
}`, name, nixVersion, homeManagerVersion, name, name),

		"home.nix": fmt.Sprintf(`{ config, pkgs, ... }:

{
  home.username = "%s";
  home.homeDirectory = "%s";
  home.stateVersion = "23.11";  # Use a constant version

  programs.home-manager.enable = true;

  # Isolated environment settings
  home.packages = with pkgs; [
    # Base isolation packages
    git
    nix
    home-manager
  ];

  # Environment-specific isolation
  nix.settings = {
    sandbox = true;
    restrict-eval = true;
    allowed-users = [ "%s" ];
  };
}`, username, homeDir, username),
	}

	// Write configuration files
	for name, content := range files {
		filePath := filepath.Join(isolatedPath, name)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", name, err)
		}
	}

	// Initialize git repository for version control
	cmd := exec.Command("git", "init")
	cmd.Dir = isolatedPath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	s.logger.Info("Environment isolated successfully", "name", name, "path", isolatedPath)
	return nil
}

// copyEnvironment copies an environment directory with all its contents
func (s *ServiceImpl) copyEnvironment(src, dest string) error {
	s.logger.Debug("Copying environment", "from", src, "to", dest)

	// Create destination directory
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Walk through the source directory
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to access path %s: %w", path, err)
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}
		destPath := filepath.Join(dest, relPath)

		// Handle directories and files differently
		switch {
		case info.IsDir():
			// Create directory
			return os.MkdirAll(destPath, info.Mode())
		case info.Mode()&os.ModeSymlink != 0:
			// Handle symlinks
			target, err := os.Readlink(path)
			if err != nil {
				return fmt.Errorf("failed to read symlink %s: %w", path, err)
			}
			return os.Symlink(target, destPath)
		default:
			// Copy regular files
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", path, err)
			}
			if err := os.WriteFile(destPath, data, info.Mode()); err != nil {
				return fmt.Errorf("failed to write file %s: %w", destPath, err)
			}
		}
		return nil
	})
}

// Helper method for checking rollback conflicts
func (s *ServiceImpl) checkRollbackConflicts(backupPath string) error {
	s.logger.Debug("Checking for rollback conflicts", "backup", backupPath)

	// Check for uncommitted changes
	currentEnv := filepath.Join(s.configDir, "environments", "current")
	if _, err := os.Stat(filepath.Join(currentEnv, ".git")); err == nil {
		// If git repository exists, check for uncommitted changes
		cmd := exec.Command("git", "-C", currentEnv, "status", "--porcelain")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to check git status: %w", err)
		}
		if len(output) > 0 {
			return fmt.Errorf("uncommitted changes detected in current environment")
		}
	}

	// Verify backup exists and is readable
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("backup not accessible: %w", err)
	}

	return nil
}

// Helper functions for error handling and validation

func (s *ServiceImpl) checkSwitchConflicts(currentEnv string) error {
	// Check for uncommitted changes in git repository
	if _, err := os.Stat(filepath.Join(currentEnv, ".git")); err == nil {
		cmd := exec.Command("git", "-C", currentEnv, "status", "--porcelain")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to check git status: %w", err)
		}
		if len(output) > 0 {
			return errors.NewConflictError(
				fmt.Errorf("uncommitted changes"),
				"current environment has uncommitted changes",
			)
		}
	}
	return nil
}

func (s *ServiceImpl) createBackup(src, dest string) error {
	if err := os.RemoveAll(dest); err != nil {
		return fmt.Errorf("failed to clean existing backup: %w", err)
	}
	if err := s.copyEnvironment(src, dest); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}
	return nil
}

func (s *ServiceImpl) restoreFromBackup(src, dest string) error {
	if err := os.RemoveAll(dest); err != nil {
		return fmt.Errorf("failed to clean destination: %w", err)
	}
	if err := s.copyEnvironment(src, dest); err != nil {
		return fmt.Errorf("failed to restore from backup: %w", err)
	}
	return nil
}

func (s *ServiceImpl) performSwitch(targetPath string) error {
	currentLink := filepath.Join(s.configDir, "environments", "current")
	tempLink := currentLink + ".tmp"

	// Create temporary symlink
	if err := os.Symlink(targetPath, tempLink); err != nil {
		return fmt.Errorf("failed to create temporary symlink: %w", err)
	}

	// Atomically replace current symlink
	if err := os.Rename(tempLink, currentLink); err != nil {
		os.Remove(tempLink) // Clean up on failure
		return fmt.Errorf("failed to switch symlink: %w", err)
	}

	return nil
}
