package config

import (
	"fmt"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/spf13/cobra"
)

func NewResetCommand(cfgSvc config.Service) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "reset <key>",
		Short: "Reset configuration values to defaults",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := strings.ToLower(args[0])
			if !force {
				return fmt.Errorf("reset requires --force flag")
			}
			return cfgSvc.ResetValue(key)
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Confirm reset operation")
	return cmd
}
