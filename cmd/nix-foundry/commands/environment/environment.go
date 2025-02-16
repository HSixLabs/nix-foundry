package environment

import (
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/spf13/cobra"
)

func NewCmd(svc environment.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "environment",
		Short: "Manage nix environments",
	}

	cmd.AddCommand(
		NewCreateCommand(svc),
		NewSwitchCommand(svc),
		NewListCommand(svc),
		NewRollbackCommand(svc),
	)

	return cmd
}
