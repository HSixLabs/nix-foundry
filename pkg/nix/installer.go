/*
Package nix provides Nix package manager installation functionality.
It handles installation, uninstallation, and verification of Nix package manager
installations in both single-user and multi-user modes.
*/
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

/*
Installer handles Nix package manager installation operations.
It provides functionality for installing, uninstalling, and verifying
Nix installations using a provided filesystem abstraction.
*/
type Installer struct {
	fs filesystem.FileSystem
}

/*
NewInstaller creates a new Nix installer with the provided filesystem implementation.
*/
func NewInstaller(fs filesystem.FileSystem) *Installer {
	return &Installer{fs: fs}
}

/*
IsInstalled checks if Nix is installed by verifying:
1. The presence of the nix binary in PATH
2. The existence of the Nix store directory
3. The existence of a Nix profile
Returns true if any of these checks succeed.
*/
func (i *Installer) IsInstalled() bool {
	fmt.Println("Checking Nix installation status...")

	nixPath, lookPathErr := exec.LookPath("nix")
	if lookPathErr == nil {
		fmt.Printf("Found nix binary at: %s\n", nixPath)
		cmd := exec.Command("nix", "--version")
		if out, versionErr := cmd.CombinedOutput(); versionErr == nil {
			fmt.Printf("Nix version: %s\n", strings.TrimSpace(string(out)))
			return true
		}
		fmt.Printf("Nix binary found but not working: %v\n", lookPathErr)
	}

	if i.fs.Exists("/nix/store") {
		fmt.Println("Found Nix store directory")
		return true
	}

	homeDir, homeDirErr := os.UserHomeDir()
	if homeDirErr == nil {
		profile := filepath.Join(homeDir, ".nix-profile")
		if i.fs.Exists(profile) {
			fmt.Println("Found Nix profile")
			return true
		}
	}

	fmt.Println("No working Nix installation found")
	return false
}

/*
IsMultiUser checks if Nix is installed in multi-user mode by checking for:
1. The presence of systemd service on Linux
2. The presence of launchd service on macOS
3. The existence of the Nix daemon directory
Returns true if any of these checks succeed.
*/
func (i *Installer) IsMultiUser() (bool, error) {
	fmt.Println("Checking Nix installation mode...")

	if i.fs.Exists("/etc/systemd/system/nix-daemon.service") {
		fmt.Println("Found Nix daemon systemd service")
		return true, nil
	}

	if i.fs.Exists("/Library/LaunchDaemons/org.nixos.nix-daemon.plist") {
		fmt.Println("Found Nix daemon launchd service")
		return true, nil
	}

	if i.fs.Exists("/nix/var/nix/daemon") {
		fmt.Println("Found Nix daemon service")
		return true, nil
	}

	fmt.Println("No multi-user installation detected")
	return false, nil
}

