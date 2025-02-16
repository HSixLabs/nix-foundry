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
	}

	cmd.AddCommand(
		NewCreateCmd(svc),
		NewRestoreCmd(svc),
		NewListCmd(svc),
		NewDeleteCmd(svc),
		NewRotateCmd(svc),
		newConfigCmd(svc),
		newDecryptCmd(svc),
		newEncryptCmd(svc),
	)

	return cmd
}

// Removed duplicate function declarations - implementations exist in restore.go and list.go
