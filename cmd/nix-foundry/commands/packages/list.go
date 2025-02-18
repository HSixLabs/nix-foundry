package packages

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"github.com/spf13/cobra"
)

func NewListCommand(pkgSvc packages.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List custom packages",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			packages, err := pkgSvc.ListCustomPackages(ctx)
			if err != nil {
				return fmt.Errorf("failed to load custom packages: %w", err)
			}

			if len(packages) == 0 {
				fmt.Println("No custom packages configured")
				return nil
			}

			fmt.Println("Custom packages:")
			for i, pkg := range packages {
				fmt.Printf("%d. %s\n", i+1, pkg)
			}
			return nil
		},
	}
}
