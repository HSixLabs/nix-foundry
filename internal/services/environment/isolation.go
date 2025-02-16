package environment

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// SetupIsolation initializes an isolated environment structure
func (s *ServiceImpl) SetupIsolation(testMode bool) error {
	if testMode {
		s.logger.Debug("Skipping isolation setup in test mode")
		return nil
	}

	// Create required directory structure
	if err := s.createDirectoryStructure(); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Initialize default environment
	if err := s.setupEnvironment("default"); err != nil {
		return fmt.Errorf("failed to setup default environment: %w", err)
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

// initializeNixFlake sets up the Nix flake environment
func (s *ServiceImpl) initializeNixFlake() error {
	s.logger.Info("Initializing Nix flake environment")

	defaultEnvPath := filepath.Join(s.configDir, "environments", "default")

	// Create flake.nix with proper configuration
	flakeContent := `{
  description = "Isolated development environment";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, home-manager, ... }: {
    defaultPackage.x86_64-linux = self.packages.x86_64-linux.default;

    packages.x86_64-linux.default = nixpkgs.legacyPackages.x86_64-linux.buildEnv {
      name = "isolated-env";
      paths = [];
    };
  };
}`

	if err := os.WriteFile(filepath.Join(defaultEnvPath, "flake.nix"), []byte(flakeContent), 0644); err != nil {
		return fmt.Errorf("failed to initialize flake: %w", err)
	}

	// Initialize git repository for flake
	cmd := exec.Command("git", "init")
	cmd.Dir = defaultEnvPath
	if output, err := cmd.CombinedOutput(); err != nil {
		s.logger.Error("Git initialization failed", "output", string(output))
		return fmt.Errorf("git initialization failed: %w", err)
	}

	// Initialize flake
	cmd = exec.Command("nix", "flake", "init")
	cmd.Dir = defaultEnvPath
	if output, err := cmd.CombinedOutput(); err != nil {
		s.logger.Error("Flake initialization failed", "output", string(output))
		return fmt.Errorf("nix flake initialization failed: %w", err)
	}

	return nil
}

// setupEnvironmentSymlink creates the current environment symlink
func (s *ServiceImpl) setupEnvironmentSymlink() error {
	defaultEnv := filepath.Join(s.configDir, "environments", "default")
	currentEnv := filepath.Join(s.configDir, "environments", "current")

	// Remove existing symlink if it exists
	if _, err := os.Lstat(currentEnv); err == nil {
		if err := os.Remove(currentEnv); err != nil {
			return fmt.Errorf("failed to remove existing symlink: %w", err)
		}
	}

	if err := os.Symlink(defaultEnv, currentEnv); err != nil {
		return fmt.Errorf("failed to create environment symlink: %w", err)
	}

	return nil
}

// setupEnvironment initializes a new environment
func (s *ServiceImpl) setupEnvironment(env string) error {
	envPath := filepath.Join(s.configDir, "environments", env)

	if err := os.MkdirAll(envPath, 0755); err != nil {
		return fmt.Errorf("failed to create environment directory: %w", err)
	}

	// Initialize environment files
	return s.initializeEnvironmentFiles(envPath)
}

// getEnvironmentPath returns the full path to an environment
func (s *ServiceImpl) getEnvironmentPath(env string) (string, error) {
	path := filepath.Join(s.configDir, "environments", env)
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("environment not found: %w", err)
	}
	return path, nil
}

// initializeEnvironmentFiles creates the initial environment configuration files
func (s *ServiceImpl) initializeEnvironmentFiles(envPath string) error {
	files := map[string]string{
		"flake.nix": `{
  description = "Nix environment configuration";
  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    home-manager.url = "github:nix-community/home-manager";
  };
  outputs = { nixpkgs, home-manager, ... }: {
    # Environment configuration goes here
  };
}`,
		"home.nix": `{ config, pkgs, ... }: {
  home.packages = with pkgs; [
    # Add your packages here
  ];
}`,
	}

	for name, content := range files {
		filePath := filepath.Join(envPath, name)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", name, err)
		}
	}

	return nil
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

func loadExistingEnvironments(configDir string) map[string]string {
	envs := make(map[string]string)
	environmentsDir := filepath.Join(configDir, "environments")

	entries, err := os.ReadDir(environmentsDir)
	if err != nil {
		return envs
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "current" {
			envs[entry.Name()] = filepath.Join(environmentsDir, entry.Name())
		}
	}
	return envs
}
