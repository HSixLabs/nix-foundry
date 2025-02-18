package config

import (
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/spf13/cobra"
)

func NewValidateCommand(cfgSvc config.Service) *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cfgSvc.ValidateConfiguration(verbose)
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed validation output")
	return cmd
}
