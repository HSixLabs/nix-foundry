package platform

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// InstallHomebrew installs Homebrew on macOS systems
func InstallHomebrew() error {
	if runtime.GOOS != "darwin" {
		return fmt.Errorf("homebrew installation is only supported on macOS")
	}

	// Install Homebrew using the official script
	cmd := exec.Command("/bin/bash", "-c",
		`"$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install homebrew: %w", err)
	}

	return nil
}

// InstallHomeManager installs home-manager using nix-env
func InstallHomeManager() error {
	// Add home-manager channel
	addChannel := exec.Command("nix-channel", "--add",
		"https://github.com/nix-community/home-manager/archive/master.tar.gz", "home-manager")
	if err := addChannel.Run(); err != nil {
		return fmt.Errorf("failed to add home-manager channel: %w", err)
	}

	// Update channels
	updateChannel := exec.Command("nix-channel", "--update")
	if err := updateChannel.Run(); err != nil {
		return fmt.Errorf("failed to update channels: %w", err)
	}

	// Install home-manager
	install := exec.Command("nix-env", "-iA", "nixpkgs.home-manager")
	if err := install.Run(); err != nil {
		return fmt.Errorf("failed to install home-manager: %w", err)
	}

	return nil
}
