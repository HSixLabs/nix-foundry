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
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	backupSvc "github.com/shawnkhoffman/nix-foundry/internal/services/backup"
	configSvc "github.com/shawnkhoffman/nix-foundry/internal/services/config"
	environmentSvc "github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
	projectSvc "github.com/shawnkhoffman/nix-foundry/internal/services/project"
	"github.com/spf13/cobra"
)

var (
	version    = "dev"     // Set during build
	buildDate  = "unknown" // Set during build
	commitHash = "HEAD"    // Set during build
)

var RootCmd = &cobra.Command{
	Use:   "nix-foundry",
	Short: "Nix environment management framework",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Get debug flag value from PERSISTENT flag
		debug, _ := cmd.Flags().GetBool("debug")

		// Initialize logger once
		if err := logging.InitLogger(debug); err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}

		// Then load configuration
		cfgSvc := configSvc.NewService()
		if _, err := cfgSvc.Load(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	// Add global flags including debug
	flags.AddGlobalFlags(RootCmd)

	// Initialize service dependencies
	cfgSvc := configSvc.NewService()
	platformSvc := platform.NewService()

	envSvc := environmentSvc.NewService(
		cfgSvc.GetConfigDir(),
		cfgSvc,
		platformSvc,
		false, // testMode flag (default to false for main command)
		true,  // isolationEnabled (default to true)
		true,  // autoInstall (default to true)
	)

	pkgSvc := packages.NewService(cfgSvc.GetConfigDir())
	projectService := projectSvc.NewService(cfgSvc, envSvc, pkgSvc)

	// Use the helper function for interface conversion
	backupService := backupSvc.NewService(cfgSvc, envSvc, asService(projectService))

	RootCmd.AddCommand(
		doctorCmd.NewDoctorCommand(asService(projectService)),
		projectCmd.NewProjectCommandGroup(asService(projectService)),
		backupCmd.NewCmd(backupService),
		uninstallCmd.NewUninstallCommand(),
		versionCmd,
		environmentCmd.NewCmd(envSvc),
		configCmd.NewConfigCommand(cfgSvc, asService(projectService)),
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

// Add helper function for interface conversion
func asService(ps project.ProjectService) project.Service {
	return ps.(project.Service)
}
