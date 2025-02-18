package config

import (
	"fmt"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/spf13/cobra"
)

func NewSetCommand(cfgSvc config.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set configuration values",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := strings.ToLower(args[0])
			value := strings.Join(args[1:], " ")

			switch key {
			case "shell":
				return cfgSvc.SetValue("shell.type", value)
			case "editor":
				return cfgSvc.SetValue("editor.type", value)
			case "git.name":
				return cfgSvc.SetValue("git.name", value)
			case "git.email":
				return cfgSvc.SetValue("git.email", value)
			default:
				return fmt.Errorf("invalid config key: %s", key)
			}
		},
	}

	return cmd
}
