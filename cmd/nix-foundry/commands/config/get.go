package config

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/spf13/cobra"
)

func newGetCmd(cfgSvc config.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Long: `Get the value of a configuration setting.
Examples:
  nix-foundry config get backup.maxBackups
  nix-foundry config get environment.default`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value, err := cfgSvc.GetValue(key)
			if err != nil {
				return fmt.Errorf("failed to get configuration value: %w", err)
			}
			fmt.Printf("%s = %v\n", key, value)
			return nil
		},
	}

	return cmd
}
