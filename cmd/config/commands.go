/*
Package config provides configuration management commands for Nix Foundry.
It implements Cobra commands for initializing, applying, listing, and showing
configuration details. The package supports different configuration types
(user, team, project) and handles inheritance between configurations.
*/
package config

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/pkg/config"
	"github.com/shawnkhoffman/nix-foundry/pkg/schema"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	configType string
	configName string
	showType   string
)

/*
ApplyCmd represents the apply command for executing the current configuration.
It loads the active configuration (including any inherited configurations) and
applies all settings, including shell configuration, package installation,
and script execution.
*/
var ApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply the current configuration",
	Long: `Apply the current configuration.
This command will load and apply the active configuration, including any inherited configurations.`,
	RunE: runApply,
}

/*
InitCmd represents the init command for creating new configurations.
It supports creating different types of configurations:
- User configuration: Personal settings and preferences
- Team configuration: Shared settings for a team
- Project configuration: Project-specific requirements
Each type has different requirements and inheritance capabilities.
*/
var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new configuration",
	Long: `Initialize a new configuration.
This command will create a new configuration file of the specified type.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		configSvc := config.GetConfigService()

		if configType == "user" {
			if err := configSvc.InitConfig(); err != nil {
				return fmt.Errorf("failed to initialize user config: %w", err)
			}
			fmt.Println("✨ User configuration initialized successfully!")
			return nil
		}

		if configName == "" {
			return fmt.Errorf("name is required for %s configs", configType)
		}

		if err := configSvc.InitConfigWithType(schema.ConfigType(configType), configName); err != nil {
			return fmt.Errorf("failed to initialize %s config: %w", configType, err)
		}

		caser := cases.Title(language.English)
		fmt.Printf("✨ %s configuration '%s' initialized successfully!\n",
			caser.String(string(configType)), configName)
		return nil
	},
}

/*
ListCmd represents the list command for displaying available configurations.
It shows all configurations across different types (user, team, project),
including their inheritance relationships and basic metadata.
*/
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available configurations",
	Long: `List available configurations.
This command will show all available configurations of each type.`,
	RunE: runList,
}

/*
ShowCmd represents the show command for displaying detailed configuration information.
It can show either the active configuration or a specific configuration by name
and type. The output includes all configuration details including metadata,
settings, packages, and scripts.
*/
var ShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show configuration details",
	Long: `Show configuration details.
This command will display the full details of a configuration.
If no name is provided, it will show the active configuration.`,
	RunE: runShow,
}

/*
runApply executes the configuration application process.
It retrieves and applies the active configuration, which includes:
1. Shell environment configuration
2. Package installation
3. Script execution
Returns an error if any part of the application process fails.
*/
func runApply(_ *cobra.Command, _ []string) error {
	configSvc := config.GetConfigService()

	if err := configSvc.ApplyConfig(); err != nil {
		return fmt.Errorf("failed to apply configuration: %w", err)
	}

	fmt.Println("✨ Configuration applied successfully!")
	return nil
}

/*
runList displays all available configurations in the system.
It shows each configuration's:
- Name and type
- Inheritance relationships (if any)
Returns an error if the configuration listing process fails.
*/
func runList(_ *cobra.Command, _ []string) error {
	configSvc := config.GetConfigService()

	configs, err := configSvc.ListConfigs()
	if err != nil {
		return fmt.Errorf("failed to list configurations: %w", err)
	}

	if len(configs) == 0 {
		fmt.Println("No configurations found")
		return nil
	}

	for _, config := range configs {
		fmt.Printf("• %s (%s)\n", config.Metadata.Name, config.Type)
		if config.Base != "" {
			fmt.Printf("  ↳ Extends: %s\n", config.Base)
		}
	}

	return nil
}

/*
runShow displays detailed information about a configuration.
If no specific configuration is requested, it shows the active configuration.
Otherwise, it shows the requested configuration by name and type.
Returns an error if the configuration cannot be found or displayed.
*/
func runShow(_ *cobra.Command, args []string) error {
	configSvc := config.GetConfigService()

	if len(args) == 0 {
		config, err := configSvc.GetActiveConfig()
		if err != nil {
			return fmt.Errorf("failed to get active config: %w", err)
		}

		return showConfig(config)
	}

	name := args[0]
	config, err := configSvc.GetConfig(schema.ConfigType(showType), name)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	return showConfig(config)
}

/*
showConfig formats and displays the full details of a configuration.
It shows:
1. Basic metadata (name, type, inheritance)
2. Settings (shell, logging, updates)
3. Nix configuration (package manager, packages)
4. Scripts (if any)
Returns an error if the display process fails.
*/
func showConfig(config *schema.Config) error {
	fmt.Printf("Configuration: %s (%s)\n", config.Metadata.Name, config.Type)
	if config.Base != "" {
		fmt.Printf("Base: %s\n", config.Base)
	}
	fmt.Printf("Description: %s\n", config.Metadata.Description)
	fmt.Printf("Created: %s\n", config.Metadata.Created.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", config.Metadata.Updated.Format("2006-01-02 15:04:05"))

	fmt.Println("\nSettings:")
	fmt.Printf("  Shell: %s\n", config.Settings.Shell)
	fmt.Printf("  Log Level: %s\n", config.Settings.LogLevel)
	fmt.Printf("  Auto Update: %v\n", config.Settings.AutoUpdate)
	fmt.Printf("  Update Interval: %s\n", config.Settings.UpdateInterval)

	fmt.Println("\nNix:")
	fmt.Printf("  Manager: %s\n", config.Nix.Manager)

	if len(config.Nix.Packages.Core) > 0 {
		fmt.Println("\n  Core Packages:")
		for _, pkg := range config.Nix.Packages.Core {
			fmt.Printf("    • %s\n", pkg)
		}
	}

	if len(config.Nix.Packages.Optional) > 0 {
		fmt.Println("\n  Optional Packages:")
		for _, pkg := range config.Nix.Packages.Optional {
			fmt.Printf("    • %s\n", pkg)
		}
	}

	if len(config.Nix.Scripts) > 0 {
		fmt.Println("\n  Scripts:")
		for _, script := range config.Nix.Scripts {
			fmt.Printf("    • %s: %s\n", script.Name, script.Description)
		}
	}

	return nil
}

/*
init initializes the configuration commands by setting up flags and options.
It configures:
- Init command flags for type and name
- Show command flags for type specification
This function is automatically called during package initialization.
*/
func init() {
	InitCmd.Flags().StringVarP(&configType, "type", "t", "user", "Configuration type (user|team|project)")
	InitCmd.Flags().StringVarP(&configName, "name", "n", "", "Configuration name (required for team and project configs)")
	ShowCmd.Flags().StringVarP(&showType, "type", "t", "", "Configuration type (user|team|project)")
}
