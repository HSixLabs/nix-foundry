package main

import (
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	validShells  = []string{"zsh", "bash", "fish"}
	validEditors = []string{"nano", "vim", "nvim", "emacs", "neovim", "vscode"}
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage nix-foundry configuration",
	Long:  `Manage personal and team configurations for nix-foundry.`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initEnvironment(); err != nil {
			return fmt.Errorf("failed to initialize environment: %w", err)
		}
		return initConfig()
	},
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Long:  `Validate the current configuration for syntax and semantic correctness.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := checkPrerequisites(); err != nil {
			return fmt.Errorf("prerequisites check failed: %w", err)
		}

		if err := setupEnvironmentIsolation(); err != nil {
			return fmt.Errorf("environment isolation setup failed: %w", err)
		}

		configDir := getConfigDir()
		configPath := filepath.Join(configDir, "config.yaml")

		// Check if config exists
		if _, err := os.Stat(configPath); err != nil {
			return fmt.Errorf("configuration not found at %s", configPath)
		}

		// Read config
		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read configuration: %w", err)
		}

		// Validate structure
		var config PersonalConfig
		if err := yaml.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("invalid configuration format: %w", err)
		}

		// Validate version
		if config.Version == "" {
			return fmt.Errorf("missing version in configuration")
		}

		// Validate shell
		if config.Shell.Type != "" && !contains(validShells, config.Shell.Type) {
			return fmt.Errorf("invalid shell type: %s", config.Shell.Type)
		}

		// Validate editor
		if config.Editor.Type != "" && !contains(validEditors, config.Editor.Type) {
			return fmt.Errorf("invalid editor type: %s", config.Editor.Type)
		}

		// Validate development tools if specified
		if config.Development.Languages.Go.Version != "" {
			if !isValidVersion(config.Development.Languages.Go.Version) {
				return fmt.Errorf("invalid Go version: %s", config.Development.Languages.Go.Version)
			}
		}
		if config.Development.Languages.Node.Version != "" {
			if !isValidVersion(config.Development.Languages.Node.Version) {
				return fmt.Errorf("invalid Node version: %s", config.Development.Languages.Node.Version)
			}
		}

		// Validate environment variables
		for envName, env := range config.Environments {
			for key, value := range env {
				if key == "PATH" {
					if !isValidPathFormat(value) {
						return fmt.Errorf("invalid PATH format in environment %s", envName)
					}
				}
			}
		}

		fmt.Println("Configuration validation successful")
		return nil
	},
}

var configCheckConflictsCmd = &cobra.Command{
	Use:   "check-conflicts",
	Short: "Check for configuration conflicts",
	Long:  `Check for conflicts between personal and team configurations.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return checkConfigConflicts()
	},
}

var configApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply configuration changes",
	Long:  `Apply changes from configuration files to the current environment.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return applyConfig()
	},
}

var projectImportCmd = &cobra.Command{
	Use:   "project import [path]",
	Short: "Import project configuration",
	Long: `Import project configuration from a file or directory.
Examples:
  nix-foundry project import ./project-config.yaml   # Import from file
  nix-foundry project import ../other-project       # Import from directory`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return importProjectConfig(args[0])
	},
}

var backupCreateCmd = &cobra.Command{
	Use:   "backup create [name]",
	Short: "Create a named backup",
	Long: `Create a named backup of the current environment state.
Examples:
  nix-foundry backup create                  # Create timestamped backup
  nix-foundry backup create pre-update       # Create named backup
  nix-foundry backup create project-switch   # Create named backup before switching`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		if len(args) > 0 {
			name = args[0]
		}
		return createNamedBackup(name)
	},
}

var configRestoreCmd = &cobra.Command{
	Use:   "restore [backup-path]",
	Short: "Restore configuration from backup",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return restoreBackup(args[0])
	},
}

func init() {
	// Add subcommands to configCmd
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configCheckConflictsCmd)
	configCmd.AddCommand(configApplyCmd)
	configCmd.AddCommand(projectImportCmd)
	configCmd.AddCommand(backupCreateCmd)
	configCmd.AddCommand(configRestoreCmd)

	// Set up flags
	configInitCmd.Flags().BoolVar(&forceConfig, "force", false, "Force initialization even if config exists")
}

