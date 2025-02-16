package backup

import (
	"fmt"

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

func NewRestoreCmd(svc backup.Service) *cobra.Command {
	var (
		force    bool
		testMode bool
	)

	cmd := &cobra.Command{
		Use:   "restore [name]",
		Short: "Restore a backup",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize dependencies
			cfgManager := config.NewService()
			validator := validation.NewService()
			platformSvc := platform.NewService()
			envSvc := environment.NewService(
				cfgManager.GetConfigDir(),
				cfgManager,
				validator,
				platformSvc,
			)
			pkgSvc := packages.NewService(cfgManager.GetConfigDir())
			projectSvc := project.NewService(cfgManager, envSvc, pkgSvc)
			backupSvc := backup.NewService(cfgManager, envSvc, projectSvc)

			if len(args) == 0 {
				return fmt.Errorf("backup name required")
			}
			name := args[0]

			spin := progress.NewSpinner("Restoring backup...")
			spin.Start()
			defer spin.Stop()

			if err := backupSvc.RestoreBackup(name); err != nil {
				spin.Fail("Restore failed")
				return fmt.Errorf("restore failed: %w", err)
			}

			spin.Success("Backup restored")
			return nil
		},
	}

	cmd.Flags().BoolVar(&testMode, "test", false, "Enable test mode")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force restore")

	return cmd
}
