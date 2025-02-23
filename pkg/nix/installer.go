// Package nix provides Nix package manager installation functionality.
package nix

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
)

// Installer handles Nix package manager installation.
type Installer struct {
	fs filesystem.FileSystem
}

// NewInstaller creates a new Nix installer.
func NewInstaller(fs filesystem.FileSystem) *Installer {
	return &Installer{fs: fs}
}

// IsInstalled checks if Nix is installed.
func (i *Installer) IsInstalled() bool {
	fmt.Println("Checking Nix installation status...")

	nixPath, err := exec.LookPath("nix")
	if err == nil {
		fmt.Printf("Found nix binary at: %s\n", nixPath)
		cmd := exec.Command("nix", "--version")
		if out, err := cmd.CombinedOutput(); err == nil {
			fmt.Printf("Nix version: %s\n", strings.TrimSpace(string(out)))
			return true
		}
		fmt.Printf("Nix binary found but not working: %v\n", err)
	}

	if i.fs.Exists("/nix/store") {
		fmt.Println("Found Nix store directory")
		return true
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		profile := filepath.Join(homeDir, ".nix-profile")
		if i.fs.Exists(profile) {
			fmt.Println("Found Nix profile")
			return true
		}
	}

	fmt.Println("No working Nix installation found")
	return false
}

// IsMultiUser checks if Nix is installed in multi-user mode.
func (i *Installer) IsMultiUser() (bool, error) {
	fmt.Println("Checking Nix installation mode...")

	if i.fs.Exists("/nix/var/nix/daemon") {
		fmt.Println("Found Nix daemon service")
		return true, nil
	}

	cmd := exec.Command("systemctl", "is-active", "nix-daemon.service")
	if err := cmd.Run(); err == nil {
		fmt.Println("Found active Nix daemon systemd service")
		return true, nil
	}

	if i.fs.Exists("/Library/LaunchDaemons/org.nixos.nix-daemon.plist") {
		fmt.Println("Found Nix daemon launchd service")
		return true, nil
	}

	fmt.Println("No multi-user installation detected")
	return false, nil
}

// cleanupBackupFiles removes old backup files that might interfere with installation.
func (i *Installer) cleanupBackupFiles() error {
	backupFiles := []string{
		"/etc/bashrc.backup-before-nix",
		"/etc/zshrc.backup-before-nix",
		"/etc/bash.bashrc.backup-before-nix",
	}

	for _, file := range backupFiles {
		if i.fs.Exists(file) {
			fmt.Printf("Removing old backup file: %s\n", file)
			cmd := exec.Command("sudo", "rm", "-f", file)
			if err := cmd.Run(); err != nil {
				fmt.Printf("Warning: Failed to remove backup file %s: %v\n", file, err)
			}
		}
	}

	return nil
}

// Install installs Nix in either single-user or multi-user mode.
func (i *Installer) Install(multiUser bool) error {
	fmt.Printf("Installing Nix in %s mode...\n",
		map[bool]string{true: "multi-user", false: "single-user"}[multiUser])

	if err := i.cleanupBackupFiles(); err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "nix-install-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	scriptPath := filepath.Join(tmpDir, "install.sh")
	fmt.Println("Downloading Nix installation script...")
	cmd := exec.Command("curl", "-L", "https://nixos.org/nix/install", "-o", scriptPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download installation script: %w", err)
	}

	if err := os.Chmod(scriptPath, 0755); err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	fmt.Println("Running Nix installation script...")
	var installCmd *exec.Cmd
	if multiUser {
		installCmd = exec.Command("sh", scriptPath, "--daemon")
	} else {
		installCmd = exec.Command("sh", scriptPath)
	}

	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("installation script failed: %w", err)
	}

	fmt.Println("Waiting for installation to complete...")
	time.Sleep(2 * time.Second)

	fmt.Println("Verifying installation...")
	if !i.IsInstalled() {
		return fmt.Errorf("installation verification failed")
	}

	return nil
}

// Uninstall removes Nix installation.
func (i *Installer) Uninstall(force bool) error {
	fmt.Println("Starting Nix uninstallation...")

	if !i.IsInstalled() {
		fmt.Println("No Nix installation found")
		return nil
	}

	if !force {
		fmt.Println("Checking for running Nix processes...")
		cmd := exec.Command("pgrep", "-f", "nix")
		if err := cmd.Run(); err == nil {
			return fmt.Errorf("nix processes are still running. Please stop them first or use --force")
		}
	}

	if multiUser, _ := i.IsMultiUser(); multiUser {
		fmt.Println("Stopping Nix daemon service...")
		stopCmd := exec.Command("sudo", "systemctl", "stop", "nix-daemon.service")
		stopCmd.Run()

		stopCmd = exec.Command("sudo", "launchctl", "unload", "/Library/LaunchDaemons/org.nixos.nix-daemon.plist")
		stopCmd.Run()
	}

	paths := []string{
		"/nix",
		"/etc/nix",
		"/etc/profile.d/nix.sh",
		"/etc/synthetic.conf",
		"/etc/fstab",
		"/Library/LaunchDaemons/org.nixos.nix-daemon.plist",
		"/Library/LaunchDaemons/org.nixos.darwin-store.plist",
	}

	if homeDir, err := os.UserHomeDir(); err == nil {
		paths = append(paths,
			filepath.Join(homeDir, ".nix-profile"),
			filepath.Join(homeDir, ".nix-defexpr"),
			filepath.Join(homeDir, ".nix-channels"),
		)
	}

	shellFiles := []string{
		"/etc/bashrc",
		"/etc/zshrc",
		"/etc/bash.bashrc",
	}

	for _, file := range shellFiles {
		if i.fs.Exists(file) {
			content, err := os.ReadFile(file)
			if err != nil {
				continue
			}

			lines := strings.Split(string(content), "\n")
			var newLines []string
			for _, line := range lines {
				if !strings.Contains(line, "nix") {
					newLines = append(newLines, line)
				}
			}

			newContent := strings.Join(newLines, "\n")
			if force {
				cmd := exec.Command("sudo", "tee", file)
				cmd.Stdin = strings.NewReader(newContent)
				cmd.Run()
			} else {
				os.WriteFile(file, []byte(newContent), 0644)
			}
		}
	}

	fmt.Println("Removing Nix files and directories...")
	for _, path := range paths {
		if !i.fs.Exists(path) {
			fmt.Printf("Skipping non-existent path: %s\n", path)
			continue
		}

		fmt.Printf("Removing: %s\n", path)
		if force {
			cmd := exec.Command("sudo", "rm", "-rf", path)
			if err := cmd.Run(); err != nil {
				fmt.Printf("Warning: Failed to force remove %s: %v\n", path, err)
			}
		} else {
			if err := os.RemoveAll(path); err != nil {
				if strings.Contains(err.Error(), "resource busy") {
					return fmt.Errorf("failed to remove %s: resource busy. Try using --force", path)
				}
				return fmt.Errorf("failed to remove %s: %w", path, err)
			}
		}
	}

	if err := i.cleanupBackupFiles(); err != nil {
		fmt.Printf("Warning: Failed to clean up backup files: %v\n", err)
	}

	fmt.Println("Verifying uninstallation...")
	if i.IsInstalled() {
		if force {
			fmt.Println("Warning: Nix appears to still be installed, but continuing due to --force")
			return nil
		}
		return fmt.Errorf("uninstallation verification failed")
	}

	fmt.Println("Nix uninstallation completed successfully")
	return nil
}
