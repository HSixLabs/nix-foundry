/*
Package packages provides package management functionality for Nix Foundry.
It handles package installation, removal, and validation across different platforms.
*/
package packages

import (
	"fmt"
	"os/exec"

	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/shawnkhoffman/nix-foundry/pkg/platform"
)

/*
Manager handles package management operations.
It provides functionality for installing, removing, and managing packages
using the Nix package manager.
*/
type Manager struct {
	fs     filesystem.FileSystem
	isWSL  bool
	groups map[string][]string
}

/*
NewManager creates a new package manager instance with the provided filesystem.
*/
func NewManager(fs filesystem.FileSystem) *Manager {
	return &Manager{
		fs:     fs,
		isWSL:  platform.IsWSL(),
		groups: make(map[string][]string),
	}
}

/*
InstallPackage installs a package using nix-env.
*/
func (m *Manager) InstallPackage(pkg string) error {
	cmd := exec.Command("bash", "-c", fmt.Sprintf(
		". /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh && "+
			"NIXPKGS_ALLOW_UNFREE=1 NIXPKGS_ALLOW_UNSUPPORTED_SYSTEM=1 "+
			"/nix/var/nix/profiles/default/bin/nix-env -iA nixpkgs.%s -Q",
		pkg))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install package %s: %s: %w", pkg, output, err)
	}
	return nil
}

/*
RemovePackage removes a package using nix-env.
*/
func (m *Manager) RemovePackage(pkg string) error {
	cmd := exec.Command("bash", "-c", fmt.Sprintf(
		". /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh && "+
			"/nix/var/nix/profiles/default/bin/nix-env -e %s",
		pkg))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to remove package %s: %s: %w", pkg, output, err)
	}
	return nil
}

/*
ListInstalledPackages returns a list of installed packages.
*/
func (m *Manager) ListInstalledPackages() ([]string, error) {
	cmd := exec.Command("bash", "-c",
		". /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh && "+
			"/nix/var/nix/profiles/default/bin/nix-env -q")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}
	return []string{string(output)}, nil
}

/*
SearchPackages searches for available packages matching the query.
Returns a map of package names to their descriptions.
*/
func (m *Manager) SearchPackages(query string) (map[string]string, error) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf(
		". /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh && "+
			"/nix/var/nix/profiles/default/bin/nix-env -qaP '*%s*'",
		query))
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to search packages: %w", err)
	}
	return map[string]string{string(output): ""}, nil
}

/*
GetPackageName returns the platform-specific package name for a given package.
For language packages, it returns both the language and its essential tools.
*/
func (m *Manager) GetPackageName(pkg string) string {
	platformSpecificNames := map[platform.Platform]map[string]string{
		platform.Linux: {
			"vscode": "code",
		},
	}

	languagePackages := map[string][]string{
		"python": {"python3", "python3-pip"},
		"node":   {"nodejs", "npm"},
		"go":     {"go", "gopls"},
		"java":   {"openjdk", "maven"},
	}

	if components, ok := languagePackages[pkg]; ok {
		return components[0]
	}

	if m.isWSL {
		if name, ok := platformSpecificNames[platform.Linux][pkg]; ok {
			return name
		}
	}

	return pkg
}

/*
GetDefaultPackages returns the default packages for the current platform.
These are packages that are absolutely required by nix-foundry.
*/
func (m *Manager) GetDefaultPackages() []string {
	return []string{}
}

/*
ValidatePackage checks if a package is valid for the current platform.
*/
func (m *Manager) ValidatePackage(pkg string) error {
	cmd := exec.Command("bash", "-c", fmt.Sprintf(
		". /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh && "+
			"/nix/var/nix/profiles/default/bin/nix-env -qaP '^%s$'",
		pkg))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("package %s not found", pkg)
	}
	return nil
}

/*
GetPackageGroups returns predefined groups of packages for common use cases.
*/
func (m *Manager) GetPackageGroups() map[string][]string {
	return map[string][]string{
		"development": {"git", "make", "gcc"},
		"web":         {"nodejs", "yarn", "nginx"},
		"data":        {"postgresql", "redis", "mongodb"},
	}
}

/*
GetPackageDescription returns a description of what a package does.
*/
func (m *Manager) GetPackageDescription(pkg string) string {
	descriptions := map[string]string{
		"python3": "Python interpreter with pip",
		"nodejs":  "Node.js runtime with npm",
		"go":      "Go compiler and tools",
		"openjdk": "Java Development Kit with Maven",
		"gcc":     "GNU Compiler Collection",
		"code":    "VS Code editor",
		"git":     "Git version control",
	}

	if desc, ok := descriptions[pkg]; ok {
		return desc
	}
	return fmt.Sprintf("Package %s", pkg)
}
