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

func NewProjectUpdateCommand() *cobra.Command {
	var team string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update project configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			spin := progress.NewSpinner("Updating project configuration...")
			spin.Start()
			if err := projectSvc.UpdateProjectConfig(team); err != nil {
				spin.Fail("Update failed")
				return fmt.Errorf("project update failed: %w", err)
			}
			spin.Success("Project updated")

			fmt.Println("\n🔄 Run 'nix-foundry apply' to apply changes")
			return nil
		},
	}

	cmd.Flags().StringVarP(&team, "team", "t", "",
		"Team configuration to merge into project")

	return cmd
}
