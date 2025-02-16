package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/internal/services/packages"
	"github.com/shawnkhoffman/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func NewPackageCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "package",
		Short: "Manage nix-foundry packages",
		Long:  `Add, remove, and manage packages in your nix-foundry environment.`,
	}

	cmd.AddCommand(
		NewAddPackageCommand(),
		NewRemovePackageCommand(),
		NewListPackagesCommand(),
	)

	return cmd
}

func NewAddPackageCommand() *cobra.Command {
	var pkgType string

	cmd := &cobra.Command{
		Use:   "add [packages...]",
		Short: "Add packages to configuration",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir := getConfigDir()
			pkgSvc := packages.NewService(configDir)

			spin := progress.NewSpinner("Adding packages...")
			spin.Start()

			if err := pkgSvc.Add(args, pkgType); err != nil {
				spin.Fail("Failed to add packages")
				return fmt.Errorf("error adding packages: %w", err)
			}

			spin.Success(fmt.Sprintf("Added %d %s packages", len(args), pkgType))
			fmt.Println("\nℹ️  Run 'nix-foundry update' to apply changes")
			return nil
		},
	}

	cmd.Flags().StringVarP(&pkgType, "type", "t", "user", "Package type (core|user|team)")
	return cmd
}

func NewRemovePackageCommand() *cobra.Command {
	var pkgType string

	cmd := &cobra.Command{
		Use:   "remove [packages...]",
		Short: "Remove packages from configuration",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir := getConfigDir()
			pkgSvc := packages.NewService(configDir)

			spin := progress.NewSpinner("Removing packages...")
			spin.Start()

			if err := pkgSvc.Remove(args, pkgType); err != nil {
				spin.Fail("Failed to remove packages")
				return fmt.Errorf("error removing packages: %w", err)
			}

			spin.Success(fmt.Sprintf("Removed %d %s packages", len(args), pkgType))
			fmt.Println("\nℹ️  Run 'nix-foundry update' to apply changes")
			return nil
		},
	}

	cmd.Flags().StringVarP(&pkgType, "type", "t", "user", "Package type (core|user|team)")
	return cmd
}

func NewListPackagesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured packages",
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir := getConfigDir()
			pkgSvc := packages.NewService(configDir)

			pkgs, err := pkgSvc.List()
			if err != nil {
				return fmt.Errorf("failed to list packages: %w", err)
			}

			if len(pkgs) == 0 {
				fmt.Println("No packages configured")
				return nil
			}

			fmt.Println("Configured Packages:")
			for t, list := range pkgs {
				fmt.Printf("\n%s (%d):\n", cases.Title(language.English).String(t), len(list))
				fmt.Println("  " + strings.Join(list, "\n  "))
			}
			return nil
		},
	}
}

func getConfigDir() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "nix-foundry")
}
