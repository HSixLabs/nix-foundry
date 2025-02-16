package rollback

import (
	"fmt"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/validation"
	"github.com/shawnkhoffman/nix-foundry/internal/services/backup"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
	"github.com/shawnkhoffman/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
)

func NewRollbackApplyCommand() *cobra.Command {
	var (
		force      bool
		keepBackup time.Duration
	)

	cmd := &cobra.Command{
		Use:   "apply [backup-id]",
		Short: "Restore from backup",
		Args:  cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize required services
			cfgSvc := config.NewService()
			validator := validation.NewService()
			platformSvc := platform.NewService()

			envSvc := environment.NewService(
				cfgSvc.GetConfigDir(),
				cfgSvc,
				validator,
				platformSvc,
			)

			pkgSvc := packages.NewService(cfgSvc.GetConfigDir())
			projectSvc := project.NewService(cfgSvc, envSvc, pkgSvc)
			backupSvc := backup.NewService(cfgSvc, envSvc, projectSvc)

			var backupID string

			// Handle latest backup if no ID provided
			if len(args) == 0 {
				backups, err := backupSvc.ListBackups()
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

			// Safety confirmation
			if !force {
				fmt.Printf("Restore backup %q and overwrite current environment? [y/N]: ", backupID)
				var confirm string
				if _, err := fmt.Scanln(&confirm); err != nil || (confirm != "y" && confirm != "Y") {
					fmt.Println("Restore cancelled")
					return nil
				}
			}

			// Create safety backup (no ID parameter needed)
			spin := progress.NewSpinner("Creating safety backup")
			spin.Start()
			if _, err := backupSvc.CreateBackup(); err != nil {
				spin.Fail("Safety backup failed")
				return fmt.Errorf("safety backup failed: %w", err)
			}
			spin.Success("Safety backup created")

			// Perform atomic restore
			spin = progress.NewSpinner("Performing atomic restore...")
			spin.Start()
			if err := backupSvc.RestoreBackup(backupID); err != nil {
				spin.Fail("Atomic restore failed")
				return fmt.Errorf("restore failed: %w", err)
			}
			spin.Success("Environment restored")

			// Post-restore validation
			currentEnv := envSvc.GetCurrentEnvironment()
			if err := envSvc.ValidateRestoredEnvironment(currentEnv); err != nil {
				return fmt.Errorf("restore validation failed: %w", err)
			}

			// Cleanup old backups if requested
			if keepBackup > 0 {
				spin = progress.NewSpinner("Cleaning up old backups...")
				spin.Start()
				if err := backupSvc.Rotate(keepBackup); err != nil {
					spin.Fail("Backup rotation failed")
				} else {
					spin.Success("Backups cleaned")
				}
			}

			fmt.Println("\n✅ Restoration complete!")
			fmt.Println("   Run 'nix-foundry apply' to activate the restored environment")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip safety backup")
	cmd.Flags().DurationVar(&keepBackup, "keep-backups", 7*24*time.Hour,
		"Duration to keep safety backups")

	return cmd
}
