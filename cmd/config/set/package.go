/*
Package set provides commands for setting various configuration options in Nix Foundry.
This package specifically handles the configuration of packages, scripts,
and other settable properties within the Nix Foundry configuration system.
*/
package set

import (
	"fmt"

	"github.com/spf13/cobra"
)

/*
packageCmd represents the package management command for Nix Foundry configuration.
It serves as a parent command for subcommands that handle specific package
management operations such as adding, removing, or updating packages in the
configuration.
*/
var packageCmd = &cobra.Command{
	Use:   "package",
	Short: "Manage Nix packages in configuration",
	Long: `Manage Nix packages in configuration.
This command provides subcommands for managing Nix packages in your configuration.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return fmt.Errorf("please specify a subcommand")
	},
}

/*
init initializes the package command by adding it to the parent command.
This function is automatically called during package initialization.
*/
func init() {
	Cmd.AddCommand(packageCmd)
}
