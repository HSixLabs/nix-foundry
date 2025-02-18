package packages

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"github.com/shawnkhoffman/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
)

func NewRemoveCommand(pkgSvc packages.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "remove [package...]",
		Short: "Remove custom packages",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			spin := progress.NewSpinner("Removing packages...")
			spin.Start()

			if err := pkgSvc.RemovePackages(ctx, args); err != nil {
				spin.Fail("Failed to remove packages")
				return fmt.Errorf("failed to remove packages: %w", err)
			}

			spin.Success("Packages removed")
			fmt.Println("\nℹ️  Run 'nix-foundry update' to apply changes")
			return nil
		},
	}
}
