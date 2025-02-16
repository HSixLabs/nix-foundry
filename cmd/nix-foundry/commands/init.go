package commands

import (
	"fmt"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/validation"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/shawnkhoffman/nix-foundry/internal/services/update"
	"github.com/shawnkhoffman/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func NewInitCommand() *cobra.Command {
	var (
		forceConfig bool
		autoConfig  bool
		testMode    bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize nix-foundry configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize platform service
			platformSvc := platform.NewService()

			// Initialize configuration service first
			configSvc := config.NewService()

			// Initialize environment service with proper config directory
			envSvc := environment.NewService(
				configSvc.GetConfigDir(),
				configSvc,
				validation.NewService(),
				platformSvc,
			)

			// Setup environment
			if err := envSvc.Initialize(testMode); err != nil {
				return fmt.Errorf("failed to initialize environment: %w", err)
			}

			// Generate and preview configuration
			nixConfig, err := configSvc.GenerateInitialConfig("zsh", "nano", "", "")
			if err != nil {
				return fmt.Errorf("failed to generate configuration: %w", err)
			}

			// Show preview
			fmt.Print(configSvc.PreviewConfiguration(nixConfig))

			// Apply configuration if not in interactive mode or user confirms
			var spin *progress.Spinner
			if autoConfig || testMode || confirmApply() {
				spin = progress.NewSpinner("Applying configuration...")
				spin.Start()
				updateSvc := update.NewService()
				if err := updateSvc.ApplyConfiguration(configSvc.GetConfigDir(), testMode); err != nil {
					spin.Fail("Failed to apply configuration")
					return err
				}
				spin.Success("Configuration applied")
			}

			// Convert flags to configuration map, excluding special flags
			configMap := make(map[string]string)
			cmd.Flags().VisitAll(func(f *pflag.Flag) {
				if f.Name != "test" && f.Name != "yes" {
					if f.Value.String() != "" {
						configMap[f.Name] = f.Value.String()
					}
				}
			})

			// Create config object from map
			mergedConfig := configSvc.CreateConfigFromMap(configMap)

			// Apply configuration using the config service
			if err := configSvc.Apply(mergedConfig, testMode); err != nil {
				if spin != nil {
					spin.Fail("Failed to apply configuration")
				}
				return fmt.Errorf("failed to apply configuration: %w", err)
			}
			if spin != nil {
				spin.Success("Configuration applied successfully")
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&forceConfig, "force", false, "Force configuration overwrite")
	cmd.Flags().BoolVar(&autoConfig, "yes", false, "Automatically confirm configuration")
	cmd.Flags().BoolVar(&testMode, "test", false, "Run in test mode")

	return cmd
}

func confirmApply() bool {
	fmt.Print("\nWould you like to apply this configuration? [y/N]: ")
	var confirm string
	if _, err := fmt.Scanln(&confirm); err != nil {
		return false
	}
	return strings.EqualFold(confirm, "y") || strings.EqualFold(confirm, "yes")
}
