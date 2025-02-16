package backup

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/internal/services/backup"
	"github.com/shawnkhoffman/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
)

func NewDeleteCmd(svc backup.Service) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a backup",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force && !confirmAction() {
				fmt.Println("Deletion cancelled")
				return nil
			}

			spin := progress.NewSpinner("Deleting backup...")
			spin.Start()

			if err := svc.Delete(args[0]); err != nil {
				spin.Fail("Delete failed")
				return fmt.Errorf("failed to delete backup: %w", err)
			}

			spin.Success("Backup deleted")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")
	return cmd
}
