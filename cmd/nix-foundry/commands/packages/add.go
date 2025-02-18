package packages

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"github.com/shawnkhoffman/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
)

func NewAddCommand(pkgSvc packages.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "add [package...]",
		Short: "Add custom packages",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			spin := progress.NewSpinner("Adding packages...")
			spin.Start()

			if err := pkgSvc.AddPackages(ctx, args); err != nil {
				spin.Fail("Failed to add packages")
				return fmt.Errorf("failed to add packages: %w", err)
			}

			spin.Success("Packages added")
			fmt.Println("\nℹ️  Run 'nix-foundry update' to apply changes")
			return nil
		},
	}
}
