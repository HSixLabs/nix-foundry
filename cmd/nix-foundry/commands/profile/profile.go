package profile

import (
	"github.com/spf13/cobra"
)

func NewProfileCommandGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage environment profiles",
	}

	cmd.AddCommand(
		NewCreateCommand(),
		NewListCommand(),
		NewDeleteCommand(),
	)

	return cmd
}