/*
cleanupBackupFiles removes old backup files that might interfere with installation.
It removes backup files created by previous Nix installations.
*/
func (i *Installer) cleanupBackupFiles() error {
	backupFiles := []string{
		"/etc/bashrc.backup-before-nix",
		"/etc/zshrc.backup-before-nix",
		"/etc/bash.bashrc.backup-before-nix",
	}

	var errs []error
	for _, file := range backupFiles {
		if i.fs.Exists(file) {
			fmt.Printf("Removing old backup file: %s\n", file)
			removeCmd := exec.Command("sudo", "rm", "-fv", file)
			removeCmd.Stdout = os.Stdout
			removeCmd.Stderr = os.Stderr
			if removeErr := removeCmd.Run(); removeErr != nil {
				errs = append(errs, fmt.Errorf("failed to remove backup file %s: %w", file, removeErr))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to cleanup backup files: %v", errs)
	}
	return nil
}

/*
Install installs Nix in either single-user or multi-user mode.
It performs the following steps:
1. Cleans up any old backup files
2. Downloads the Nix installation script
3. Executes the installation script with appropriate flags
4. Verifies the installation was successful
*/
func (i *Installer) Install(multiUser bool) error {
	fmt.Printf("Installing Nix in %s mode...\n",
		map[bool]string{true: "multi-user", false: "single-user"}[multiUser])

	if cleanupErr := i.cleanupBackupFiles(); cleanupErr != nil {
		return cleanupErr
	}

	tmpDir, tmpDirErr := os.MkdirTemp("", "nix-install-*")
	if tmpDirErr != nil {
		return fmt.Errorf("failed to create temp directory: %w", tmpDirErr)
	}
	defer os.RemoveAll(tmpDir)

	scriptPath := filepath.Join(tmpDir, "install.sh")
	fmt.Println("Downloading Nix...")
	cmd := exec.Command("curl", "-L", "--progress-bar", "https://nixos.org/nix/install", "-o", scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to download Nix: %w", err)
	}

	if chmodErr := os.Chmod(scriptPath, 0755); chmodErr != nil {
		return fmt.Errorf("failed to make script executable: %w", chmodErr)
	}

	fmt.Println("Installing Nix...")
	var installCmd *exec.Cmd
	if multiUser {
		installCmd = exec.Command("sh", scriptPath, "--daemon")
	} else {
		installCmd = exec.Command("sh", scriptPath)
	}

	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if installErr := installCmd.Run(); installErr != nil {
		return fmt.Errorf("failed to install Nix: %w", installErr)
	}

	fmt.Println("Waiting for installation to complete...")
	time.Sleep(2 * time.Second)

	fmt.Println("Verifying installation...")
	if !i.IsInstalled() {
		return fmt.Errorf("installation verification failed")
	}

	return nil
}

/*
uninstallPackages uninstalls all packages from both multi-user and single-user profiles.
*/
func (i *Installer) uninstallPackages() {
	fmt.Println("Uninstalling all Nix packages...")

	// Multi-user profile
	fmt.Println("Checking multi-user profile...")
	listCmd := exec.Command("bash", "-c", ". /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh && nix-env -q")
	listCmd.Stdout = os.Stdout
	listCmd.Stderr = os.Stderr
	output, listErr := listCmd.Output()
	if listErr == nil {
		i.uninstallPackagesFromOutput(string(output), "/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh")
	} else {
		fmt.Printf("Note: No packages found in multi-user profile or profile not accessible\n")
	}

	// Single-user profile
	homeDir, homeDirErr := os.UserHomeDir()
	if homeDirErr == nil {
		fmt.Println("Checking single-user profile...")
		profilePath := filepath.Join(homeDir, ".nix-profile/etc/profile.d/nix.sh")
		listCmd = exec.Command("bash", "-c", fmt.Sprintf(". %s && nix-env -q", profilePath))
		listCmd.Stdout = os.Stdout
		listCmd.Stderr = os.Stderr
		output, listErr = listCmd.Output()
		if listErr == nil {
			i.uninstallPackagesFromOutput(string(output), profilePath)
		} else {
			fmt.Printf("Note: No packages found in single-user profile or profile not accessible\n")
		}
	}
}

/*
uninstallPackagesFromOutput uninstalls packages from a specific profile.
*/
func (i *Installer) uninstallPackagesFromOutput(output, profilePath string) {
	packages := strings.Split(strings.TrimSpace(output), "\n")
	for _, pkg := range packages {
		if pkg == "" {
			continue
		}
		fmt.Printf("Uninstalling package: %s\n", pkg)
		uninstallCmd := exec.Command("bash", "-c", fmt.Sprintf(". %s && nix-env -e %s", profilePath, pkg))
		uninstallCmd.Stdout = os.Stdout
		uninstallCmd.Stderr = os.Stderr
		if uninstallErr := uninstallCmd.Run(); uninstallErr != nil {
			fmt.Printf("Warning: Failed to uninstall package %s: %v\n", pkg, uninstallErr)
		}
	}
}

/*
stopDaemonServices stops Nix daemon services in multi-user mode.
*/
func (i *Installer) stopDaemonServices() {
	multiUser, multiUserErr := i.IsMultiUser()
	if multiUserErr == nil && multiUser {
		fmt.Println("Stopping Nix daemon service...")

		if i.fs.Exists("/etc/systemd/system/nix-daemon.service") {
			fmt.Println("Stopping systemd service...")
			stopCmd := exec.Command("sudo", "systemctl", "stop", "nix-daemon.service")
			stopCmd.Stdout = os.Stdout
			stopCmd.Stderr = os.Stderr
			if stopErr := stopCmd.Run(); stopErr != nil {
				fmt.Printf("Warning: Failed to stop systemd service: %v\n", stopErr)
			}
		}

		if i.fs.Exists("/Library/LaunchDaemons/org.nixos.nix-daemon.plist") {
			fmt.Println("Unloading launchd service...")
			stopCmd := exec.Command("sudo", "launchctl", "unload", "/Library/LaunchDaemons/org.nixos.nix-daemon.plist")
			stopCmd.Stdout = os.Stdout
			stopCmd.Stderr = os.Stderr
			if stopErr := stopCmd.Run(); stopErr != nil {
				fmt.Printf("Warning: Failed to unload launchd service: %v\n", stopErr)
			}
		}
	}
}

/*
cleanupShellFiles removes Nix-related lines from shell configuration files.
*/
func (i *Installer) cleanupShellFiles(force bool) {
	shellFiles := []string{
		"/etc/bashrc",
		"/etc/zshrc",
		"/etc/bash.bashrc",
	}

	for _, file := range shellFiles {
		if i.fs.Exists(file) {
			content, readErr := os.ReadFile(file)
			if readErr != nil {
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
				writeCmd := exec.Command("sudo", "tee", file)
				writeCmd.Stdin = strings.NewReader(newContent)
				if writeErr := writeCmd.Run(); writeErr != nil {
					fmt.Printf("Warning: Failed to update shell file %s: %v\n", file, writeErr)
				}
			} else {
				if writeErr := os.WriteFile(file, []byte(newContent), 0644); writeErr != nil {
					fmt.Printf("Warning: Failed to update shell file %s: %v\n", file, writeErr)
				}
			}
		}
	}
}

/*
removeNixPaths removes all Nix-related files and directories.
*/
func (i *Installer) removeNixPaths(force bool) error {
	paths := []string{
		"/nix",
		"/etc/nix",
		"/etc/profile.d/nix.sh",
		"/etc/synthetic.conf",
		"/etc/fstab",
		"/Library/LaunchDaemons/org.nixos.nix-daemon.plist",
		"/Library/LaunchDaemons/org.nixos.darwin-store.plist",
	}

	if userHomeDir, homeDirErr := os.UserHomeDir(); homeDirErr == nil {
		paths = append(paths,
			filepath.Join(userHomeDir, ".nix-profile"),
			filepath.Join(userHomeDir, ".nix-defexpr"),
			filepath.Join(userHomeDir, ".nix-channels"),
		)
	}

	for _, path := range paths {
		if !i.fs.Exists(path) {
			fmt.Printf("Skipping non-existent path: %s\n", path)
			continue
		}

		fmt.Printf("Removing: %s\n", path)
		if path == "/nix" {
			i.unmountNix()
		}

		removeCmd := exec.Command("sudo", "rm", "-rfv", path)
		removeCmd.Stdout = os.Stdout
		removeCmd.Stderr = os.Stderr
		if removeErr := removeCmd.Run(); removeErr != nil {
			if !force {
				return fmt.Errorf("failed to remove %s: %w", path, removeErr)
			}
			fmt.Printf("Warning: Failed to remove %s: %v\n", path, removeErr)
		}
	}
	return nil
}

/*
unmountNix attempts to unmount the Nix store on both macOS and Linux.
*/
func (i *Installer) unmountNix() {
	if i.fs.Exists("/usr/sbin/diskutil") {
		fmt.Println("Attempting macOS unmount...")
		unmountCmd := exec.Command("sudo", "diskutil", "unmount", "force", "/nix")
		unmountCmd.Stdout = os.Stdout
		unmountCmd.Stderr = os.Stderr
		if unmountErr := unmountCmd.Run(); unmountErr != nil {
			fmt.Printf("Warning: Failed to unmount /nix: %v\n", unmountErr)
		}
	}

	if i.fs.Exists("/bin/umount") {
		fmt.Println("Attempting Linux unmount...")
		unmountCmd := exec.Command("sudo", "umount", "-f", "/nix")
		unmountCmd.Stdout = os.Stdout
		unmountCmd.Stderr = os.Stderr
		if unmountErr := unmountCmd.Run(); unmountErr != nil {
			fmt.Printf("Warning: Failed to unmount /nix: %v\n", unmountErr)
		}
	}
}

/*
Uninstall removes Nix installation from the system.
It performs the following steps:
1. Verifies Nix is installed
2. Checks for running Nix processes (unless force is true)
3. Uninstalls all packages
4. Stops Nix daemon services if in multi-user mode
5. Removes Nix files and directories
6. Cleans up shell configurations
7. Verifies uninstallation was successful

The force parameter allows bypassing certain checks and errors.
*/
func (i *Installer) Uninstall(force bool) error {
	fmt.Println("Starting Nix uninstallation...")

	if !i.IsInstalled() {
		fmt.Println("No Nix installation found")
		return nil
	}

	if !force {
		fmt.Println("Checking for running Nix processes...")
		processCmd := exec.Command("pgrep", "-f", "nix")
		if processErr := processCmd.Run(); processErr == nil {
			return fmt.Errorf("nix processes are still running. Please stop them first or use --force")
		}
	}

	i.uninstallPackages()
	i.stopDaemonServices()
	i.cleanupShellFiles(force)

	if removeErr := i.removeNixPaths(force); removeErr != nil {
		return removeErr
	}

	if cleanupErr := i.cleanupBackupFiles(); cleanupErr != nil {
		fmt.Printf("Warning: Failed to clean up backup files: %v\n", cleanupErr)
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
