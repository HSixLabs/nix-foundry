package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/shawnkhoffman/nix-foundry/service/config"

	"github.com/spf13/cobra"
)

func NewSetupCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup the initial configuration",
		Long: `Sets up the initial configuration for Nix Foundry.

This creates a default user configuration that serves as the active configuration.
If an active configuration already exists, use the --force flag to overwrite it.

You can customize the configuration by selecting preferred packages and tools.
These selections can be modified at any time.

Examples:
  nix-foundry config setup
  nix-foundry config setup --force`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := config.NewConfigService()

			// Define default configuration parameters
			defaultKind := "user"
			defaultName := "default"
			basePath := "" // No base for initial setup

			// Initialize the default configuration
			if err := svc.InitConfig(defaultKind, defaultName, force, false, basePath); err != nil {
				return fmt.Errorf("failed to setup initial config: %w", err)
			}

			// Define paths
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			activeConfigPath := filepath.Join(homeDir, ".config", "nix-foundry", "config.yaml")
			createdConfigPath := filepath.Join(homeDir, ".config", "nix-foundry", "users", defaultName, "config.yaml")

			// Set the active configuration by copying the created config
			if copyErr := svc.CopyConfig(createdConfigPath, activeConfigPath); copyErr != nil {
				return fmt.Errorf("failed to set active config: %w", copyErr)
			}

			// Prompt user for additional package/tool selections
			selectedPackages, err := promptForPackages()
			if err != nil {
				return fmt.Errorf("failed to get package selections: %w", err)
			}

			// Update the active configuration with selected packages
			if err := svc.UpdateActiveConfigWithPackages(activeConfigPath, selectedPackages); err != nil {
				return fmt.Errorf("failed to update active config with packages: %w", err)
			}

			fmt.Println("Successfully set up the initial configuration at ~/.config/nix-foundry/config.yaml")
			fmt.Println("You can customize your configuration further by modifying the config.yaml file or using the CLI commands.")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing active configuration")
	return cmd
}

// promptForPackages interactively prompts the user to select common packages and tools.
func promptForPackages() ([]string, error) {
	var selected []string
	commonPackages := []string{
		"git",
		"htop",
		"jq",
		"curl",
		"vim",
		"docker",
		"tmux",
		"fzf",
		"zsh",
		"neovim",
	}

	prompt := &survey.MultiSelect{
		Message: "Select the common packages and tools you'd like to include in your configuration:",
		Options: commonPackages,
	}

	err := survey.AskOne(prompt, &selected)
	if err != nil {
		return nil, err
	}

	fmt.Println("You can add or remove packages later by modifying your configuration or using the CLI.")
	return selected, nil
}
