package rollback

import (
	"github.com/spf13/cobra"
)

func NewRollbackCommand() *cobra.Command {
	var (
		force    bool
		list     bool
		backupID string
	)

	cmd := &cobra.Command{
		Use:   "rollback [backup-id]",
		Short: "Revert to last working state",
		Long:  "Revert to the last known working state using environment backups",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				backupID = args[0]
			}

			if list {
				return NewRollbackListCommand().RunE(cmd, args)
			}
			return NewRollbackApplyCommand().RunE(cmd, []string{backupID})
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")
	cmd.Flags().BoolVarP(&list, "list", "l", false, "List backups")

	cmd.AddCommand(
		NewRollbackListCommand(),
		NewRollbackApplyCommand(),
	)

	return cmd
}
