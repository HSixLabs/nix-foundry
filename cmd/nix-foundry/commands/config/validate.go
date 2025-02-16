package config

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/internal/flags"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/spf13/cobra"
)

func NewValidateCommand() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			service := config.NewService()

			if err := service.ValidateConfiguration(); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}

			if verbose {
				fmt.Println("Configuration details:")
				fmt.Println("- Valid shell configuration")
				fmt.Println("- Consistent package versions")
				fmt.Println("- Proper environment links")
			}

			fmt.Println("✅ Configuration is valid")
			return nil
		},
	}

	// Use shared flags package
	flags.AddVerboseFlag(cmd, &verbose)

	return cmd
}
