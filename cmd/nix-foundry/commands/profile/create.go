package profile

import (
	"fmt"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/profile"
	"github.com/spf13/cobra"
)

func NewCreateCommand() *cobra.Command {
	var (
		packages []string
		force    bool
	)

	cmd := &cobra.Command{
		Use:  "create <name>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configSvc := config.NewService()
			profileDir := filepath.Join(configSvc.GetConfigDir(), "profiles")
			svc := profile.NewService(profileDir)
			if err := svc.Create(args[0], packages, force); err != nil {
				return fmt.Errorf("creation failed: %w", err)
			}
			fmt.Printf("Profile '%s' created\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&packages, "packages", "p", nil, "Base packages")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite existing profile")
	return cmd
}
