package packages

import (
	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"github.com/spf13/cobra"
)

func NewPackagesCommand(pkgSvc packages.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "packages",
		Short: "Manage custom packages",
		Long:  `Add, remove, or list custom packages in your nix-foundry environment.`,
	}

	cmd.AddCommand(
		NewListCommand(pkgSvc),
		NewAddCommand(pkgSvc),
		NewRemoveCommand(pkgSvc),
	)
	return cmd
}
