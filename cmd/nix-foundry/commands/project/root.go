package project

import (
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
	"github.com/spf13/cobra"
)

func NewProjectCommandGroup(svc project.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage project configurations",
	}

	cmd.AddCommand(
		NewProjectInitCommand(),
		NewProjectUpdateCommand(),
	)

	return cmd
}
