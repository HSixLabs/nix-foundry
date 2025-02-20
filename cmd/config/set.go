package config

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/shawnkhoffman/nix-foundry/pkg/schema"
	"github.com/shawnkhoffman/nix-foundry/service/config"
	"github.com/spf13/cobra"
)

func NewSetCmd() *cobra.Command {
	var (
		shell    string
		packages []string
	)

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Update configuration values",
		Long: `Modifies specific values in the active configuration.

Examples:
  nix-foundry config set --shell /bin/zsh
  nix-foundry config set --package tmux --package neovim`,
		RunE: func(cmd *cobra.Command, args []string) error {
			svc := config.NewConfigService(filesystem.NewOSFileSystem())

			activePath, err := svc.ActiveConfigPath()
			if err != nil {
				return fmt.Errorf("failed to get active config: %w", err)
			}

			updates := make([]config.UpdateFunc, 0)

			if shell != "" {
				updates = append(updates, func(c *schema.Config) {
					c.Nix.Shell = shell
				})
			}

			if len(packages) > 0 {
				updates = append(updates, func(c *schema.Config) {
					c.Nix.Packages.Core = append(c.Nix.Packages.Core, packages...)
					c.Nix.Packages.Core = config.Unique(c.Nix.Packages.Core)
				})
			}

			if len(updates) == 0 {
				return fmt.Errorf("no configuration changes specified")
			}

			if err := svc.UpdateConfig(activePath, updates...); err != nil {
				return fmt.Errorf("failed to update config: %w", err)
			}

			fmt.Println("\n🎯 Configuration updated successfully!")
			fmt.Println("Run \033[36mnix-foundry apply\033[0m to apply your changes")
			return nil
		},
	}

	cmd.Flags().StringVar(&shell, "shell", "", "Set development shell path")
	cmd.Flags().StringSliceVarP(&packages, "package", "p", []string{}, "Add package to core packages")
	return cmd
}
