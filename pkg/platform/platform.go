// Package platform provides platform-specific functionality and detection for different operating systems.
package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Platform represents a supported operating system platform.
type Platform string

const (
	Linux   Platform = "linux"
	MacOS   Platform = "darwin"
	Windows Platform = "windows"
)

// IsWSL determines if the current environment is running under Windows Subsystem for Linux.
func IsWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return false
	}

	return strings.Contains(strings.ToLower(string(data)), "microsoft")
}

// GetPlatform returns the current operating system platform.
func GetPlatform() Platform {
	switch runtime.GOOS {
	case "darwin":
		return MacOS
	case "windows":
		return Windows
	default:
		return Linux
	}
}

// IsMultiUserNixSupported checks if the current platform supports multi-user Nix installation.
func IsMultiUserNixSupported() bool {
	return !IsWSL() && runtime.GOOS != "windows"
}

// GetDefaultShell returns the default shell for the current platform.
func GetDefaultShell() string {
	if runtime.GOOS == "darwin" {
		return "/bin/zsh"
	}
	if runtime.GOOS == "windows" {
		if IsWSL() {
			return "/bin/bash"
		}
		return "powershell.exe"
	}
	return "/bin/bash"
}

// GetHomeDir returns the user's home directory with proper platform-specific path handling.
func GetHomeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if IsWSL() {
		homeDir = strings.Replace(homeDir, "/mnt/c/Users", "/home", 1)
	}

	return homeDir, nil
}

// GetConfigDir returns the configuration directory for Nix Foundry.
func GetConfigDir() (string, error) {
	homeDir, err := GetHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".config", "nix-foundry"), nil
}

// GetNixConfigDir returns the appropriate Nix configuration directory for the current platform.
func GetNixConfigDir() (string, error) {
	if runtime.GOOS == "windows" && !IsWSL() {
		homeDir, err := GetHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, ".nix-config"), nil
	}
	return "/etc/nix", nil
}

// GetNixProfileDir returns the appropriate Nix profile directory.
func GetNixProfileDir() (string, error) {
	homeDir, err := GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".nix-profile"), nil
}

// GetShellConfigFile returns the appropriate shell configuration file path for the given shell.
func GetShellConfigFile(shell string) (string, error) {
	homeDir, err := GetHomeDir()
	if err != nil {
		return "", err
	}

	switch shell {
	case "zsh":
		return filepath.Join(homeDir, ".zshrc"), nil
	case "bash":
		if runtime.GOOS == "darwin" {
			return filepath.Join(homeDir, ".bash_profile"), nil
		}
		return filepath.Join(homeDir, ".bashrc"), nil
	case "fish":
		return filepath.Join(homeDir, ".config", "fish", "config.fish"), nil
	default:
		return filepath.Join(homeDir, ".profile"), nil
	}
}
