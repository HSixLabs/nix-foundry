package cmd

import (
	backupCmd "github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/commands/backup"
	envcmd "github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/commands/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/validation"
	"github.com/shawnkhoffman/nix-foundry/internal/services/backup"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
	"github.com/spf13/cobra"
)

func InitializeServices() (backup.Service, environment.Service) {
	cfgManager := config.NewService()
	validator := validation.NewService()
	platformSvc := platform.NewService()

	envSvc := environment.NewService(
		cfgManager.GetConfigDir(),
		cfgManager,
		validator,
		platformSvc,
	)

	pkgSvc := packages.NewService(cfgManager.GetConfigDir())
	projectSvc := project.NewService(cfgManager, envSvc, pkgSvc)

	return backup.NewService(cfgManager, envSvc, projectSvc), envSvc
}

func NewRootCommand() *cobra.Command {
	backupSvc, envSvc := InitializeServices()

	rootCmd := &cobra.Command{
		Use:   "nix-foundry",
		Short: "Nix environment management toolkit",
		Run: func(cmd *cobra.Command, args []string) {
			// Placeholder for the root command
		},
	}

	rootCmd.AddCommand(
		backupCmd.NewCmd(backupSvc),
		envcmd.NewCmd(envSvc),
		// ... other command groups
	)

	return rootCmd
}
