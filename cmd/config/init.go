package config

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/shawnkhoffman/nix-foundry/service/config"

	"github.com/spf13/cobra"
)

func NewInitCmd() *cobra.Command {
	var force bool
	var newConfig bool

	cmd := &cobra.Command{
		Use:   "init [kind] [name]",
		Short: "Initialize a new configuration",
		Long: `Initializes a new configuration of specified type (user|team|project)

Examples:
  nix-foundry config init project myproject --base team/myteam
  nix-foundry config init user myuser --base project/myproject`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			kind := args[0]
			name := args[1]
			base, _ := cmd.Flags().GetString("base")

			svc := config.NewConfigService(filesystem.NewOSFileSystem())
			if err := svc.InitConfig(kind, name, force, newConfig, base); err != nil {
				return fmt.Errorf("failed to initialize config: %w", err)
			}

			fmt.Printf("Successfully initialized configuration '%s'\n", name)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing configuration")
	cmd.Flags().BoolVar(&newConfig, "new", false, "Create minimal new configuration")
	return cmd
}
