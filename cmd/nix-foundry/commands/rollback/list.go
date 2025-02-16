package rollback

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

func NewRollbackListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available backups",
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

			spin := progress.NewSpinner("Loading backups...")
			spin.Start()
			backups, err := backupSvc.ListBackups()
			spin.Stop()

			if err != nil {
				return fmt.Errorf("failed to list backups: %w", err)
			}

			fmt.Println("\nAvailable Backups:")
			for _, backup := range backups {
				fmt.Printf("  ⏰ %s | ID: %s\n    📦 Size: %d bytes\n",
					backup.Timestamp.Format("2006-01-02 15:04:05"),
					backup.ID,
					backup.Size)
			}
			return nil
		},
	}
}
