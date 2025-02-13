package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	DefaultEnv = "default"
	CurrentEnv = "current"
)

func Apply(configDir string) error {
	// Enable flakes if not already enabled
	if err := enableFlakes(); err != nil {
		return fmt.Errorf("failed to enable flakes: %w\nTry adding 'experimental-features = nix-command flakes' to ~/.config/nix/nix.conf manually", err)
	}

	// Check if home-manager is available
	if _, err := exec.LookPath("home-manager"); err != nil {
		return fmt.Errorf("home-manager not found: please install it first using 'nix-env -iA nixpkgs.home-manager'")
	}

	// Get the active environment directory
	envDir, err := getActiveEnvironment(configDir)
	if err != nil {
		return fmt.Errorf("failed to get active environment: %w", err)
	}

	// Get absolute path to environment directory
	absEnvDir, err := filepath.Abs(envDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Verify flake.nix exists in the environment directory
	flakePath := filepath.Join(absEnvDir, "flake.nix")
	if _, err := os.Stat(flakePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("flake.nix not found in environment directory %s", absEnvDir)
		}
		return fmt.Errorf("failed to check flake.nix: %w", err)
	}

	// Initialize the command
	cmd := exec.Command("home-manager", "switch", "--flake", ".#default")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = absEnvDir // Set working directory to where flake.nix is located

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to apply configuration: %w\nTry running 'home-manager switch --flake %s#default' manually", err, absEnvDir)
	}

	return nil
}

// getActiveEnvironment returns the path to the active environment
func getActiveEnvironment(configDir string) (string, error) {
	// First check for a current symlink
	currentEnv := filepath.Join(configDir, "environments", CurrentEnv)
	if target, err := os.Readlink(currentEnv); err == nil {
		// Return the absolute path of the target
		if filepath.IsAbs(target) {
			return target, nil
		}
		// Convert relative symlink target to absolute path
		return filepath.Join(filepath.Dir(currentEnv), target), nil
	}

	// Fall back to default environment
	defaultEnv := filepath.Join(configDir, "environments", DefaultEnv)
	if _, err := os.Stat(defaultEnv); err != nil {
		return "", fmt.Errorf("no active environment found: %w", err)
	}

	// Return the absolute path
	absPath, err := filepath.Abs(defaultEnv)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	return absPath, nil
}

func enableFlakes() error {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "nix")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(
		filepath.Join(configDir, "nix.conf"),
		[]byte("experimental-features = nix-command flakes"),
		0644,
	)
}
