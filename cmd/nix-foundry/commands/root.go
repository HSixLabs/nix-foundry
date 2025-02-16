package commands

import (
	"fmt"

	backupCmd "github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/commands/backup"
	configCmd "github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/commands/config"
	doctorCmd "github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/commands/doctor"
	environmentCmd "github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/commands/environment"
	profileCmd "github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/commands/profile"
	projectCmd "github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/commands/project"
	statusCmd "github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/commands/status"
	uninstallCmd "github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/commands/uninstall"
	"github.com/shawnkhoffman/nix-foundry/internal/flags"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/validation"
	backupSvc "github.com/shawnkhoffman/nix-foundry/internal/services/backup"
	configSvc "github.com/shawnkhoffman/nix-foundry/internal/services/config"
	environmentSvc "github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	projectSvc "github.com/shawnkhoffman/nix-foundry/internal/services/project"
	"github.com/spf13/cobra"
)

var (
	version    = "dev"     // Set during build
	buildDate  = "unknown" // Set during build
	commitHash = "HEAD"    // Set during build
)

var rootCmd = &cobra.Command{
	Use:   "nix-foundry",
	Short: "Nix environment management framework",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize configuration through service
		cfgSvc := configSvc.NewService()
		if err := cfgSvc.Load(); err != nil {
			return fmt.Errorf("configuration error: %w", err)
		}
		return nil
	},
}

func init() {
	flags.AddGlobalFlags(rootCmd)

	// Initialize service dependencies
	cfgSvc := configSvc.NewService()
	validator := validation.NewService()
	platformSvc := platform.NewService()

	envSvc := environmentSvc.NewService(
		cfgSvc.GetConfigDir(),
		cfgSvc,
		validator,
		platformSvc,
	)

	pkgSvc := packages.NewService(cfgSvc.GetConfigDir())
	projectService := projectSvc.NewService(cfgSvc, envSvc, pkgSvc)
	backupService := backupSvc.NewService(cfgSvc, envSvc, projectService)

	rootCmd.AddCommand(
		doctorCmd.NewDoctorCommand(projectService),
		projectCmd.NewProjectCommandGroup(projectService),
		backupCmd.NewCmd(backupService),
		uninstallCmd.NewUninstallCommand(),
		versionCmd,
		environmentCmd.NewCmd(envSvc),
		configCmd.NewConfigCommandGroup(cfgSvc),
		profileCmd.NewProfileCommandGroup(),
		statusCmd.NewStatusCommandGroup(),
	)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("nix-foundry %s\n", version)
		fmt.Printf("Build date: %s\n", buildDate)
		fmt.Printf("Commit: %s\n", commitHash)
	},
}
