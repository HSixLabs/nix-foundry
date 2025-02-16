package backup

import (
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/services/backup"
	"github.com/spf13/cobra"
)

func NewRotateCmd(svc backup.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "rotate",
		Short: "Rotate old backups",
		RunE: func(cmd *cobra.Command, args []string) error {
			retention := svc.GetConfig().RetentionDays
			return svc.Rotate(time.Duration(retention) * 24 * time.Hour)
		},
	}
}
