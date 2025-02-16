package backup

import (
	"fmt"
	"github.com/shawnkhoffman/nix-foundry/internal/services/backup"
	"github.com/spf13/cobra"
)

func newDecryptCmd(svc backup.Service) *cobra.Command {
	var keyPath string
	cmd := &cobra.Command{
		Use:   "decrypt <backup>",
		Short: "Decrypt an encrypted backup",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirmAction() {
				fmt.Println("Decryption cancelled")
				return nil
			}
			return svc.DecryptBackup(args[0], keyPath)
		},
	}
	cmd.Flags().StringVar(&keyPath, "key", "", "Path to decryption key")
	return cmd
}
