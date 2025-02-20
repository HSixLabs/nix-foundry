package cmd

import (
	"github.com/spf13/cobra"
)

func NewPackagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "packages",
		Short: "Manage Nix packages",
		Long:  `Commands for managing Nix packages.`,
	}

	return cmd
}
