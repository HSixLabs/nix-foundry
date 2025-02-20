package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
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
			svc := config.NewConfigService(filesystem.NewOSFileSystem())

			currentShell := os.Getenv("SHELL")
			if currentShell == "" {
				currentShell = "/bin/sh"
			}

			selectedShell, err := promptForShell(currentShell)
			if err != nil {
				return fmt.Errorf("failed to get shell selection: %w", err)
			}

			selectedPackages, err := promptForPackages()
			if err != nil {
				return fmt.Errorf("failed to get package selections: %w", err)
			}

			confirmed, err := confirmSelections(selectedShell, selectedPackages)
			if err != nil {
				return fmt.Errorf("confirmation failed: %w", err)
			}
			if !confirmed {
				return nil
			}

			defaultKind := "user"
			defaultName := "default"
			basePath := ""

			if err := svc.InitConfig(defaultKind, defaultName, force, false, basePath); err != nil {
				return fmt.Errorf("failed to setup initial config: %w", err)
			}

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			activeConfigPath := filepath.Join(homeDir, ".config", "nix-foundry", "config.yaml")
			createdConfigPath := filepath.Join(homeDir, ".config", "nix-foundry", "users", defaultName, "config.yaml")

			if copyErr := svc.CopyConfig(createdConfigPath, activeConfigPath); copyErr != nil {
				return fmt.Errorf("failed to set active config: %w", copyErr)
			}

			if err := svc.UpdateActiveConfigWithPackages(activeConfigPath, selectedPackages, selectedShell); err != nil {
				return fmt.Errorf("failed to update active config with packages: %w", err)
			}

			fmt.Println("Successfully set up the initial configuration at \033[33m~/.config/nix-foundry/config.yaml\033[0m")
			fmt.Println("You can customize your configuration further by modifying the \033[33mconfig.yaml\033[0m file or using the CLI commands.")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing active configuration")
	return cmd
}

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
		"neovim",
	}

	prompt := &survey.MultiSelect{
		Message: "Select the packages and tools you'd like to start with:",
		Options: commonPackages,
	}

	err := survey.AskOne(prompt, &selected)
	if err != nil {
		return nil, err
	}

	fmt.Println("You can add or remove packages later by modifying your configuration or using the CLI.")
	return selected, nil
}

func promptForShell(currentShell string) (string, error) {
	baseName := filepath.Base(currentShell)
	cleanName := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	cleanName = strings.Split(cleanName, "-")[0]

	options := []string{
		fmt.Sprintf("Current shell (%s)", cleanName),
		"zsh",
		"fish",
		"bash",
		"I'll choose one later",
	}

	var selected string
	prompt := &survey.Select{
		Message: "Choose your development shell:",
		Options: options,
		Default: options[0],
	}

	if err := survey.AskOne(prompt, &selected); err != nil {
		return "", err
	}

	switch {
	case selected == options[0]:
		return currentShell, nil
	case selected == "I'll choose one later":
		return "", nil
	default:
		return selected, nil
	}
}

func confirmSelections(shell string, packages []string) (bool, error) {
	baseShell := filepath.Base(shell)
	if shell == "" {
		baseShell = "not selected"
	}

	message := fmt.Sprintf(
		"\033[1mAbout to apply configuration with:\033[0m\n\n"+
			"- Shell: \033[36m%s\033[0m\n"+
			"- Packages: \033[36m%s\033[0m\n\n"+
			"(You'll be able to change these later)\n\n"+
			"Would you like to proceed?",
		baseShell,
		packages,
	)

	var confirm bool
	prompt := &survey.Confirm{
		Message: message,
		Default: true,
	}

	err := survey.AskOne(prompt, &confirm)
	return confirm, err
}
