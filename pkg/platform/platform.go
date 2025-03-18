/*
Package platform provides platform-specific functionality and detection for different operating systems.
It handles platform detection, path resolution, and system-specific configurations.
*/
package platform

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

/*
Platform represents a supported operating system platform.
*/
type Platform string

const (
	Linux   Platform = "linux"
	MacOS   Platform = "darwin"
	Windows Platform = "windows"
)

/*
IsWSL determines if the current environment is running under Windows Subsystem for Linux.
It checks the system version information for Microsoft-specific identifiers.
*/
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

/*
GetPlatform returns the current operating system platform.
*/
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

/*
IsMultiUserNixSupported checks if the current platform supports multi-user Nix installation.
*/
func IsMultiUserNixSupported() bool {
	return !IsWSL() && runtime.GOOS != "windows"
}

/*
GetDefaultShell returns the default shell for the current platform.
*/
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

/*
GetHomeDir returns the user's home directory with proper platform-specific path handling.
For WSL environments, it adjusts Windows paths to their Linux equivalents.
*/
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

/*
GetConfigDir returns the configuration directory for Nix Foundry.
*/
func GetConfigDir() (string, error) {
	homeDir, err := GetHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".config", "nix-foundry"), nil
}

/*
GetNixSystem returns the Nix system identifier for the current platform.
It determines the appropriate system identifier based on the operating system
and CPU architecture.
*/
func GetNixSystem() string {
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "aarch64-darwin"
		}
		return "x86_64-darwin"
	case "linux":
		if runtime.GOARCH == "arm64" {
			return "aarch64-linux"
		}
		return "x86_64-linux"
	default:
		return "x86_64-linux"
	}
}

/*
GetNixConfigDir returns the appropriate Nix configuration directory for the current platform.
*/
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

/*
GetNixProfileDir returns the appropriate Nix profile directory.
*/
func GetNixProfileDir() (string, error) {
	homeDir, err := GetHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".nix-profile"), nil
}

/*
GetShellConfigFile returns the path to the configuration file for the specified shell.
It uses the real user's home directory when running under sudo.
*/
func GetShellConfigFile(shell string) (string, error) {
	homeDir, err := GetRealUserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	var rcFile string
	switch shell {
	case "bash":
		rcFile = filepath.Join(homeDir, ".bashrc")
	case "zsh":
		rcFile = filepath.Join(homeDir, ".zshrc")
	case "fish":
		rcFile = filepath.Join(homeDir, ".config", "fish", "config.fish")
	default:
		return "", fmt.Errorf("unsupported shell: %s", shell)
	}

	return rcFile, nil
}

/*
GetRealUserHomeDir returns the home directory of the real user, even when running under sudo.
It first checks SUDO_USER environment variable and falls back to the current user's home directory.
*/
func GetRealUserHomeDir() (string, error) {
	if os.Getenv("SUDO_USER") == "" {
		return GetHomeDir()
	}

	username := os.Getenv("SUDO_USER")

	if runtime.GOOS == "linux" {
		if IsWSL() {
			return filepath.Join("/home", username), nil
		}
		return filepath.Join("/home", username), nil
	}

	if runtime.GOOS == "darwin" {
		return filepath.Join("/Users", username), nil
	}

	return GetHomeDir()
}

/*
GetRealUser returns the UID and GID of the real user, even when running under sudo.
*/
func GetRealUser() (uid, gid int, err error) {
	uid = os.Getuid()
	gid = os.Getgid()

	if sudoUID := os.Getenv("SUDO_UID"); sudoUID != "" {
		if parsedUID, parseErr := strconv.Atoi(sudoUID); parseErr == nil {
			uid = parsedUID
		}
	}

	if sudoGID := os.Getenv("SUDO_GID"); sudoGID != "" {
		if parsedGID, parseErr := strconv.Atoi(sudoGID); parseErr == nil {
			gid = parsedGID
		}
	}

	return uid, gid, nil
}

/*
IsRunningAsSudo checks if the current process is running under sudo.
*/
func IsRunningAsSudo() bool {
	return os.Getenv("SUDO_USER") != ""
}
