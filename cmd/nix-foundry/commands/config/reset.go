package config

import (
	"fmt"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/spf13/cobra"
)

func newResetCmd(cfgSvc config.Service) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "reset [section]",
		Short: "Reset configuration to defaults",
		Long: `Reset configuration settings to their default values.
Optionally specify a section to reset only that part.

Examples:
  nix-foundry config reset
  nix-foundry config reset backup
  nix-foundry config reset environment`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Print("This will reset configuration to defaults. Continue? [y/N]: ")
				if !confirmAction() {
					fmt.Println("Reset cancelled")
					return nil
				}
			}

			section := ""
			if len(args) > 0 {
				section = args[0]
			}

			if err := cfgSvc.Reset(section); err != nil {
				return fmt.Errorf("failed to reset configuration: %w", err)
			}

			fmt.Println("✅ Configuration reset successfully")
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	return cmd
}

func confirmAction() bool {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false
	}
	return strings.EqualFold(response, "y") || strings.EqualFold(response, "yes")
}
