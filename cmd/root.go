package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nix-foundry",
	Short: "A tool for managing Nix environments across platforms",
	Long: `nix-foundry is a cross-platform tool for managing Nix environments.
It provides a unified interface for installing, configuring, and managing
Nix packages and environments across macOS, Linux, and Windows Subsystem
for Linux (WSL).`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
}