type PersonalConfig struct {
	Version string `yaml:"version"`
	Shell   struct {
		Type    string            `yaml:"type"`
		Plugins []string          `yaml:"plugins,omitempty"`
		Aliases map[string]string `yaml:"aliases,omitempty"`
	} `yaml:"shell"`
	Editor struct {
		Type    string   `yaml:"type"`
		Plugins []string `yaml:"plugins,omitempty"`
	} `yaml:"editor"`
	Git struct {
		Enable bool `yaml:"enable"`
	} `yaml:"git"`
	Packages struct {
		Additional       []string            `yaml:"additional"`
		PlatformSpecific map[string][]string `yaml:"platformSpecific"`
		Development      []string            `yaml:"development"`
		Team             map[string]string   `yaml:"team"`
	} `yaml:"packages"`
	Team struct {
		Enable   bool              `yaml:"enable"`
		Name     string            `yaml:"name"`
		Settings map[string]string `yaml:"settings"`
	} `yaml:"team"`
	Platform struct {
		OS   string `yaml:"os"`
		Arch string `yaml:"arch"`
	} `yaml:"platform"`
	Development struct {
		Languages struct {
			Go struct {
				Version string   `yaml:"version"`
				Modules []string `yaml:"modules,omitempty"`
			} `yaml:"go"`
			Node struct {
				Version  string   `yaml:"version"`
				Packages []string `yaml:"packages,omitempty"`
			} `yaml:"node"`
		} `yaml:"languages"`
		Tools []string `yaml:"tools,omitempty"`
	} `yaml:"development"`
	Environments map[string]map[string]string `yaml:"environments,omitempty"`
}

func initConfig() error {
	configDir := getConfigDir()
	configPath := filepath.Join(configDir, "config.yaml")

	if _, err := os.Stat(configPath); err == nil && !forceConfig {
		return fmt.Errorf("configuration already exists at %s", configPath)
	}

	config := PersonalConfig{
		Version: "1.0",
		Shell: struct {
			Type    string            `yaml:"type"`
			Plugins []string          `yaml:"plugins,omitempty"`
			Aliases map[string]string `yaml:"aliases,omitempty"`
		}{
			Type:    "zsh",
			Plugins: []string{"zsh-autosuggestions"},
			Aliases: make(map[string]string),
		},
		Editor: struct {
			Type    string   `yaml:"type"`
			Plugins []string `yaml:"plugins,omitempty"`
		}{
			Type:    "neovim",
			Plugins: make([]string, 0),
		},
		Git: struct {
			Enable bool `yaml:"enable"`
		}{
			Enable: true,
		},
		Packages: struct {
			Additional       []string            `yaml:"additional"`
			PlatformSpecific map[string][]string `yaml:"platformSpecific"`
			Development      []string            `yaml:"development"`
			Team             map[string]string   `yaml:"team"`
		}{
			Additional: []string{"git", "ripgrep"},
			PlatformSpecific: map[string][]string{
				"darwin": {"mas"},
				"linux":  {"inotify-tools"},
			},
			Development: []string{"git", "ripgrep", "fd", "jq"},
			Team:        make(map[string]string),
		},
		Team: struct {
			Enable   bool              `yaml:"enable"`
			Name     string            `yaml:"name"`
			Settings map[string]string `yaml:"settings"`
		}{
			Enable:   false,
			Name:     "",
			Settings: make(map[string]string),
		},
		Platform: struct {
			OS   string `yaml:"os"`
			Arch string `yaml:"arch"`
		}{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		},
		Development: struct {
			Languages struct {
				Go struct {
					Version string   `yaml:"version"`
					Modules []string `yaml:"modules,omitempty"`
				} `yaml:"go"`
				Node struct {
					Version  string   `yaml:"version"`
					Packages []string `yaml:"packages,omitempty"`
				} `yaml:"node"`
			} `yaml:"languages"`
			Tools []string `yaml:"tools,omitempty"`
		}{
			Languages: struct {
				Go struct {
					Version string   `yaml:"version"`
					Modules []string `yaml:"modules,omitempty"`
				} `yaml:"go"`
				Node struct {
					Version  string   `yaml:"version"`
					Packages []string `yaml:"packages,omitempty"`
				} `yaml:"node"`
			}{
				Go: struct {
					Version string   `yaml:"version"`
					Modules []string `yaml:"modules,omitempty"`
				}{
					Version: "",
					Modules: make([]string, 0),
				},
				Node: struct {
					Version  string   `yaml:"version"`
					Packages []string `yaml:"packages,omitempty"`
				}{
					Version:  "",
					Packages: make([]string, 0),
				},
			},
			Tools: make([]string, 0),
		},
		Environments: make(map[string]map[string]string),
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	fmt.Println("Personal configuration initialized successfully")
	return nil
}

func checkConfigConflicts() error {
	configDir := getConfigDir()
	personalConfig := filepath.Join(configDir, "config.yaml")
	projectConfig := filepath.Join(configDir, "project.yaml")

	// If project config doesn't exist, no conflicts possible
	if _, err := os.Stat(projectConfig); os.IsNotExist(err) {
		return nil
	}

	// Read both configs
	personal := &PersonalConfig{}
	project := &PersonalConfig{} // Using same struct as they share structure

	personalData, err := os.ReadFile(personalConfig)
	if err != nil {
		return fmt.Errorf("failed to read personal config: %w", err)
	}

	projectData, err := os.ReadFile(projectConfig)
	if err != nil {
		return fmt.Errorf("failed to read project config: %w", err)
	}

	if err := yaml.Unmarshal(personalData, personal); err != nil {
		return fmt.Errorf("invalid personal config: %w", err)
	}

	if err := yaml.Unmarshal(projectData, project); err != nil {
		return fmt.Errorf("invalid project config: %w", err)
	}

	// Check for conflicts
	conflicts := []string{}

	// Check shell conflicts
	if personal.Shell.Type != project.Shell.Type {
		conflicts = append(conflicts, fmt.Sprintf("shell type mismatch: personal=%s, project=%s",
			personal.Shell.Type, project.Shell.Type))
	}

	// Check editor conflicts
	if personal.Editor.Type != project.Editor.Type {
		conflicts = append(conflicts, fmt.Sprintf("editor type mismatch: personal=%s, project=%s",
			personal.Editor.Type, project.Editor.Type))
	}

	// Check environment variable conflicts
	for env, value := range project.Environments {
		if personalValue, exists := personal.Environments[env]; exists {
			if !maps.Equal(value, personalValue) {
				conflicts = append(conflicts, fmt.Sprintf("environment %s has conflicting values", env))
			}
		}
	}

	if len(conflicts) > 0 {
		return fmt.Errorf("configuration conflicts found:\n- %s", strings.Join(conflicts, "\n- "))
	}

	return nil
}

func applyConfig() error {
	configDir := getConfigDir()
	currentEnv := filepath.Join(configDir, "environments", "current")

	// Create backup before applying changes
	if err := createBackup(); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Check for conflicts unless forced
	if !forceConfig {
		if err := checkConfigConflicts(); err != nil {
			return fmt.Errorf("configuration conflicts detected: %w\nUse --force to override", err)
		}
	}

	// Apply configuration using home-manager
	cmd := exec.Command("home-manager", "switch")
	cmd.Dir = currentEnv
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply configuration: %w", err)
	}

	fmt.Println("✅ Configuration applied successfully")
	return nil
}

