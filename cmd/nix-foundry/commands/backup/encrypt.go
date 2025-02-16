package backup

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/internal/services/backup"
	"github.com/spf13/cobra"
)

func newEncryptCmd(svc backup.Service) *cobra.Command {
	var keyPath string
	cmd := &cobra.Command{
		Use:   "encrypt <backup>",
		Short: "Encrypt a backup with age encryption",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirmAction() {
				fmt.Println("Encryption cancelled")
				return nil
			}
			return svc.EncryptBackup(args[0], keyPath)
		},
	}
	cmd.Flags().StringVar(&keyPath, "key", "", "Path to encryption key")
	return cmd
}
