package main

import (
	"fmt"

	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/config"
	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
)

var packagesCmd = &cobra.Command{
	Use:   "packages",
	Short: "Manage custom packages",
	Long:  `Add, remove, or list custom packages in your nix-foundry environment.`,
}

var listPackagesCmd = &cobra.Command{
	Use:   "list",
	Short: "List custom packages",
	RunE: func(cmd *cobra.Command, args []string) error {
		packages, err := config.LoadCustomPackages()
		if err != nil {
			return fmt.Errorf("failed to load custom packages: %w", err)
		}

		if len(packages) == 0 {
			fmt.Println("No custom packages configured")
			return nil
		}

		fmt.Println("Custom packages:")
		for i, pkg := range packages {
			fmt.Printf("%d. %s\n", i+1, pkg)
		}
		return nil
	},
}

var addPackageCmd = &cobra.Command{
	Use:   "add [package...]",
	Short: "Add custom packages",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("please specify at least one package to add")
		}

		spin := progress.NewSpinner("Adding packages...")
		spin.Start()

		// Load existing packages
		existing, _ := config.LoadCustomPackages()
		packages := make(map[string]bool)
		for _, pkg := range existing {
			packages[pkg] = true
		}

		// Add new packages
		for _, pkg := range args {
			packages[pkg] = true
		}

		// Convert back to slice
		var packageList []string
		for pkg := range packages {
			packageList = append(packageList, pkg)
		}

		if err := config.SaveCustomPackages(packageList); err != nil {
			spin.Fail("Failed to add packages")
			return err
		}

		spin.Success("Packages added")
		fmt.Println("\nℹ️  Run 'nix-foundry update' to apply changes")
		return nil
	},
}

var removePackageCmd = &cobra.Command{
	Use:   "remove [package...]",
	Short: "Remove custom packages",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("please specify at least one package to remove")
		}

		spin := progress.NewSpinner("Removing packages...")
		spin.Start()

		// Load existing packages
		existing, err := config.LoadCustomPackages()
		if err != nil {
			spin.Fail("Failed to load existing packages")
			return err
		}

		// Remove specified packages
		toRemove := make(map[string]bool)
		for _, pkg := range args {
			toRemove[pkg] = true
		}

		var remaining []string
		for _, pkg := range existing {
			if !toRemove[pkg] {
				remaining = append(remaining, pkg)
			}
		}

		if err := config.SaveCustomPackages(remaining); err != nil {
			spin.Fail("Failed to remove packages")
			return err
		}

		spin.Success("Packages removed")
		fmt.Println("\nℹ️  Run 'nix-foundry update' to apply changes")
		return nil
	},
}

func init() {
	packagesCmd.AddCommand(listPackagesCmd)
	packagesCmd.AddCommand(addPackageCmd)
	packagesCmd.AddCommand(removePackageCmd)
}
