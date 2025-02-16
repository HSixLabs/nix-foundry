package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/spf13/cobra"
)

func newSetCmd(cfgSvc config.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long: `Set the value of a configuration setting.
Examples:
  nix-foundry config set backup.maxBackups 20
  nix-foundry config set environment.default prod`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := parseValue(key, args[1])

			if err := cfgSvc.SetValue(key, value); err != nil {
				return fmt.Errorf("failed to set configuration value: %w", err)
			}

			if err := cfgSvc.Save(); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			fmt.Printf("✅ Set %s = %v\n", key, value)
			return nil
		},
	}

	return cmd
}

func parseValue(key, input string) interface{} {
	if strings.HasPrefix(key, "backup.") {
		if intVal, err := strconv.Atoi(input); err == nil {
			return intVal
		}
	}
	return input
}
