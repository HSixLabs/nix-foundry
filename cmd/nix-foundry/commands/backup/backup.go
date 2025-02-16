package backup

import (
	"github.com/shawnkhoffman/nix-foundry/internal/services/backup"
	"github.com/spf13/cobra"
)

// NewCmd creates the backup command group with all subcommands
func NewCmd(svc backup.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Manage environment backups",
		Long: `Manage environment backups including creation, restoration, and maintenance.
Examples:
  nix-foundry backup create daily-backup
  nix-foundry backup restore last-working
  nix-foundry backup list
  nix-foundry backup rotate`,
	}

	cmd.AddCommand(
		NewCreateCmd(svc),
		NewRestoreCommand(svc),
		NewListCmd(svc),
		NewDeleteCmd(svc),
		NewRotateCmd(svc),
		newConfigCmd(svc),
		newEncryptCmd(svc),
		newDecryptCmd(svc),
	)

	return cmd
}
