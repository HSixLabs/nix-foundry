package cmd

import (
	"github.com/shawnkhoffman/nix-foundry/cmd/config"

	"github.com/spf13/cobra"
)

func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage Nix configurations",
	}

	cmd.AddCommand(config.NewSetCmd())
	cmd.AddCommand(config.NewSetupCmd())
	cmd.AddCommand(config.NewInitCmd())
	cmd.AddCommand(config.NewApplyCmd())
	cmd.AddCommand(config.NewUninstallCmd())
	return cmd
}
