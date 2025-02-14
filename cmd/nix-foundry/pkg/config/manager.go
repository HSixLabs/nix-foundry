package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Manager struct {
	configDir string
	backupDir string
	paths     Paths
}

// ConfigOptions represents options for configuration operations
type Options struct {
	Force    bool
	Validate bool
	Backup   bool
}

func NewConfigManager() (*Manager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "nix-foundry")

	cm := &Manager{
		configDir: configDir,
		backupDir: filepath.Join(configDir, "backups"),
		paths: Paths{
			Personal: filepath.Join(configDir, "config.yaml"),
			Project:  filepath.Join(configDir, "project.yaml"),
			Team:     filepath.Join(configDir, "teams"),
			Current:  filepath.Join(configDir, "environments", "current"),
		},
	}

	return cm, nil
}

// SafeWrite writes configuration with optional backup and validation
func (cm *Manager) SafeWrite(filename string, config interface{}, opts Options) error {
	if opts.Backup {
		if err := cm.CreateBackup(); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}
	}

	if opts.Validate {
		if v, ok := config.(interface{ Validate() error }); ok {
			if err := v.Validate(); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}
		}
	}

	return cm.WriteConfig(filename, config)
}

// ReadConfig reads and unmarshals configuration
func (cm *Manager) ReadConfig(filename string, config interface{}) error {
	configPath := filepath.Join(cm.configDir, filename)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read configuration: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	return nil
}

// WriteConfig marshals and writes configuration
func (cm *Manager) WriteConfig(filename string, config interface{}) error {
	configPath := filepath.Join(cm.configDir, filename)

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	return nil
}

// CreateBackup creates a timestamped backup of the current configuration
func (cm *Manager) CreateBackup() error {
	if err := os.MkdirAll(cm.backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(cm.backupDir, fmt.Sprintf("backup-%s.tar.gz", timestamp))

	cmd := exec.Command("tar", "-czf", backupPath, "-C", cm.configDir, ".")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create backup archive: %w", err)
	}

	return nil
}

func (cm *Manager) ConfigExists(filename string) bool {
	_, err := os.Stat(filepath.Join(cm.configDir, filename))
	return err == nil
}

func (cm *Manager) GetConfigDir() string {
	return cm.configDir
}

func (cm *Manager) GetBackupDir() string {
	return cm.backupDir
}

func (cm *Manager) loadTeamConfig(teamName string) (*ProjectConfig, error) {
	var config ProjectConfig
	teamConfigPath := filepath.Join(cm.configDir, "teams", teamName+".yaml")

	if err := cm.ReadConfig(teamConfigPath, &config); err != nil {
		return nil, fmt.Errorf("failed to load team config: %w", err)
	}

	return &config, nil
}

// Apply applies the configuration after validation and backup
func (cm *Manager) Apply(config interface{}, testMode bool) error {
	// Skip actual home-manager application in test mode
	if testMode {
		// Create a dummy backup in test mode
		if err := cm.CreateBackup(); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}

		// Create dummy configuration files
		files := []string{
			filepath.Join(cm.configDir, "flake.nix"),
			filepath.Join(cm.configDir, "home.nix"),
		}
		for _, file := range files {
			var content string
			var err error

			// Convert config to map[string]string if it's not already
			configMap := make(map[string]string)
			switch c := config.(type) {
			case map[string]string:
				configMap = c
			case map[string]interface{}:
				for k, v := range c {
					if str, ok := v.(string); ok {
						configMap[k] = str
					}
				}
			}

			switch filepath.Base(file) {
			case "home.nix":
				content, err = cm.GenerateTestConfig(config)
			case "flake.nix":
				content = fmt.Sprintf(`{
					description = "Home Manager configuration";
					inputs = {
						nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
						home-manager = {
							url = "github:nix-community/home-manager";
							inputs.nixpkgs.follows = "nixpkgs";
						};
					};
					outputs = { nixpkgs, home-manager, ... }: {
						defaultPackage.x86_64-linux = home-manager.defaultPackage.x86_64-linux;
						homeConfigurations = {
							test = {
								team = "%s";
								package = pkgs.%s;
							};
						};
					};
				}`, configMap["team"], configMap["shell"])
				err = nil
			default:
				err = fmt.Errorf("unknown file type: %s", file)
			}

			if err != nil {
				return fmt.Errorf("failed to generate config for %s: %w", file, err)
			}

			if err := os.WriteFile(file, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to create test file %s: %w", file, err)
			}
		}
		return nil
	}

	// Check if home-manager is installed
	if _, err := exec.LookPath("home-manager"); err != nil {
		return fmt.Errorf("home-manager not found: please install it first using 'nix-env -iA nixpkgs.home-manager'")
	}

	// First validate the configuration
	if v, ok := config.(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	// Create backup before applying
	if err := cm.CreateBackup(); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	// Generate Nix configuration
	if err := cm.generateNixConfig(config); err != nil {
		return fmt.Errorf("failed to generate nix config: %w", err)
	}

	// Apply using home-manager
	cmd := exec.Command("home-manager", "switch")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply configuration: %w", err)
	}

	return nil
}

