package backup

import (
	"fmt"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/services/backup"
	"github.com/spf13/cobra"
)

func NewListCmd(svc backup.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available backups",
		Long: `List all available nix-foundry configuration backups.
Shows backup name, creation time, and size.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			backups, err := svc.List()
			if err != nil {
				return fmt.Errorf("failed to list backups: %w", err)
			}

			if len(backups) == 0 {
				fmt.Println("No backups found")
				return nil
			}

			fmt.Println("Available backups:")
			for _, b := range backups {
				fmt.Printf("- %s (created: %s, size: %d bytes)\n",
					b.Name,
					b.CreatedAt.Format(time.RFC3339),
					b.Size,
				)
			}

			return nil
		},
	}
}