func initEnvironment() error {
	// Check prerequisites
	if err := checkPrerequisites(); err != nil {
		return fmt.Errorf("prerequisite check failed: %w", err)
	}

	// Create directory structure
	configDir := getConfigDir()
	dirs := []string{
		configDir,
		filepath.Join(configDir, "environments"),
		filepath.Join(configDir, "environments", "default"),
		filepath.Join(configDir, "backups"),
		filepath.Join(configDir, "logs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create symlink to default environment as current
	currentEnv := filepath.Join(configDir, "environments", "current")
	defaultEnv := filepath.Join(configDir, "environments", "default")
	// Remove existing symlink if it exists
	if err := os.Remove(currentEnv); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing current environment link: %w", err)
	}
	if err := os.Symlink(defaultEnv, currentEnv); err != nil {
		return fmt.Errorf("failed to create current environment symlink: %w", err)
	}

	// Initialize default configuration if it doesn't exist
	configPath := filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := initConfig(); err != nil {
			return fmt.Errorf("failed to initialize configuration: %w", err)
		}
	}

	// Setup environment isolation
	if err := setupEnvironmentIsolation(); err != nil {
		return fmt.Errorf("failed to setup environment isolation: %w", err)
	}

	fmt.Println("nix-foundry environment initialized successfully")
	return nil
}

func checkPrerequisites() error {
	// Check for nix installation
	if _, err := exec.LookPath("nix"); err != nil {
		return fmt.Errorf("nix is not installed: %w", err)
	}

	// Check for home-manager
	if _, err := exec.LookPath("home-manager"); err != nil {
		return fmt.Errorf("home-manager is not installed: %w", err)
	}

	return nil
}

