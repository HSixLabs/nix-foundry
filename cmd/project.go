package cmd

import (
	"github.com/spf13/cobra"
)

// NewProjectCmd creates a new project command for Nix Foundry.
func NewProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage Nix projects",
		Long:  `Commands for managing Nix projects.`,
	}

	return cmd
}
