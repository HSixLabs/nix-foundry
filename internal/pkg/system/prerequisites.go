package system

import (
	"fmt"
	"os/exec"
)

func CheckPrerequisites(testMode bool) error {
	// Skip prerequisite checks in test mode
	if testMode {
		return nil
	}

	// Check for nix installation
	if _, err := exec.LookPath("nix"); err != nil {
		return fmt.Errorf("nix is not installed: %w", err)
	}

	// Check for home-manager
	if _, err := exec.LookPath("home-manager"); err != nil {
		return fmt.Errorf("home-manager is not installed: %w", err)
	}

	return nil
}
