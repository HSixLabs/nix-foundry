// Package docs provides documentation generation functionality.
package docs

import (
	"fmt"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/shawnkhoffman/nix-foundry/pkg/packages"
)

// Generator handles documentation generation.
type Generator struct {
	fs filesystem.FileSystem
}

// NewGenerator creates a new documentation generator.
func NewGenerator(fs filesystem.FileSystem) *Generator {
	return &Generator{fs: fs}
}

// GenerateInstallGuide generates the installation guide.
func (g *Generator) GenerateInstallGuide() (string, error) {
	var sb strings.Builder
	sb.WriteString("# Installation Guide\n\n")
	sb.WriteString("## Prerequisites\n\n")
	sb.WriteString("- Nix package manager installed\n")
	sb.WriteString("- Git (optional, for development)\n\n")
	sb.WriteString("## Installation Steps\n\n")
	sb.WriteString("1. Install Nix Foundry:\n")
	sb.WriteString("```bash\n")
	sb.WriteString("nix-env -i nix-foundry\n")
	sb.WriteString("```\n\n")
	sb.WriteString("2. Initialize configuration:\n")
	sb.WriteString("```bash\n")
	sb.WriteString("nix-foundry config init\n")
	sb.WriteString("```\n\n")
	return sb.String(), nil
}

// GeneratePackageGuide generates the package management guide.
func (g *Generator) GeneratePackageGuide() (string, error) {
	var sb strings.Builder
	sb.WriteString("# Package Management Guide\n\n")
	sb.WriteString("## Package Manager\n\n")
	sb.WriteString("Nix Foundry uses `nix-env` for package management.\n\n")
	sb.WriteString("## Managing Packages\n\n")
	sb.WriteString("### Installing Packages\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("nix-foundry packages install git curl wget\n")
	sb.WriteString("```\n\n")
	return sb.String(), nil
}

// GenerateTroubleshootingGuide generates the troubleshooting guide.
func (g *Generator) GenerateTroubleshootingGuide() (string, error) {
	var sb strings.Builder
	sb.WriteString("# Troubleshooting Guide\n\n")
	sb.WriteString("## Common Issues\n\n")
	sb.WriteString("### Package Installation Fails\n\n")
	sb.WriteString("1. Check package availability:\n")
	sb.WriteString("```bash\n")
	sb.WriteString("nix-foundry packages search <package>\n")
	sb.WriteString("```\n\n")
	return sb.String(), nil
}

// GenerateUninstallGuide generates the uninstallation guide.
func (g *Generator) GenerateUninstallGuide() (string, error) {
	var sb strings.Builder
	sb.WriteString("# Uninstallation Guide\n\n")
	sb.WriteString("## Steps\n\n")
	sb.WriteString("1. Remove configuration:\n")
	sb.WriteString("```bash\n")
	sb.WriteString("nix-foundry config uninstall\n")
	sb.WriteString("```\n\n")
	sb.WriteString("2. Uninstall package:\n")
	sb.WriteString("```bash\n")
	sb.WriteString("nix-env -e nix-foundry\n")
	sb.WriteString("```\n\n")
	return sb.String(), nil
}

// GeneratePackageList generates a list of available packages.
func (g *Generator) GeneratePackageList() (string, error) {
	manager := packages.NewManager(g.fs)
	groups := manager.GetPackageGroups()

	var sb strings.Builder
	sb.WriteString("# Available Packages\n\n")

	for group, pkgs := range groups {
		sb.WriteString(fmt.Sprintf("## %s\n\n", strings.Title(group)))
		for _, pkg := range pkgs {
			desc := manager.GetPackageDescription(pkg)
			sb.WriteString(fmt.Sprintf("- `%s`: %s\n", pkg, desc))
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// GeneratePackageTable generates a table of available packages.
func (g *Generator) GeneratePackageTable() (string, error) {
	manager := packages.NewManager(g.fs)
	groups := manager.GetPackageGroups()

	var sb strings.Builder
	sb.WriteString("# Package Reference\n\n")
	sb.WriteString("| Package | Group | Description |\n")
	sb.WriteString("|---------|--------|-------------|\n")

	for group, pkgs := range groups {
		for _, pkg := range pkgs {
			desc := manager.GetPackageDescription(pkg)
			sb.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", pkg, group, desc))
		}
	}

	return sb.String(), nil
}
