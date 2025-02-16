package config

import (
	"fmt"
	"strings"
)

// PreviewConfiguration generates a human-readable preview of the configuration
func PreviewConfiguration(nixCfg *NixConfig) string {
	var preview strings.Builder

	preview.WriteString("\nConfiguration Preview:\n")
	preview.WriteString("--------------------\n")

	// Shell configuration
	preview.WriteString(fmt.Sprintf("Shell: %s\n", nixCfg.Shell.Type))
	if len(nixCfg.Shell.Plugins) > 0 {
		preview.WriteString("  Plugins:\n")
		for _, plugin := range nixCfg.Shell.Plugins {
			preview.WriteString(fmt.Sprintf("    - %s\n", plugin))
		}
	}

	// Editor configuration
	preview.WriteString(fmt.Sprintf("\nEditor: %s\n", nixCfg.Editor.Type))
	if len(nixCfg.Editor.Extensions) > 0 {
		preview.WriteString("  Extensions:\n")
		for _, ext := range nixCfg.Editor.Extensions {
			preview.WriteString(fmt.Sprintf("    - %s\n", ext))
		}
	}

	// Git configuration
	preview.WriteString("\nGit Configuration:\n")
	if nixCfg.Git.Enable {
		preview.WriteString(fmt.Sprintf("  User: %s\n", nixCfg.Git.User.Name))
		preview.WriteString(fmt.Sprintf("  Email: %s\n", nixCfg.Git.User.Email))
	} else {
		preview.WriteString("  Disabled\n")
	}

	// Packages
	preview.WriteString("\nPackages:\n")
	if len(nixCfg.Packages.Additional) > 0 {
		preview.WriteString("  Additional:\n")
		for _, pkg := range nixCfg.Packages.Additional {
			preview.WriteString(fmt.Sprintf("    - %s\n", pkg))
		}
	}
	if len(nixCfg.Packages.Development) > 0 {
		preview.WriteString("  Development:\n")
		for _, pkg := range nixCfg.Packages.Development {
			preview.WriteString(fmt.Sprintf("    - %s\n", pkg))
		}
	}
	if len(nixCfg.Packages.PlatformSpecific) > 0 {
		preview.WriteString("  Platform Specific:\n")
		for platform, pkgs := range nixCfg.Packages.PlatformSpecific {
			preview.WriteString(fmt.Sprintf("    %s:\n", platform))
			for _, pkg := range pkgs {
				preview.WriteString(fmt.Sprintf("      - %s\n", pkg))
			}
		}
	}

	return preview.String()
}
