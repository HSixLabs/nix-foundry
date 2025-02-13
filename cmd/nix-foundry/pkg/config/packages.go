package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/platform"
)

type PackageSet struct {
	Common []string
	Darwin []string
	Linux  []string
	WSL    []string
	Custom []string // New field for user-defined packages
}

var defaultPackages = PackageSet{
	Common: []string{
		"git",
		"curl",
		"wget",
		"ripgrep",
		"fd",
		"jq",
		"tree",
	},
	Darwin: []string{
		"coreutils",
		"gnu-sed",
		"gnupg",
		"iterm2",
	},
	Linux: []string{
		"gnupg",
		"xclip",
		"htop",
	},
	WSL: []string{
		"wslu", // WSL utilities
		"wsl-clipboard",
	},
}

// Load custom packages from ~/.config/nix-foundry/packages.json
func loadCustomPackages() ([]string, error) {
	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return nil, homeErr
	}

	packagesFile := filepath.Join(homeDir, ".config/nix-foundry/packages.json")
	if _, err := os.Stat(packagesFile); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(packagesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read packages file: %w", err)
	}

	var packages struct {
		Packages []string `json:"packages"`
	}
	if err := json.Unmarshal(data, &packages); err != nil {
		return nil, fmt.Errorf("failed to parse packages file: %w", err)
	}

	return packages.Packages, nil
}

func GetPackagesForPlatform(sys *platform.System) []string {
	packages := make([]string, 0, len(defaultPackages.Common))
	packages = append(packages, defaultPackages.Common...)

	switch sys.OS {
	case "darwin":
		packages = append(packages, defaultPackages.Darwin...)
	case "linux":
		packages = append(packages, defaultPackages.Linux...)
		if sys.IsWSL {
			packages = append(packages, defaultPackages.WSL...)
		}
	}

	// Add custom packages if available
	if custom, err := loadCustomPackages(); err == nil && len(custom) > 0 {
		packages = append(packages, custom...)
	}

	return packages
}

// LoadCustomPackages loads the list of custom packages
func LoadCustomPackages() ([]string, error) {
	home, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return nil, homeErr
	}

	packagesFile := filepath.Join(home, ".config", "nix-foundry", "packages.json")
	if _, err := os.Stat(packagesFile); os.IsNotExist(err) {
		return []string{}, nil
	}

	data, err := os.ReadFile(packagesFile)
	if err != nil {
		return nil, err
	}

	var packages []string
	if err := json.Unmarshal(data, &packages); err != nil {
		return nil, err
	}

	return packages, nil
}

// SaveCustomPackages saves the list of custom packages
func SaveCustomPackages(packages []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(packages, "", "  ")
	if err != nil {
		return err
	}

	packagesFile := filepath.Join(home, ".config", "nix-foundry", "packages.json")
	return os.WriteFile(packagesFile, data, 0644)
}