// MergeProjectConfigs merges two project configurations
func (cm *Manager) MergeProjectConfigs(base, overlay ProjectConfig) ProjectConfig {
	result := overlay

	// Merge required packages
	seen := make(map[string]bool)
	var required []string

	for _, pkg := range base.Required {
		if !seen[pkg] {
			seen[pkg] = true
			required = append(required, pkg)
		}
	}

	for _, pkg := range overlay.Required {
		if !seen[pkg] {
			seen[pkg] = true
			required = append(required, pkg)
		}
	}

	result.Required = required

	// Merge tools
	result.Tools.Go = cm.mergeLists(base.Tools.Go, overlay.Tools.Go)
	result.Tools.Node = cm.mergeLists(base.Tools.Node, overlay.Tools.Node)
	result.Tools.Python = cm.mergeLists(base.Tools.Python, overlay.Tools.Python)

	// Merge environment variables
	if result.Environment == nil {
		result.Environment = make(map[string]string)
	}
	for k, v := range base.Environment {
		if _, exists := result.Environment[k]; !exists {
			result.Environment[k] = v
		}
	}

	return result
}

func (cm *Manager) mergeLists(a, b []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range a {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	for _, item := range b {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

func (cm *Manager) generateNixConfig(config interface{}) error {
	// Convert config to NixConfig type
	nixConfig, ok := config.(*NixConfig)
	if !ok {
		return fmt.Errorf("invalid configuration type: expected *NixConfig")
	}

	// Generate home-manager configuration
	configPath := filepath.Join(cm.configDir, "home-manager", "home.nix")

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create home-manager directory: %w", err)
	}

	// Generate Nix expression
	nixExpr := generateHomeManagerConfig(nixConfig)

	// Write configuration
	if err := os.WriteFile(configPath, []byte(nixExpr), 0644); err != nil {
		return fmt.Errorf("failed to write home-manager configuration: %w", err)
	}

	return nil
}

func generateHomeManagerConfig(config *NixConfig) string {
	// Basic home-manager configuration template
	return fmt.Sprintf(`
{ config, pkgs, ... }:

{
  home.username = builtins.getEnv "USER";
  home.homeDirectory = builtins.getEnv "HOME";
  home.stateVersion = "23.11";

  programs.home-manager.enable = true;

  # Shell configuration
  programs.%s.enable = true;

  # Editor configuration
  programs.%s.enable = true;

  # Git configuration
  programs.git = {
    enable = %v;
    userName = "%s";
    userEmail = "%s";
  };

  # Package management
  home.packages = with pkgs; [
    %s
  ];
}`,
		config.Shell.Type,
		config.Editor.Type,
		config.Git.Enable,
		config.Git.User.Name,
		config.Git.User.Email,
		strings.Join(config.Packages.Additional, "\n    "),
	)
}

// LoadConfig loads any configuration type with proper validation
func (cm *Manager) LoadConfig(configType Type, name string) (interface{}, error) {
	var config interface{}
	var path string

	switch configType {
	case PersonalConfigType:
		config = &NixConfig{}
		path = cm.paths.Personal
	case ProjectConfigType:
		config = &ProjectConfig{
			BaseConfig: BaseConfig{
				Type: ProjectConfigType,
			},
		}
		path = cm.paths.Project
		if name != "" {
			path = filepath.Join(cm.configDir, "projects", name+".yaml")
		}
	case TeamConfigType:
		config = &ProjectConfig{
			BaseConfig: BaseConfig{
				Type: TeamConfigType,
			},
		}
		path = filepath.Join(cm.paths.Team, name+".yaml")
	default:
		return nil, fmt.Errorf("unknown config type: %s", configType)
	}

	if err := cm.ReadConfig(path, config); err != nil {
		return nil, err
	}

	return config, nil
}

// GenerateTestConfig generates a test configuration for testing purposes
func (cm *Manager) GenerateTestConfig(config interface{}) (string, error) {
	// Convert config to map[string]string if it's not already
	configMap := make(map[string]string)
	switch c := config.(type) {
	case map[string]string:
		configMap = c
	case map[string]interface{}:
		for k, v := range c {
			if str, ok := v.(string); ok {
				configMap[k] = str
			}
		}
	case nil:
		// Use defaults for nil config
		configMap = map[string]string{
			"shell":  "zsh",
			"editor": "nano",
		}
	}

	// Default shell if not specified
	shell := "zsh"
	if s, ok := configMap["shell"]; ok && s != "" {
		shell = s
	}

	// Default editor if not specified
	editor := "nano"
	if e, ok := configMap["editor"]; ok && e != "" {
		editor = e
	}

	// Default git configuration
	gitName := configMap["git-name"]
	gitEmail := configMap["git-email"]
	gitEnabled := gitName != "" && gitEmail != ""

	// Ensure the config directory exists
	if err := os.MkdirAll(cm.configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return fmt.Sprintf(`{ config, pkgs, ... }:

{
  # Let Home Manager manage itself
  programs.home-manager.enable = true;

  home = {
    username = builtins.getEnv "USER";
    homeDirectory = builtins.getEnv "HOME";
    stateVersion = "23.11";

    # Package management
    packages = with pkgs; [
      %s    # Shell package
    ];
  };

  # Shell configuration
  programs.%s = {
    enable = true;
    package = pkgs.%s;
  };

  # Editor configuration
  programs.%s = {
    enable = true;
  };

  %s  # Git configuration (conditionally included)
}`,
		shell,
		shell, shell,
		editor,
		generateGitConfig(gitEnabled, gitName, gitEmail)), nil
}

// Helper function to generate git configuration
func generateGitConfig(enabled bool, name, email string) string {
	if !enabled {
		return "# Git configuration disabled"
	}
	return fmt.Sprintf(`# Git configuration
  programs.git = {
    enable = true;
    userName = "%s";
    userEmail = "%s";
  };`, name, email)
}
