// Package config provides configuration management functionality.
package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/shawnkhoffman/nix-foundry/pkg/schema"
)

// ApplyService handles configuration application operations.
type ApplyService struct {
	fs filesystem.FileSystem
}

// NewApplyService creates a new configuration apply service.
func NewApplyService(fs filesystem.FileSystem) *ApplyService {
	return &ApplyService{fs: fs}
}

// Apply applies the configuration by installing packages and running scripts.
func (s *ApplyService) Apply(config *schema.Config) error {
	if err := s.applyPackages(config); err != nil {
		return fmt.Errorf("failed to apply packages: %w", err)
	}

	if err := s.applyScripts(config); err != nil {
		return fmt.Errorf("failed to apply scripts: %w", err)
	}

	return nil
}

// applyPackages installs packages specified in the configuration.
func (s *ApplyService) applyPackages(config *schema.Config) error {
	allPackages := append(config.Nix.Packages.Core, config.Nix.Packages.Optional...)

	if len(allPackages) == 0 {
		fmt.Println("No packages to install")
		return nil
	}

	// Validate packages for macOS compatibility
	if err := s.validatePackagesForPlatform(allPackages); err != nil {
		return fmt.Errorf("package validation failed: %w", err)
	}

	fmt.Printf("Installing %d packages...\n", len(allPackages))

	for _, pkg := range allPackages {
		fmt.Printf("Installing %s...\n", pkg)

		cmd := exec.Command("bash", "-c", fmt.Sprintf(
			". /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh && "+
				"NIXPKGS_ALLOW_UNFREE=1 NIXPKGS_ALLOW_UNSUPPORTED_SYSTEM=1 "+
				"/nix/var/nix/profiles/default/bin/nix-env -iA nixpkgs.%s -Q",
			pkg))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install package %s: %w. "+
				"This may be due to macOS System Integrity Protection. "+
				"Try installing manually with: nix-env -iA nixpkgs.%s", pkg, err, pkg)
		}
	}

	fmt.Printf("✨ %d packages installed successfully\n", len(allPackages))
	return nil
}

// validatePackagesForPlatform validates packages for platform-specific compatibility
func (s *ApplyService) validatePackagesForPlatform(packages []string) error {
	problematicPackages := []string{
		"vscode", "code", "visual-studio-code",
		"intellij-idea-ultimate", "intellij-idea-community",
		"sublime-text", "atom", "android-studio",
	}

	var foundProblematic []string
	for _, pkg := range packages {
		for _, problematic := range problematicPackages {
			if pkg == problematic {
				foundProblematic = append(foundProblematic, pkg)
			}
		}
	}

	if len(foundProblematic) > 0 {
		fmt.Printf("⚠️  Warning: The following packages may have installation issues on macOS:\n")
		for _, pkg := range foundProblematic {
			fmt.Printf("   • %s (GUI application with potential SIP conflicts)\n", pkg)
		}
		fmt.Printf("\nNote: If installation fails, you can install these manually using:\n")
		fmt.Printf("  brew install --cask <package-name>\n")
		fmt.Printf("Or remove them from your configuration and continue with other packages.\n\n")
	}

	return nil
}

// applyScripts runs scripts specified in the configuration.
func (s *ApplyService) applyScripts(config *schema.Config) error {
	if len(config.Nix.Scripts) == 0 {
		fmt.Println("No scripts to run")
		return nil
	}

	fmt.Printf("Running %d scripts...\n", len(config.Nix.Scripts))

	for _, script := range config.Nix.Scripts {
		fmt.Printf("Running script: %s\n", script.Name)

		tmpDir, err := os.MkdirTemp("", "nix-foundry-script-*")
		if err != nil {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		scriptPath := filepath.Join(tmpDir, "script.sh")

		// Add Nix environment sourcing to script
		scriptContent := fmt.Sprintf(`#!/bin/bash
set -e

# Source Nix environment
if [ -e '/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh' ]; then
    . '/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh'
elif [ -e "$HOME/.nix-profile/etc/profile.d/nix.sh" ]; then
    . "$HOME/.nix-profile/etc/profile.d/nix.sh"
fi

# Allow unfree packages
export NIXPKGS_ALLOW_UNFREE=1
export NIXPKGS_ALLOW_UNSUPPORTED_SYSTEM=1

# Run user script
%s
`, script.Commands)

		if err := s.fs.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
			return fmt.Errorf("failed to write script file: %w", err)
		}

		cmd := exec.Command("bash", scriptPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run script '%s': %w", script.Name, err)
		}
	}

	fmt.Printf("✨ %d scripts completed successfully\n", len(config.Nix.Scripts))
	return nil
}
