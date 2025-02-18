package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
)

var RootCmd = &cobra.Command{
	Use:   "nix-foundry",
	Short: "A brief description of your application",
	Long:  "A longer description of your application",
	Run: func(cmd *cobra.Command, args []string) {
		// Implementation of the command
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		debug, _ := cmd.Flags().GetBool("debug")

		// Initialize logger FIRST
		if err := logging.InitLogger(debug); err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}

		// Then load configuration
		return initConfig()
	},
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initConfig() error {
	cfgSvc := config.NewService()
	_, err := cfgSvc.Load() // Add underscore to ignore the config return value
	return err
}
