package profile

import (
	"fmt"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/profile"
	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			configSvc := config.NewService()
			profileDir := filepath.Join(configSvc.GetConfigDir(), "profiles")
			svc := profile.NewService(profileDir)

			profiles, err := svc.List()
			if err != nil {
				return fmt.Errorf("failed to list profiles: %w", err)
			}

			if len(profiles) == 0 {
				fmt.Println("No profiles found")
				return nil
			}

			fmt.Println("Available profiles:")
			for _, p := range profiles {
				fmt.Printf("  - %s\n", p)
			}
			return nil
		},
	}
	return cmd
}
