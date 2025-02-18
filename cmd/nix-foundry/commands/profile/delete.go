package profile

import (
	"fmt"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/profile"
	"github.com/spf13/cobra"
)

func NewDeleteCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <name>",
		Args:  cobra.ExactArgs(1),
		Short: "Delete a profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			debug, _ := cmd.Root().Flags().GetBool("debug")
			logger := logging.GetLogger()
			if debug {
				logger.Debug("Starting profile deletion", "profile", args[0])
			}
			configSvc := config.NewService()
			profileDir := filepath.Join(configSvc.GetConfigDir(), "profiles")
			svc := profile.NewService(profileDir)

			if err := svc.Delete(args[0]); err != nil {
				if !force {
					return fmt.Errorf("deletion failed: %w (use --force to override)", err)
				}
			}

			fmt.Printf("Profile '%s' deleted\n", args[0])
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force deletion without confirmation")
	return cmd
}