func setupEnvironmentIsolation() error {
	// Create default environment profile
	envDir := filepath.Join(getConfigDir(), "environments")
	defaultEnv := filepath.Join(envDir, "default")

	if err := os.MkdirAll(defaultEnv, 0755); err != nil {
		return fmt.Errorf("failed to create default environment: %w", err)
	}

	// Create initial flake.nix if it doesn't exist
	flakePath := filepath.Join(defaultEnv, "flake.nix")
	if _, err := os.Stat(flakePath); os.IsNotExist(err) {
		flakeContent := `{
  description = "Home Manager configuration";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { nixpkgs, home-manager, ... }:
    let
      system = "x86_64-linux";  # Adjust this based on your system
      pkgs = nixpkgs.legacyPackages.${system};
    in {
      homeConfigurations.default = home-manager.lib.homeManagerConfiguration {
        inherit pkgs;
        modules = [ ./home.nix ];
      };
    };
}`
		if err := os.WriteFile(flakePath, []byte(flakeContent), 0644); err != nil {
			return fmt.Errorf("failed to create flake.nix: %w", err)
		}
	}

	// Create initial home.nix if it doesn't exist
	homePath := filepath.Join(defaultEnv, "home.nix")
	if _, err := os.Stat(homePath); os.IsNotExist(err) {
		homeContent := `{ config, pkgs, ... }:

{
  home.username = "${USER}";
  home.homeDirectory = "${HOME}";
  home.stateVersion = "23.11";

  programs.home-manager.enable = true;

  # Basic packages
  home.packages = with pkgs; [
    git
    ripgrep
  ];
}`
		homeContent = strings.ReplaceAll(homeContent, "${USER}", os.Getenv("USER"))
		homeContent = strings.ReplaceAll(homeContent, "${HOME}", os.Getenv("HOME"))

		if err := os.WriteFile(homePath, []byte(homeContent), 0644); err != nil {
			return fmt.Errorf("failed to create home.nix: %w", err)
		}
	}

	return nil
}

func createBackup() error {
	backupDir := filepath.Join(getConfigDir(), "backups")
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	backupPath := filepath.Join(backupDir, fmt.Sprintf("backup-%s.tar.gz", timestamp))

	// Create backup archive
	cmd := exec.Command("tar", "-czf", backupPath, "-C", getConfigDir(), ".")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create backup archive: %w", err)
	}

	return nil
}

func restoreBackup(backupPath string) error {
	configDir := getConfigDir()

	// Extract backup
	cmd := exec.Command("tar", "-xzf", backupPath, "-C", configDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}

func importProjectConfig(path string) error {
	// Create backup before import
	if err := createBackup(); err != nil {
		return fmt.Errorf("failed to create backup before import: %w", err)
	}

	configDir := getConfigDir()
	projectConfig := filepath.Join(configDir, "project.yaml")

	// Check if path is a directory
	fi, statErr := os.Stat(path)
	if statErr != nil {
		return fmt.Errorf("failed to access path %s: %w", path, statErr)
	}

	var configPath string
	if fi.IsDir() {
		// Look for config file in directory
		configPath = filepath.Join(path, "nix-foundry.yaml")
		if _, err := os.Stat(configPath); err != nil {
			return fmt.Errorf("no nix-foundry.yaml found in directory %s", path)
		}
	} else {
		configPath = path
	}

	// Read and validate the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse and validate the config
	config := &PersonalConfig{} // Using same struct as they share structure
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("invalid config file: %w", err)
	}

	// Check for conflicts unless forced
	if !forceConfig {
		if err := checkConfigConflicts(); err != nil {
			return fmt.Errorf("configuration conflicts detected: %w\nUse --force to override", err)
		}
	}

	// Copy the config file
	if err := os.WriteFile(projectConfig, data, 0644); err != nil {
		return fmt.Errorf("failed to write project config: %w", err)
	}

	fmt.Println("✅ Project configuration imported successfully")

	// If not forced, show a warning about applying changes
	if !forceConfig {
		fmt.Println("\nℹ️  Run 'nix-foundry config apply' to apply the changes")
	} else {
		// Apply changes immediately if forced
		if err := applyConfig(); err != nil {
			return fmt.Errorf("failed to apply configuration: %w", err)
		}
	}

	return nil
}

func createNamedBackup(name string) error {
	backupDir := filepath.Join(getConfigDir(), "backups")

	// Ensure backup directory exists
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate backup name
	timestamp := time.Now().Format("20060102-150405")
	backupName := timestamp
	if name != "" {
		// Sanitize the name to be filesystem-friendly
		name = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
				return r
			}
			return '-'
		}, name)
		backupName = fmt.Sprintf("%s-%s", name, timestamp)
	}

	backupPath := filepath.Join(backupDir, fmt.Sprintf("%s.tar.gz", backupName))

	// Create backup archive
	cmd := exec.Command("tar", "-czf", backupPath, "-C", getConfigDir(), ".")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create backup archive: %w", err)
	}

	fmt.Printf("✅ Created backup: %s\n", backupName)
	return nil
}

func isValidPathFormat(path string) bool {
	// PATH should be a colon-separated list of directories
	parts := strings.Split(path, ":")
	for _, part := range parts {
		if part == "" {
			continue // Empty parts are allowed in PATH
		}
		if _, err := os.Stat(part); err != nil {
			// Path component doesn't exist or isn't accessible
			return false
		}
	}
	return true
}

func isValidVersion(version string) bool {
	// Basic semver format: MAJOR.MINOR.PATCH
	parts := strings.Split(version, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return false
	}

	// Validate each part is a number
	for _, part := range parts {
		if _, err := strconv.Atoi(part); err != nil {
			return false
		}
	}

	return true
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
