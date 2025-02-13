package nix

import (
	"fmt"
	"os"
	"os/exec"
)

func IsInstalled() bool {
	_, err := exec.LookPath("nix")
	return err == nil
}

func Install() error {
	if IsInstalled() {
		return nil
	}

	fmt.Println("Installing Nix...")
	cmd := exec.Command("sh", "-c",
		`curl -L https://nixos.org/nix/install | sh -s -- --daemon`)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("installation failed: %w\nPlease ensure curl is installed and you have internet access", err)
	}

	// Verify installation
	if !IsInstalled() {
		return fmt.Errorf("nix was not installed correctly. Please try installing manually")
	}

	return nil
}
