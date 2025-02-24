/*
Package docs provides documentation generation functionality for Nix Foundry.
It generates various documentation files including installation guides,
package documentation, and troubleshooting guides.
*/
package docs

import (
	"sort"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/pkg/packages"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

/*
GenerateInstallGuide generates the installation guide.
It creates comprehensive documentation for installing Nix Foundry
on different platforms.
*/
func GenerateInstallGuide() (string, error) {
	var sb strings.Builder

	sb.WriteString("# Installation Guide\n\n")
	sb.WriteString("## Prerequisites\n\n")
	sb.WriteString("- Linux, macOS, or Windows (WSL2)\n")
	sb.WriteString("- Internet connection\n\n")
	sb.WriteString("## Installation Steps\n\n")
	sb.WriteString("1. Download the installer\n")
	sb.WriteString("2. Run the installation script\n")
	sb.WriteString("3. Follow the interactive setup\n")

	return sb.String(), nil
}

/*
GeneratePackageGuide generates the package management guide.
It documents available packages, their descriptions, and usage instructions.
*/
func GeneratePackageGuide() (string, error) {
	var sb strings.Builder

	sb.WriteString("# Package Management Guide\n\n")
	sb.WriteString("## Available Packages\n\n")
	sb.WriteString(GeneratePackageList())
	sb.WriteString("## Usage\n\n")
	sb.WriteString("Install packages using:\n")
	sb.WriteString("```bash\n")
	sb.WriteString("nix-foundry install <package>\n")
	sb.WriteString("```\n")

	return sb.String(), nil
}

/*
GenerateTroubleshootingGuide generates the troubleshooting guide.
It provides solutions for common issues and debugging steps.
*/
func GenerateTroubleshootingGuide() (string, error) {
	var sb strings.Builder

	sb.WriteString("# Troubleshooting Guide\n\n")
	sb.WriteString("## Common Issues\n\n")
	sb.WriteString("### Installation Fails\n")
	sb.WriteString("1. Check system requirements\n")
	sb.WriteString("2. Verify internet connection\n")
	sb.WriteString("3. Check file permissions\n")

	return sb.String(), nil
}

/*
GenerateUninstallGuide generates the uninstallation guide.
It provides instructions for removing Nix Foundry and its components.
*/
func GenerateUninstallGuide() (string, error) {
	var sb strings.Builder

	sb.WriteString("# Uninstallation Guide\n\n")
	sb.WriteString("## Steps\n\n")
	sb.WriteString("1. Run the uninstaller\n")
	sb.WriteString("2. Remove configuration files\n")
	sb.WriteString("3. Clean up package data\n")

	return sb.String(), nil
}

/*
GeneratePackageList generates a markdown list of available packages.
It organizes packages by category and includes descriptions.
*/
func GeneratePackageList() string {
	var sb strings.Builder
	pm := packages.NewManager(nil)
	groups := pm.GetPackageGroups()

	var groupNames []string
	for name := range groups {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	caser := cases.Title(language.English)
	for _, group := range groupNames {
		sb.WriteString("### " + caser.String(group) + "\n\n")
		for _, pkg := range groups[group] {
			desc := pm.GetPackageDescription(pkg)
			sb.WriteString("- `" + pkg + "`: " + desc + "\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

/*
GeneratePlatformList generates a markdown list of supported platforms.
It includes platform-specific information and requirements.
*/
func GeneratePlatformList() string {
	var sb strings.Builder

	sb.WriteString("# Supported Platforms\n\n")
	sb.WriteString("## Operating Systems\n\n")

	sb.WriteString("### Linux\n")
	sb.WriteString("- Most major distributions supported\n")
	sb.WriteString("- Multi-user installation available\n\n")

	sb.WriteString("### macOS\n")
	sb.WriteString("- Intel and Apple Silicon supported\n")
	sb.WriteString("- Homebrew not required\n\n")

	sb.WriteString("### Windows\n")
	sb.WriteString("- WSL2 required\n")
	sb.WriteString("- Ubuntu WSL recommended\n")

	return sb.String()
}
