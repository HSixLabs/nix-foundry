package config

import (
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/spf13/cobra"
)

func NewConfigCommand(cfgSvc config.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration settings",
	}

	cmd.AddCommand(
		newSetCmd(cfgSvc),
		newGetCmd(cfgSvc),
		newResetCmd(cfgSvc),
		newInitCommand(cfgSvc),
		NewApplyCommand(),
	)

	keyCmd := newKeyCmd()
	keyCmd.AddCommand(
		newKeyRotateCmd(),
		newKeyGenerateCmd(),
	)
	cmd.AddCommand(keyCmd)

	return cmd
}
