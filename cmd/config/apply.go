package config

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/shawnkhoffman/nix-foundry/service/config"
	"github.com/spf13/cobra"
)

func NewApplyCmd() *cobra.Command {
	var (
		configService = config.NewConfigService(filesystem.NewOSFileSystem())
		applyService  = config.NewApplyService(filesystem.NewOSFileSystem())
	)

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply the current configuration",
		Long:  `Applies the current configuration by generating and activating the Nix configuration.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get active config path
			activeConfigPath, err := configService.ActiveConfigPath()
			if err != nil {
				return fmt.Errorf("failed to get active config path: %w", err)
			}

			// Validate configuration first
			if err := configService.ValidateConfig(activeConfigPath); err != nil {
				return fmt.Errorf("invalid configuration: %w", err)
			}

			// Generate and activate configuration
			if err := applyService.ActivateConfig(activeConfigPath); err != nil {
				return fmt.Errorf("failed to activate configuration: %w", err)
			}

			fmt.Println("Configuration successfully applied!")
			return nil
		},
	}
	return cmd
}
