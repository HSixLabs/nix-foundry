package environment

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/spf13/cobra"
)

func NewCreateCommand(svc environment.Service) *cobra.Command {
	var template string

	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create new environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := svc.CreateEnvironment(name, template); err != nil {
				return fmt.Errorf("creation failed: %w", err)
			}

			fmt.Printf("Successfully created environment '%s'\n", name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&template, "template", "t", "default", "Environment template to use")
	return cmd
}
