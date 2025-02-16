package backup

import (
	"fmt"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/services/backup"
	"github.com/shawnkhoffman/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
)

func NewRestoreCommand(svc backup.Service) *cobra.Command {
	var (
		force      bool
		keepBackup time.Duration
	)

	cmd := &cobra.Command{
		Use:   "restore [backup-id]",
		Short: "Restore environment from backup",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var backupID string

			// Handle latest backup if no ID provided
			if len(args) == 0 {
				backups, err := svc.List()
				if err != nil {
					return fmt.Errorf("failed to find latest backup: %w", err)
				}
				if len(backups) == 0 {
					return fmt.Errorf("no backups available for restore")
				}
				backupID = backups[len(backups)-1].ID
			} else {
				backupID = args[0]
			}

			// Create safety backup
			spin := progress.NewSpinner("Creating safety backup")
			spin.Start()
			if !force {
				if _, err := svc.CreateSafetyBackup(); err != nil {
					spin.Fail("Safety backup failed")
					return fmt.Errorf("safety backup failed: %w", err)
				}
				spin.Success("Safety backup created")
			}

			// Perform atomic restore
			spin = progress.NewSpinner("Performing atomic restore...")
			spin.Start()
			if err := svc.Restore(backupID, force); err != nil {
				spin.Fail("Atomic restore failed")
				return fmt.Errorf("restore failed: %w", err)
			}
			spin.Success("Environment restored")

			fmt.Println("\n✅ Restoration complete!")
			fmt.Println("   Run 'nix-foundry apply' to activate the restored environment")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip safety backup")
	cmd.Flags().DurationVar(&keepBackup, "keep-backups", 7*24*time.Hour, "Duration to keep safety backups")

	return cmd
}
