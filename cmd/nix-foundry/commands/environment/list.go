package environment

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/spf13/cobra"
)

func NewListCommand(svc environment.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available environments",
		RunE: func(cmd *cobra.Command, args []string) error {
			envs := svc.ListEnvironments()

			fmt.Println("Available environments:")
			for _, env := range envs {
				fmt.Printf(" - %s\n", env)
			}
			return nil
		},
	}
}
