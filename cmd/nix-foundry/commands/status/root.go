package status

import (
	"github.com/spf13/cobra"
)

func NewStatusCommandGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check system and environment status",
	}

	cmd.AddCommand(
		NewEnvironmentCommand(),
		NewSystemCommand(),
		NewConfigCommand(),
	)

	return cmd
}
