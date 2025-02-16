package backup

import (
	"fmt"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/services/backup"
	"github.com/shawnkhoffman/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
)

func NewRotateCmd(svc backup.Service) *cobra.Command {
	var maxAge time.Duration

	cmd := &cobra.Command{
		Use:   "rotate",
		Short: "Rotate old backups",
		RunE: func(cmd *cobra.Command, args []string) error {
			spin := progress.NewSpinner("Rotating old backups...")
			spin.Start()

			if err := svc.Rotate(maxAge); err != nil {
				spin.Fail("Rotation failed")
				return fmt.Errorf("failed to rotate backups: %w", err)
			}

			spin.Success("Backups rotated")
			return nil
		},
	}

	cmd.Flags().DurationVar(&maxAge, "max-age", 7*24*time.Hour, "Maximum age of backups to keep")
	return cmd
}
