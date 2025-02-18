package status

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/spf13/cobra"
)

func NewConfigCommand() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Validate configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgSvc := config.NewService()
			err := cfgSvc.ValidateConfiguration(verbose)
			if err != nil && !verbose { // Only show error if not in verbose mode
				return fmt.Errorf("config validation failed: %w", err)
			}
			if verbose && err == nil {
				fmt.Println("✅ Configuration is valid")
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed validation output")
	return cmd
}
