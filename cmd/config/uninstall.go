package config

import (
	"fmt"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/shawnkhoffman/nix-foundry/service/config"
	"github.com/spf13/cobra"
)

func NewUninstallCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Completely remove Nix Foundry and all configurations",
		Long: `This will:
- Remove all configuration files
- Remove generated Nix files
- Clean up any related directories

Use with caution! This action is irreversible.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := config.NewConfigService(filesystem.NewOSFileSystem())

			activePath, err := svc.ActiveConfigPath()
			if err != nil {
				return fmt.Errorf("failed to locate config: %w", err)
			}
			configDir := filepath.Dir(activePath)

			if !force {
				confirmed, err := confirmUninstall()
				if err != nil {
					return fmt.Errorf("confirmation failed: %w", err)
				}
				if !confirmed {
					fmt.Println("Uninstall cancelled")
					return nil
				}
			}

			if err := filesystem.NewOSFileSystem().RemoveAll(configDir); err != nil {
				return fmt.Errorf("failed to remove config files: %w", err)
			}

			fmt.Println("\n🧹 Nix Foundry successfully uninstalled!")
			fmt.Println("Sorry to see you go 😢")
			fmt.Println("You can reinstall anytime by following the original installation instructions")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}

func confirmUninstall() (bool, error) {
	prompt := &survey.Confirm{
		Message: "\033[1mUninstalling Nix Foundry!\033[0m\n\n" +
			"This will permanently remove ALL Nix Foundry files and configurations.\n\n" +
			"Are you sure you want to continue?",
		Default: false,
	}
	var response bool
	err := survey.AskOne(prompt, &response)
	return response, err
}
