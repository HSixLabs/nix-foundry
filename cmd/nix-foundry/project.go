package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/config"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage project environments",
	Long:  `Manage project-specific development environments and team configurations.`,
}

var projectInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize project environment",
	Long: `Initialize a new project environment with optional team configuration.

Examples:
  # Basic project init
  nix-foundry project init

  # Init with team config
  nix-foundry project init --team myteam`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		if configManager == nil {
			configManager, err = config.NewConfigManager()
			if err != nil {
				return fmt.Errorf("failed to initialize config manager: %w", err)
			}
		}
		return initProject(configManager)
	},
}

var projectUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update project configuration",
	Long:  `Update project configuration with latest team settings and check for conflicts.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return updateProjectConfig()
	},
}

func init() {
	projectCmd.AddCommand(projectInitCmd)
	projectCmd.AddCommand(projectUpdateCmd)

	projectInitCmd.Flags().StringVar(&teamName, "team", "", "Team configuration to use")
	projectInitCmd.Flags().BoolVar(&forceProject, "force", false, "Force initialization even if project exists")
}

func initProject(cm *config.Manager) error {
	if err := cm.CreateBackup(); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	if err := initProjectEnv(cm); err != nil {
		return fmt.Errorf("failed to initialize project environment: %w", err)
	}

	fmt.Println("✅ Project environment initialized successfully")
	return nil
}

func initProjectEnv(configManager *config.Manager) error {
	if _, err := os.Stat(".nix-foundry.yaml"); err == nil && !forceProject {
		return fmt.Errorf("project configuration already exists. Use --force to override")
	}

	projectCfg := config.ProjectConfig{
		BaseConfig: config.BaseConfig{
			Type:    config.ProjectConfigType,
			Version: "1.0",
			Name:    filepath.Base(getCurrentDir()),
		},
		Required: []string{
			"git",
		},
	}

	if teamName != "" {
		teamConfig, err := configManager.LoadConfig(config.TeamConfigType, teamName)
		if err != nil {
			return fmt.Errorf("failed to load team configuration: %w", err)
		}
		teamProjectConfig, ok := teamConfig.(*config.ProjectConfig)
		if !ok {
			return fmt.Errorf("invalid team configuration type")
		}
		projectCfg = configManager.MergeProjectConfigs(projectCfg, *teamProjectConfig)
	}

	if err := configManager.WriteConfig(".nix-foundry.yaml", projectCfg); err != nil {
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	return nil
}

func updateProjectConfig() error {
	if configManager == nil {
		var initErr error
		configManager, initErr = config.NewConfigManager()
		if initErr != nil {
			return fmt.Errorf("failed to initialize config manager: %w", initErr)
		}
	}

	projectConfig, err := configManager.LoadProjectWithTeam("", teamName)
	if err != nil {
		return fmt.Errorf("failed to load project configuration: %w", err)
	}

	if err := configManager.WriteConfig(".nix-foundry.yaml", projectConfig); err != nil {
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	if err := checkConfigConflicts(); err != nil {
		fmt.Println("Warning: Configuration conflicts detected")
		fmt.Println(err)
	}

	fmt.Println("Project configuration updated successfully")
	return nil
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "project"
	}
	return dir
}

func getConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "nix-foundry")
}
