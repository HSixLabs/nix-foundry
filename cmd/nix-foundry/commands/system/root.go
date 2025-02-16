package commands

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/shawnkhoffman/nix-foundry/internal/services/update"
	"github.com/shawnkhoffman/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
)

func NewUpdateCommand() *cobra.Command {
	var testMode bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update nix-foundry packages and configuration",
		Long: `Update nix-foundry packages and apply the latest configuration.

This command will:
1. Update Nix flake inputs
2. Apply the updated configuration
3. Rebuild the environment

Example:
  nix-foundry update`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize required services
			configSvc := config.NewService()
			platformSvc := platform.NewService()
			envSvc := environment.NewService(
				configSvc.GetConfigDir(),
				configSvc,
				platformSvc,
			)

			// Update flake inputs
			spin := progress.NewSpinner("Updating Nix packages...")
			spin.Start()

			// Create update service with dependencies
			updateSvc := update.NewService(configSvc, envSvc)

			// Update flake
			if err := updateSvc.UpdateFlake(configSvc.GetConfigDir()); err != nil {
				spin.Fail("Failed to update packages")
				return fmt.Errorf("failed to update flake: %w", err)
			}
			spin.Success("Packages updated")

			// Apply updated configuration
			spin = progress.NewSpinner("Applying configuration...")
			spin.Start()
			if err := updateSvc.ApplyConfiguration(configSvc.GetConfigDir(), testMode); err != nil {
				spin.Fail("Failed to apply configuration")
				return fmt.Errorf("failed to apply configuration: %w", err)
			}
			spin.Success("Configuration applied")

			fmt.Println("\n✨ Environment updated successfully!")
			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&testMode, "test", false, "Run in test mode")

	return cmd
}
