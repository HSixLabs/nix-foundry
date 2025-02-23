// Package cmd provides the command-line interface for Nix Foundry.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/pkg/docs"
	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate documentation",
	Long: `Generate documentation for Nix Foundry.
This command generates various documentation files in Markdown format.`,
}

var docsInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Generate installation guide",
	Long: `Generate installation guide in Markdown format.
The guide includes prerequisites and step-by-step instructions.`,
	RunE: runDocsInstall,
}

var docsPackagesCmd = &cobra.Command{
	Use:   "packages",
	Short: "Generate package guide",
	Long: `Generate package management guide in Markdown format.
The guide includes information about package managers and usage.`,
	RunE: runDocsPackages,
}

var docsTroubleshootingCmd = &cobra.Command{
	Use:   "troubleshooting",
	Short: "Generate troubleshooting guide",
	Long: `Generate troubleshooting guide in Markdown format.
The guide includes common issues and their solutions.`,
	RunE: runDocsTroubleshooting,
}

var docsUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Generate uninstallation guide",
	Long: `Generate uninstallation guide in Markdown format.
The guide includes steps to remove Nix Foundry.`,
	RunE: runDocsUninstall,
}

var docsAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Generate all documentation",
	Long: `Generate all documentation files in Markdown format.
This includes installation, packages, troubleshooting, and uninstallation guides.`,
	RunE: runDocsAll,
}

func init() {
	rootCmd.AddCommand(docsCmd)
	docsCmd.AddCommand(docsInstallCmd)
	docsCmd.AddCommand(docsPackagesCmd)
	docsCmd.AddCommand(docsTroubleshootingCmd)
	docsCmd.AddCommand(docsUninstallCmd)
	docsCmd.AddCommand(docsAllCmd)
}

func runDocsInstall(cmd *cobra.Command, args []string) error {
	fs := filesystem.NewOSFileSystem()
	generator := docs.NewGenerator(fs)
	content, err := generator.GenerateInstallGuide()
	if err != nil {
		return fmt.Errorf("failed to generate installation guide: %w", err)
	}

	if err := os.WriteFile("docs/installation.md", []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write installation guide: %w", err)
	}

	fmt.Println("✨ Generated installation guide: docs/installation.md")
	return nil
}

func runDocsPackages(cmd *cobra.Command, args []string) error {
	fs := filesystem.NewOSFileSystem()
	generator := docs.NewGenerator(fs)
	content, err := generator.GeneratePackageGuide()
	if err != nil {
		return fmt.Errorf("failed to generate package guide: %w", err)
	}

	if err := os.WriteFile("docs/packages.md", []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write package guide: %w", err)
	}

	fmt.Println("✨ Generated package guide: docs/packages.md")
	return nil
}

func runDocsTroubleshooting(cmd *cobra.Command, args []string) error {
	fs := filesystem.NewOSFileSystem()
	generator := docs.NewGenerator(fs)
	content, err := generator.GenerateTroubleshootingGuide()
	if err != nil {
		return fmt.Errorf("failed to generate troubleshooting guide: %w", err)
	}

	if err := os.WriteFile("docs/troubleshooting.md", []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write troubleshooting guide: %w", err)
	}

	fmt.Println("✨ Generated troubleshooting guide: docs/troubleshooting.md")
	return nil
}

func runDocsUninstall(cmd *cobra.Command, args []string) error {
	fs := filesystem.NewOSFileSystem()
	generator := docs.NewGenerator(fs)
	content, err := generator.GenerateUninstallGuide()
	if err != nil {
		return fmt.Errorf("failed to generate uninstallation guide: %w", err)
	}

	if err := os.WriteFile("docs/uninstall.md", []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write uninstallation guide: %w", err)
	}

	fmt.Println("✨ Generated uninstallation guide: docs/uninstall.md")
	return nil
}

func runDocsAll(cmd *cobra.Command, args []string) error {
	fs := filesystem.NewOSFileSystem()
	generator := docs.NewGenerator(fs)

	// Create docs directory if it doesn't exist
	if err := os.MkdirAll("docs", 0755); err != nil {
		return fmt.Errorf("failed to create docs directory: %w", err)
	}

	// Generate all guides
	guides := map[string]func() (string, error){
		"installation.md":    generator.GenerateInstallGuide,
		"packages.md":        generator.GeneratePackageGuide,
		"troubleshooting.md": generator.GenerateTroubleshootingGuide,
		"uninstall.md":       generator.GenerateUninstallGuide,
	}

	for filename, genFunc := range guides {
		content, err := genFunc()
		if err != nil {
			return fmt.Errorf("failed to generate %s: %w", filename, err)
		}

		path := filepath.Join("docs", filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}

		fmt.Printf("✨ Generated %s\n", path)
	}

	return nil
}
