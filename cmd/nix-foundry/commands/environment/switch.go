package environment

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/internal/flags"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/spf13/cobra"
)

func NewSwitchCommand(svc environment.Service) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "switch [name]",
		Short: "Switch between environments",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			if err := svc.Switch(target, force); err != nil {
				if !force {
					fmt.Printf("Switch failed: %v\nUse --force to override conflicts", err)
					return err
				}
				return fmt.Errorf("forced switch failed: %w", err)
			}

			fmt.Printf("Successfully switched to '%s' environment\n", target)
			return nil
		},
	}

	flags.AddForceFlag(cmd)
	return cmd
}
