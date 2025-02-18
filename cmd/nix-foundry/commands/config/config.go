package config

import (
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
	"github.com/spf13/cobra"
)

func NewConfigCommand(
	cfgSvc config.Service,
	projectSvc project.Service,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration settings",
	}

	cmd.AddCommand(
		NewInitCommand(cfgSvc),
		NewSetCommand(cfgSvc),
		NewGetCommand(cfgSvc),
		NewValidateCommand(cfgSvc),
		NewApplyCommand(cfgSvc),
		NewResetCommand(cfgSvc),
		NewKeyCommand(cfgSvc),
		NewCheckConflictsCommand(projectSvc),
	)

	return cmd
}

func NewKeyCommand(cfgSvc config.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key",
		Short: "Manage encryption keys",
		Long:  "Generate and rotate encryption keys for secure configuration storage",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "generate",
			Short: "Generate new encryption key",
			RunE: func(cmd *cobra.Command, args []string) error {
				return cfgSvc.GenerateEncryptionKey()
			},
		},
		&cobra.Command{
			Use:   "rotate",
			Short: "Rotate existing encryption key",
			RunE: func(cmd *cobra.Command, args []string) error {
				return cfgSvc.RotateEncryptionKey()
			},
		},
	)

	return cmd
}
