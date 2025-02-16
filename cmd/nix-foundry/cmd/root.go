package cmd

import (
	backupcmd "github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/commands/backup"
	envcmd "github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/commands/environment"
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
	platformSvc := platform.NewService()

	envSvc := environment.NewService(
		cfgManager.GetConfigDir(),
		cfgManager,
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
	}

	rootCmd.AddCommand(
		backupcmd.NewCmd(backupSvc),
		envcmd.NewCmd(envSvc),
		// ... other command groups
	)

	return rootCmd
}
