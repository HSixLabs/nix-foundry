package main

import (
	"fmt"
	"os"

	"github.com/shawnkhoffman/nix-foundry/cmd"
	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/shawnkhoffman/nix-foundry/service/config"
	"github.com/spf13/cobra"
)

const (
	setupRequiredMsg = "🚫 Initial setup is required to use Nix Foundry."
	setupSuccessMsg  = "\n✅ Setup completed successfully! You can now use all Nix Foundry features."
)

func main() {
	configService := config.NewConfigService(filesystem.NewOSFileSystem())

	// Check for first run, but don't intercept help or version commands
	if !configService.ConfigExists() && !isHelpCommand(os.Args[1:]) && !isSetupCommand(os.Args[1:]) {
		// Ask user if they want to run setup
		runSetup, err := configService.PromptForSetup()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if runSetup {
			fmt.Println("\nStarting initial setup...")
			setupCmd := getSetupCommand()
			if err := setupCmd.Execute(); err != nil {
				fmt.Printf("Setup failed: %v\n", err)
				os.Exit(1)
			}
			if configService.ConfigExists() {
				fmt.Println(setupSuccessMsg)
				fmt.Println("\nYou can run this setup again later with: \033[36mnix-foundry config setup\033[0m")
				fmt.Println("To apply the configuration, run: \033[36mnix-foundry apply\033[0m")
				fmt.Println("\nTry \033[36mnix-foundry --help\033[0m to see available commands.")
			} else {
				fmt.Println(setupRequiredMsg)
				os.Exit(1)
			}
			return
		} else {
			fmt.Println("\nExiting...")
			fmt.Println(setupRequiredMsg)
			os.Exit(1)
		}
	}

	// Normal command execution
	rootCmd := createRootCommand()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func isHelpCommand(args []string) bool {
	return len(args) > 0 && (args[0] == "help" || args[0] == "--help" || args[0] == "-h")
}

func isSetupCommand(args []string) bool {
	return len(args) >= 2 && args[0] == "config" && args[1] == "setup"
}

func getSetupCommand() *cobra.Command {
	root := &cobra.Command{Use: "nix-foundry"}
	root.AddCommand(cmd.NewConfigCmd())
	root.SetArgs([]string{"config", "setup"})
	return root
}

func createRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "nix-foundry",
		Short: "Nix configuration management tool",
		Long:  `Nix Foundry - Powerful CLI for Nix configuration management.`,
	}
	rootCmd.AddCommand(
		cmd.NewConfigCmd(),
		// packages.NewPackagesCmd(),
		// project.NewProjectCmd(),
	)
	return rootCmd
}
