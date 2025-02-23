// Package cmd provides the command-line interface for Nix Foundry.
package cmd

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/shawnkhoffman/nix-foundry/pkg/packages"
	"github.com/spf13/cobra"
)

var packagesCmd = &cobra.Command{
	Use:   "packages",
	Short: "Manage Nix packages",
	Long: `Manage Nix packages with platform-specific optimizations.
Supports installing, removing, listing, and searching packages.`,
}

var installPackageCmd = &cobra.Command{
	Use:   "install [package...]",
	Short: "Install one or more packages",
	Long: `Install one or more Nix packages with platform-specific optimizations.
Package names are automatically adjusted for the current platform.

Examples:
  # Install a single package
  nix-foundry packages install git

  # Install multiple packages
  nix-foundry packages install git curl wget

  # Install from a specific group
  nix-foundry packages install --group development`,
	RunE: runInstallPackages,
}

var removePackageCmd = &cobra.Command{
	Use:   "remove [package...]",
	Short: "Remove one or more packages",
	Long: `Remove one or more installed Nix packages.

Examples:
  # Remove a single package
  nix-foundry packages remove git

  # Remove multiple packages
  nix-foundry packages remove git curl wget`,
	RunE: runRemovePackages,
}

var listPackagesCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed packages",
	Long: `List all installed Nix packages.
Optionally filter by group or show package descriptions.

Examples:
  # List all installed packages
  nix-foundry packages list

  # List packages with descriptions
  nix-foundry packages list --describe

  # List packages from a specific group
  nix-foundry packages list --group development`,
	RunE: runListPackages,
}

var searchPackagesCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search available packages",
	Long: `Search available Nix packages by name or description.
Results are filtered to show only packages available on your platform.

Examples:
  # Search for a package
  nix-foundry packages search git

  # Search with descriptions
  nix-foundry packages search --describe git`,
	RunE: runSearchPackages,
}

var (
	group    string
	describe bool
)

func init() {
	rootCmd.AddCommand(packagesCmd)
	packagesCmd.AddCommand(installPackageCmd)
	packagesCmd.AddCommand(removePackageCmd)
	packagesCmd.AddCommand(listPackagesCmd)
	packagesCmd.AddCommand(searchPackagesCmd)

	installPackageCmd.Flags().StringVar(&group, "group", "", "Install packages from predefined group")
	listPackagesCmd.Flags().StringVar(&group, "group", "", "List packages from specific group")
	listPackagesCmd.Flags().BoolVar(&describe, "describe", false, "Show package descriptions")
	searchPackagesCmd.Flags().BoolVar(&describe, "describe", false, "Show package descriptions")
}

func runInstallPackages(cmd *cobra.Command, args []string) error {
	if len(args) == 0 && group == "" {
		return fmt.Errorf("at least one package name or group is required")
	}

	manager := packages.NewManager(filesystem.NewOSFileSystem())

	var pkgs []string
	if group != "" {
		groups := manager.GetPackageGroups()
		if groupPkgs, ok := groups[group]; ok {
			pkgs = groupPkgs
		} else {
			return fmt.Errorf("unknown package group: %s", group)
		}
	} else {
		pkgs = args
	}

	for _, pkg := range pkgs {
		if err := manager.ValidatePackage(pkg); err != nil {
			return err
		}
		platformPkg := manager.GetPackageName(pkg)
		fmt.Printf("Installing %s...\n", platformPkg)
		if err := manager.InstallPackage(pkg); err != nil {
			return err
		}
	}

	fmt.Println("✨ Package installation completed successfully!")
	return nil
}

func runRemovePackages(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("at least one package name is required")
	}

	manager := packages.NewManager(filesystem.NewOSFileSystem())
	for _, pkg := range args {
		platformPkg := manager.GetPackageName(pkg)
		fmt.Printf("Removing %s...\n", platformPkg)
		if err := manager.RemovePackage(pkg); err != nil {
			return err
		}
	}

	fmt.Println("✨ Package removal completed successfully!")
	return nil
}

func runListPackages(cmd *cobra.Command, args []string) error {
	manager := packages.NewManager(filesystem.NewOSFileSystem())

	var pkgs []string
	var err error

	if group != "" {
		groups := manager.GetPackageGroups()
		if groupPkgs, ok := groups[group]; ok {
			pkgs = groupPkgs
		} else {
			return fmt.Errorf("unknown package group: %s", group)
		}
	} else {
		pkgs, err = manager.ListInstalledPackages()
		if err != nil {
			return fmt.Errorf("failed to list installed packages: %w", err)
		}
	}

	fmt.Println("Installed packages:")
	for _, pkg := range pkgs {
		if describe {
			desc := manager.GetPackageDescription(pkg)
			fmt.Printf("  %s: %s\n", pkg, desc)
		} else {
			fmt.Printf("  %s\n", pkg)
		}
	}

	return nil
}

func runSearchPackages(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("search query is required")
	}

	manager := packages.NewManager(filesystem.NewOSFileSystem())
	query := args[0]

	results, err := manager.SearchPackages(query)
	if err != nil {
		return fmt.Errorf("failed to search packages: %w", err)
	}

	fmt.Printf("Search results for '%s':\n", query)
	for pkg, desc := range results {
		if describe {
			fmt.Printf("  %s: %s\n", pkg, desc)
		} else {
			fmt.Printf("  %s\n", pkg)
		}
	}

	return nil
}
