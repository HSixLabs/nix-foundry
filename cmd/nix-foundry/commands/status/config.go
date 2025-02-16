package status

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/spf13/cobra"
)

func NewConfigCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Validate configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgSvc := config.NewService()
			if err := cfgSvc.ValidateConfiguration(); err != nil {
				return fmt.Errorf("config validation failed: %w", err)
			}
			fmt.Println("✅ Configuration is valid")
			return nil
		},
	}
}
