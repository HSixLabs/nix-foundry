package cmd

import (
	"github.com/spf13/cobra"
)

func NewProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage Nix projects",
		Long:  `Commands for managing Nix projects.`,
	}

	// Add subcommands here if needed
	return cmd
}
