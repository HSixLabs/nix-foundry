package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/config"
	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/progress"
)

var configManager *config.Manager

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
		var err error
		configManager, err = config.NewConfigManager()
		if err != nil {
			return err
		}

		var nixConfig config.NixConfig
		if err := configManager.ReadConfig("config.yaml", &nixConfig); err != nil {
			return err
		}

		validator := config.NewValidator(&nixConfig)
		if err := validator.ValidateConfig(); err != nil {
			return fmt.Errorf("invalid configuration: %w", err)
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
		var err error
		configManager, err = config.NewConfigManager()
		if err != nil {
			return err
		}

		var nixConfig config.NixConfig
		if err := configManager.ReadConfig("config.yaml", &nixConfig); err != nil {
			return err
		}

		spin := progress.NewSpinner("Applying configuration...")
		spin.Start()
		if err := configManager.Apply(&nixConfig, testMode); err != nil {
			spin.Fail("Configuration application failed")
			return fmt.Errorf("failed to apply configuration: %w", err)
		}
		spin.Success("Configuration applied successfully")

		return nil
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

func init() {
	// Add subcommands to configCmd
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configCheckConflictsCmd)
	configCmd.AddCommand(configApplyCmd)
	configCmd.AddCommand(projectImportCmd)
}

func initConfig() error {
	var err error
	configManager, err = config.NewConfigManager()
	if err != nil {
		return err
	}

	if configManager.ConfigExists("config.yaml") && !forceConfig {
		return fmt.Errorf("configuration already exists at %s", filepath.Join(configManager.GetConfigDir(), "config.yaml"))
	}

	nixConfig := config.NixConfig{
		Version: "1.0",
		Shell: config.ShellConfig{
			Type:    "zsh",
			Plugins: []string{"zsh-autosuggestions"},
		},
		Editor: config.EditorConfig{
			Type:       "neovim",
			Extensions: make([]string, 0),
		},
		Git: config.GitConfig{
			Enable: true,
		},
		Packages: config.PackagesConfig{
			Additional: []string{"git", "ripgrep"},
			PlatformSpecific: map[string][]string{
				"darwin": {"mas"},
				"linux":  {"inotify-tools"},
			},
			Development: []string{"git", "ripgrep", "fd", "jq"},
			Team:        make(map[string][]string),
		},
		Team: config.TeamConfig{
			Enable:   false,
			Name:     "",
			Settings: make(map[string]string),
		},
		Platform: config.PlatformConfig{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		},
		Development: config.DevelopmentConfig{
			Languages: struct {
				Go struct {
					Version  string   `yaml:"version,omitempty"`
					Packages []string `yaml:"packages,omitempty"`
				} `yaml:"go"`
				Node struct {
					Version  string   `yaml:"version"`
					Packages []string `yaml:"packages,omitempty"`
				} `yaml:"node"`
				Python struct {
					Version  string   `yaml:"version"`
					Packages []string `yaml:"packages,omitempty"`
				} `yaml:"python"`
			}{},
			Tools: []string{},
		},
	}

	if autoConfig {
		if err := configManager.WriteConfig("config.yaml", &nixConfig); err != nil {
			return fmt.Errorf("failed to write configuration: %w", err)
		}
		return nil
	}

	validator := config.NewValidator(&nixConfig)
	if err := validator.ValidateConfig(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	fmt.Println("Personal configuration initialized successfully")
	return nil
}

func checkConfigConflicts() error {
	_, err := config.NewConfigManager()
	if err != nil {
		return err
	}

	// If project config doesn't exist, no conflicts possible
	if !configManager.ConfigExists("project.yaml") {
		return nil
	}

	var personal, project config.NixConfig
	if err := configManager.ReadConfig("config.yaml", &personal); err != nil {
		return fmt.Errorf("failed to read personal config: %w", err)
	}

	if err := configManager.ReadConfig("project.yaml", &project); err != nil {
		return fmt.Errorf("failed to read project config: %w", err)
	}

	validator := config.NewValidator(&personal)
	if err := validator.ValidateConflicts(&project); err != nil {
		return fmt.Errorf("configuration conflicts found:\n%w", err)
	}

	return nil
}

func initEnvironment() error {
	// Check prerequisites
	if err := checkPrerequisites(); err != nil {
		return fmt.Errorf("prerequisite check failed: %w", err)
	}

	var err error
	configManager, err = config.NewConfigManager()
	if err != nil {
		return err
	}

	configDir := configManager.GetConfigDir()
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
	if !configManager.ConfigExists("config.yaml") {
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
	// Skip prerequisite checks in test mode
	if testMode {
		return nil
	}

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

func importProjectConfig(path string) error {
	var err error
	configManager, err = config.NewConfigManager()
	if err != nil {
		return err
	}

	// Create backup before import
	err = configManager.CreateBackup()
	if err != nil {
		return fmt.Errorf("failed to create backup before import: %w", err)
	}

	// Check if path is a directory
	fi, statErr := os.Stat(path)
	if statErr != nil {
		return fmt.Errorf("failed to access path %s: %w", path, statErr)
	}

	var configPath string
	if fi.IsDir() {
		configPath = filepath.Join(path, "nix-foundry.yaml")
		if _, statErr := os.Stat(configPath); statErr != nil {
			return fmt.Errorf("no nix-foundry.yaml found in directory %s", path)
		}
	} else {
		configPath = path
	}

	// Load and validate the config
	projectConfig, err := configManager.LoadConfig(config.ProjectConfigType, configPath)
	if err != nil {
		return fmt.Errorf("failed to load config file: %w", err)
	}

	// Check for conflicts unless forced
	if !forceConfig {
		if err := checkConfigConflicts(); err != nil {
			return fmt.Errorf("configuration conflicts detected: %w\nUse --force to override", err)
		}
	}

	if err := configManager.WriteConfig("project.yaml", projectConfig); err != nil {
		return fmt.Errorf("failed to write project config: %w", err)
	}

	fmt.Println("✅ Project configuration imported successfully")

	if !forceConfig {
		fmt.Println("\nℹ️  Run 'nix-foundry config apply' to apply the changes")
	} else {
		if err := configManager.Apply(projectConfig, testMode); err != nil {
			return fmt.Errorf("failed to apply configuration: %w", err)
		}
	}

	return nil
}
