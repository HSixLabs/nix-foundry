package config

import (
	"fmt"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/spf13/cobra"
)

func NewApplyCommand(cfgSvc config.Service) *cobra.Command {
	var (
		configPath string
		testMode   bool
	)

	cmd := &cobra.Command{
		Use:   "apply [flags]",
		Short: "Apply a configuration",
		Long:  "Apply and activate the configuration using home-manager",
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no config path specified, use default
			if configPath == "" {
				configPath = filepath.Join(cfgSvc.GetConfigDir(), "config.yaml")
			}

			// Load the configuration
			if err := cfgSvc.ReadConfig(configPath, cfgSvc.GetConfig()); err != nil {
				return fmt.Errorf("failed to read configuration: %w", err)
			}

			// Apply the configuration
			if err := cfgSvc.Apply(cfgSvc.GetConfig(), testMode); err != nil {
				return fmt.Errorf("failed to apply configuration: %w", err)
			}

			// Initialize environment service
			envService := environment.NewService(
				cfgSvc.GetConfigDir(),
				cfgSvc,
				platform.NewService(),
				testMode,
				true,
				false, // autoInstall
			)

			// Apply the configuration through home-manager
			if err := envService.ApplyConfiguration(); err != nil {
				return fmt.Errorf("failed to activate configuration: %w", err)
			}

			fmt.Println("✅ Configuration applied successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "Path to configuration file (default: $HOME/.config/nix-foundry/config.yaml)")
	cmd.Flags().BoolVar(&testMode, "test", false, "Run in test mode")
	return cmd
}
