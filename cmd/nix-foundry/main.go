package main

import (
	"os"

	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/commands"
	configcmd "github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/commands/config"
	packagescmd "github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/commands/packages"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
)

// Add helper function for interface conversion
func asService(ps project.ProjectService) project.Service {
	return ps.(project.Service)
}

func main() {
	// Initialize core services first
	cfgSvc := config.NewService()
	platformSvc := platform.NewService()

	// Get config directory from config service
	configDir := cfgSvc.GetConfigDir()

	// Initialize environment service with proper dependencies
	envSvc := environment.NewService(
		configDir,
		cfgSvc,
		platformSvc,
		false, // testMode
		true,  // isolationEnabled
		true,  // autoInstall
	)

	// Initialize package service with config directory
	pkgSvc := packages.NewService(configDir)

	// Create project service with all dependencies
	projectSvc := project.NewService(cfgSvc, envSvc, pkgSvc)

	// Configure root command with proper services
	rootCmd := commands.RootCmd
	rootCmd.AddCommand(
		configcmd.NewInitCommand(cfgSvc),
		configcmd.NewConfigCommand(cfgSvc, asService(projectSvc)),
		packagescmd.NewPackagesCommand(pkgSvc),
		// Add other commands...
	)

	if err := rootCmd.Execute(); err != nil {
		logging.GetLogger().WithError(err).Error("Command failed")
		os.Exit(1)
	}
}
