package project

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/validation"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
	"github.com/shawnkhoffman/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
)

func NewProjectInitCommand() *cobra.Command {
	var (
		team  string
		force bool
	)

	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Initialize a new project environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]

			// Initialize required services
			cfgSvc := config.NewService()
			validator := validation.NewService()
			platformSvc := platform.NewService()

			envSvc := environment.NewService(
				cfgSvc.GetConfigDir(),
				cfgSvc,
				validator,
				platformSvc,
			)

			pkgSvc := packages.NewService(cfgSvc.GetConfigDir())
			projectSvc := project.NewService(cfgSvc, envSvc, pkgSvc)

			spin := progress.NewSpinner("Initializing project...")
			spin.Start()

			if err := projectSvc.InitializeProject(projectName, team, force); err != nil {
				spin.Fail("Project initialization failed")
				return fmt.Errorf("error initializing project: %w", err)
			}

			spin.Success(fmt.Sprintf("Project '%s' initialized", projectName))
			fmt.Println("\nℹ️  Run 'nix-foundry project update' to apply team configurations")
			return nil
		},
	}

	cmd.Flags().StringVarP(&team, "team", "t", "", "Team configuration to use")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing configuration")
	return cmd
}
