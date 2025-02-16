package uninstall

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/internal/services/uninstall"
	"github.com/shawnkhoffman/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
)

func NewUninstallCommand() *cobra.Command {
	var keepBackups bool

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove nix-foundry and all managed environments",
		RunE: func(cmd *cobra.Command, args []string) error {
			uninstallSvc := uninstall.NewService()

			fmt.Println("🚨 This will permanently remove:")
			fmt.Println(" - All nix-foundry environments")
			fmt.Println(" - Project configurations")
			fmt.Println(" - Cached packages")
			if !keepBackups {
				fmt.Println(" - All backups")
			}

			fmt.Print("\nContinue? [y/N]: ")
			var confirm string
			if _, err := fmt.Scanln(&confirm); err != nil {
				fmt.Println("Invalid input, uninstall cancelled")
				return nil
			}
			if confirm != "y" && confirm != "Y" {
				fmt.Println("Uninstall cancelled")
				return nil
			}

			spin := progress.NewSpinner("Removing nix-foundry...")
			spin.Start()
			if err := uninstallSvc.Execute(keepBackups); err != nil {
				spin.Fail("Uninstall failed")
				return fmt.Errorf("uninstall failed: %w", err)
			}
			spin.Success("nix-foundry removed")

			fmt.Println("\n💔 Successfully uninstalled. Backup retention: ", keepBackups)
			return nil
		},
	}

	cmd.Flags().BoolVar(&keepBackups, "keep-backups", false,
		"Preserve backup history during uninstall")
	return cmd
}
