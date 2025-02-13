package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
		return initProject()
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

type ProjectConfig struct {
	Name     string   `yaml:"name"`
	Version  string   `yaml:"version"`
	Required []string `yaml:"required,omitempty"`
	Tools    struct {
		Go     []string `yaml:"go,omitempty"`
		Node   []string `yaml:"node,omitempty"`
		Python []string `yaml:"python,omitempty"`
	} `yaml:"tools,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
}

func initProject() error {
	// Create backup before initialization
	if err := createBackup(); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Initialize project environment
	if err := initProjectEnv(); err != nil {
		return fmt.Errorf("failed to initialize project environment: %w", err)
	}

	// Switch to project environment
	if err := switchEnvironment("project"); err != nil {
		return fmt.Errorf("failed to switch to project environment: %w", err)
	}

	fmt.Println("✅ Project environment initialized successfully")
	return nil
}

func initProjectEnv() error {
	if _, err := os.Stat(".nix-foundry.yaml"); err == nil && !forceProject {
		return fmt.Errorf("project configuration already exists. Use --force to override")
	}

	config := ProjectConfig{
		Name:    filepath.Base(getCurrentDir()),
		Version: "1.0",
		Required: []string{
			"git",
		},
	}

	if teamName != "" {
		teamConfig, err := loadTeamConfig(teamName)
		if err != nil {
			return fmt.Errorf("failed to load team configuration: %w", err)
		}
		config = mergeWithTeamConfig(config, teamConfig)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	if err := os.WriteFile(".nix-foundry.yaml", data, 0644); err != nil {
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	fmt.Println("Project environment initialized successfully")
	return nil
}

func updateProjectConfig() error {
	projectConfig, err := loadProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to load project configuration: %w", err)
	}

	if teamName != "" {
		teamConfig, err := loadTeamConfig(teamName)
		if err != nil {
			return fmt.Errorf("failed to load team configuration: %w", err)
		}

		projectConfig = mergeWithTeamConfig(projectConfig, teamConfig)

		data, err := yaml.Marshal(projectConfig)
		if err != nil {
			return fmt.Errorf("failed to marshal configuration: %w", err)
		}

		if err := os.WriteFile(".nix-foundry.yaml", data, 0644); err != nil {
			return fmt.Errorf("failed to write configuration: %w", err)
		}
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

func loadProjectConfig() (ProjectConfig, error) {
	var config ProjectConfig
	data, err := os.ReadFile(".nix-foundry.yaml")
	if err != nil {
		return config, fmt.Errorf("failed to read project configuration: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse project configuration: %w", err)
	}

	return config, nil
}

func loadTeamConfig(team string) (ProjectConfig, error) {
	var config ProjectConfig
	configPath := filepath.Join(getConfigDir(), "teams", team+".yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, fmt.Errorf("failed to read team configuration: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse team configuration: %w", err)
	}

	return config, nil
}

func mergeWithTeamConfig(project, team ProjectConfig) ProjectConfig {
	if team.Version != "" {
		project.Version = team.Version
	}

	project.Required = mergeLists(project.Required, team.Required)

	project.Tools.Go = mergeLists(project.Tools.Go, team.Tools.Go)
	project.Tools.Node = mergeLists(project.Tools.Node, team.Tools.Node)
	project.Tools.Python = mergeLists(project.Tools.Python, team.Tools.Python)

	if project.Environment == nil {
		project.Environment = make(map[string]string)
	}
	for k, v := range team.Environment {
		project.Environment[k] = v
	}

	return project
}

func mergeLists(a, b []string) []string {
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

func getConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "nix-foundry")
}
