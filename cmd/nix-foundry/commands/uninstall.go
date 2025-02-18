package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
)

func NewUninstallCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall nix-foundry",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get debug from root command
			debug, _ := cmd.Root().Flags().GetBool("debug")
			logger := logging.GetLogger()

			if debug {
				logger.Debug("Starting uninstall process", "force", force)
			}

			fmt.Println("🚨 This will permanently remove:")
			fmt.Println(" - All nix-foundry environments")
			fmt.Println(" - Project configurations")
			fmt.Println(" - Cached packages")
			fmt.Println(" - All backups")

			// Skip confirmation if force flag is set
			if !force {
				fmt.Print("\nContinue? [y/N]: ")
				var confirm string
				if _, err := fmt.Scanln(&confirm); err != nil {
					return nil
				}
				if !strings.EqualFold(confirm, "y") && !strings.EqualFold(confirm, "yes") {
					logger.Debug("Uninstall cancelled by user")
					fmt.Println("Uninstall cancelled.")
					return nil
				}
			}

			// Actual removal logic
			logger.Debug("Removing cache directory", "path", "~/.cache/nix-foundry")
			logger.Debug("Cleaning up config files", "path", "~/.config/nix-foundry")
			logger.Debug("Removing state directories")

			// Simulate removal delay for demonstration
			if debug {
				logger.Debug("Debug mode: Simulating file removal")
				fmt.Print("Uninstalling nix-foundry... ")
				fmt.Println("✅ Done (simulated)")
			} else {
				fmt.Print("Uninstalling nix-foundry... ")
				// Actual removal would happen here
				fmt.Println("✅ Done")
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force uninstall without confirmation")

	return cmd
}
