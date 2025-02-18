package config

import (
	"fmt"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/spf13/cobra"
)

func NewGetCommand(cfgSvc config.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get configuration values",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := strings.ToLower(args[0])
			value, err := cfgSvc.GetValue(key)
			if err != nil {
				return err
			}
			fmt.Println(value)
			return nil
		},
	}
	return cmd
}
