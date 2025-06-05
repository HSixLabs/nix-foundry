/*
Package cmd provides the command-line interface for Nix Foundry.
It implements various commands for managing Nix installations, package management,
and system configuration. The commands follow the Cobra command pattern for
consistent CLI behavior.
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/shawnkhoffman/nix-foundry/pkg/nix"
	"github.com/shawnkhoffman/nix-foundry/pkg/platform"
	"github.com/shawnkhoffman/nix-foundry/pkg/schema"
	"github.com/shawnkhoffman/nix-foundry/pkg/tui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	multiUser bool
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Nix package manager",
	Long: `Install Nix package manager.
This command will install Nix in either single-user or multi-user mode.`,
	RunE: runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().BoolVar(&multiUser, "multi-user", false, "Install in multi-user mode (requires sudo)")
}

/*
getCurrentShell retrieves the current user's shell from the SHELL environment
variable and returns just the base name of the shell (e.g., "bash", "zsh").
*/
func getCurrentShell() string {
	shell := os.Getenv("SHELL")
	return filepath.Base(shell)
}

/*
installShell installs the specified shell using nix-env if it's different from
the current shell. It also configures the system to use the new shell by:
1. Installing the shell package via nix-env
2. Adding the shell to /etc/shells if possible
3. Attempting to change the user's default shell
If any step fails, appropriate warnings are displayed but the process continues.
*/
func installShell(shell string) error {
	currentShell := getCurrentShell()
	if shell == currentShell {
		return nil
	}

	fmt.Printf("Installing %s shell...\n", shell)

	nixEnvCmd := fmt.Sprintf(". %s && NIXPKGS_ALLOW_UNFREE=1 NIXPKGS_ALLOW_UNSUPPORTED_SYSTEM=1 /nix/var/nix/profiles/default/bin/nix-env -iA nixpkgs.%s -Q",
		"/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh",
		shell)
	execCmd := exec.Command("bash", "-c", nixEnvCmd)
	if output, err := execCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install %s: %s: %w", shell, output, err)
	}

	shellPath := filepath.Join("/nix/var/nix/profiles/default/bin", shell)

	execCmd = exec.Command("sudo", "sh", "-c", fmt.Sprintf("command -v %s >> /etc/shells 2>/dev/null || true", shellPath))
	_ = execCmd.Run()

	if _, err := exec.LookPath("chsh"); err == nil {
		execCmd = exec.Command("chsh", "-s", shellPath)
		if err := execCmd.Run(); err != nil {
			fmt.Printf("Warning: Failed to change shell to %s: %v\nYou may need to change your shell manually.\n", shell, err)
		} else {
			fmt.Printf("Successfully changed shell to %s. Please log out and back in for changes to take effect.\n", shell)
		}
	} else {
		fmt.Printf("Note: Shell change command not available. You may need to change your shell to %s manually.\n", shellPath)
	}
	return nil
}

/*
determineMultiUserMode determines if multi-user mode is required based on:
1. Platform requirements (e.g., macOS always needs multi-user mode)
2. Selected packages that require multi-user mode (e.g., docker)
*/
func determineMultiUserMode(packages []string) bool {
	if platform.GetNixSystem() == "aarch64-darwin" || platform.GetNixSystem() == "x86_64-darwin" {
		return true
	}

	for _, pkg := range packages {
		switch pkg {
		case "docker":
			return true
		}
	}
	return false
}

/*
createInitialConfig creates and saves the initial configuration with the provided settings.
*/
func createInitialConfig(manager, shell string, packages []string) error {
	config := schema.NewDefaultConfig()
	config.Settings.Shell = shell
	config.Nix.Manager = manager
	config.Nix.Packages.Optional = packages

	configPath, pathErr := schema.GetConfigPath()
	if pathErr != nil {
		return fmt.Errorf("failed to get config path: %w", pathErr)
	}

	configDir := filepath.Dir(configPath)
	if mkdirErr := os.MkdirAll(configDir, 0755); mkdirErr != nil {
		return fmt.Errorf("failed to create config directory: %w", mkdirErr)
	}

	uid, gid, err := platform.GetRealUser()
	if err != nil {
		return fmt.Errorf("failed to get real user: %w", err)
	}

	if chownErr := os.Chown(configDir, uid, gid); chownErr != nil {
		return fmt.Errorf("failed to set config directory ownership: %w", chownErr)
	}

	content, marshalErr := yaml.Marshal(config)
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal config: %w", marshalErr)
	}

	if writeErr := os.WriteFile(configPath, content, 0644); writeErr != nil {
		return fmt.Errorf("failed to write config: %w", writeErr)
	}

	if chownErr := os.Chown(configPath, uid, gid); chownErr != nil {
		return fmt.Errorf("failed to set config file ownership: %w", chownErr)
	}

	return nil
}

/*
initializeNixChannels initializes the Nix channels and waits for the daemon to be ready.
*/
func initializeNixChannels() error {
	fmt.Println("Initializing Nix channels...")
	cmd := exec.Command("bash", "-c", ". /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh && nix-channel --add https://nixos.org/channels/nixpkgs-unstable && nix-channel --update")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: Failed to initialize Nix channels: %v\n", err)
	}

	fmt.Println("Waiting for Nix daemon to be ready...")
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			fmt.Printf("Retrying in 2 seconds (attempt %d/%d)...\n", i+1, maxRetries)
			time.Sleep(2 * time.Second)
		}

		checkCmd := exec.Command("bash", "-c", "/nix/var/nix/profiles/default/bin/nix-env --version")
		if checkErr := checkCmd.Run(); checkErr == nil {
			return nil
		}

		if i == maxRetries-1 {
			return fmt.Errorf("nix daemon not ready after %d attempts", maxRetries)
		}
	}

	return nil
}

/*
configureNixSettings creates and configures Nix settings for the current user.
*/
func configureNixSettings(uid, gid int) error {
	homeDir, homeDirErr := platform.GetRealUserHomeDir()
	if homeDirErr != nil {
		return fmt.Errorf("failed to get home directory: %w", homeDirErr)
	}

	nixConfigDir := filepath.Join(homeDir, ".config", "nixpkgs")
	if mkdirErr := os.MkdirAll(nixConfigDir, 0755); mkdirErr != nil {
		return fmt.Errorf("failed to create nixpkgs config directory: %w", mkdirErr)
	}

	if chownErr := os.Chown(nixConfigDir, uid, gid); chownErr != nil {
		return fmt.Errorf("failed to set nixpkgs config directory ownership: %w", chownErr)
	}

	nixConfig := fmt.Sprintf(`{
  allowUnfree = true;
  allowUnsupportedSystem = true;
  crossSystem = null;
  system = "%s";
}
`, platform.GetNixSystem())
	nixConfigPath := filepath.Join(nixConfigDir, "config.nix")
	if writeErr := os.WriteFile(nixConfigPath, []byte(nixConfig), 0644); writeErr != nil {
		return fmt.Errorf("failed to write nixpkgs config: %w", writeErr)
	}

	if chownErr := os.Chown(nixConfigPath, uid, gid); chownErr != nil {
		return fmt.Errorf("failed to set nixpkgs config file ownership: %w", chownErr)
	}

	return nil
}

/*
runInstall handles the main installation process for Nix Foundry. It:
1. Verifies proper permissions for multi-user installation
2. Runs the installation TUI to gather user preferences
3. Determines if multi-user mode is required based on platform and package selection
4. Creates and saves initial configuration
5. Installs Nix package manager
6. Sets up selected shell and packages
7. Configures system paths and environment

Returns an error if any critical step fails.
*/
func runInstall(_ *cobra.Command, _ []string) error {
	if multiUser && os.Geteuid() != 0 {
		return fmt.Errorf("multi-user installation requires root privileges. Please run with sudo")
	}

	manager, shell, packages, confirmed, initErr := tui.RunInstallTUI()
	if initErr != nil {
		return initErr
	}

	if !confirmed {
		return fmt.Errorf("installation cancelled")
	}

	multiUser = determineMultiUserMode(packages)
	if multiUser && os.Geteuid() != 0 {
		var reason string
		if platform.GetNixSystem() == "aarch64-darwin" || platform.GetNixSystem() == "x86_64-darwin" {
			reason = "macOS requires multi-user mode"
		} else {
			reason = "selected packages (docker) require multi-user mode"
		}
		return fmt.Errorf("multi-user installation is required (%s). Please run with sudo", reason)
	}

	if configErr := createInitialConfig(manager, shell, packages); configErr != nil {
		return configErr
	}

	fs := filesystem.NewOSFileSystem()
	installer := nix.NewInstaller(fs)

	if installer.IsInstalled() {
		currentMultiUser, modeErr := installer.IsMultiUser()
		if modeErr != nil {
			return fmt.Errorf("failed to check installation mode: %w", modeErr)
		}

		if currentMultiUser == multiUser {
			fmt.Println("✨ Nix is already installed in the requested mode")
			return nil
		}

		fmt.Printf("Nix is already installed in %s mode. Please uninstall first to change modes.\n",
			map[bool]string{true: "multi-user", false: "single-user"}[currentMultiUser])
		return nil
	}

	if installErr := installer.Install(multiUser); installErr != nil {
		return fmt.Errorf("installation failed: %w", installErr)
	}

	if shellErr := installShell(shell); shellErr != nil {
		fmt.Printf("Warning: Failed to install shell: %v\n", shellErr)
	}

	if pathErr := addToPath(shell); pathErr != nil {
		fmt.Printf("Warning: Failed to add nix-foundry to PATH: %v\n", pathErr)
	}

	fmt.Printf("✨ Nix installed successfully in %s mode\n",
		map[bool]string{true: "multi-user", false: "single-user"}[multiUser])

	if channelErr := initializeNixChannels(); channelErr != nil {
		return channelErr
	}

	uid, gid, err := platform.GetRealUser()
	if err != nil {
		return fmt.Errorf("failed to get real user: %w", err)
	}

	if configErr := configureNixSettings(uid, gid); configErr != nil {
		return configErr
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎉 INSTALLATION COMPLETE!")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Close and reopen your terminal (or run: source ~/.zshrc)")
	fmt.Println("2. Install your selected packages by running:")
	fmt.Println("   nix-foundry config apply")
	fmt.Println()
	fmt.Println("Note: Package installation runs in user context to avoid permission issues.")

	return nil
}

/*
addToPath adds nix-foundry to the user's PATH by modifying their shell configuration file.
*/
func addToPath(shell string) error {
	rcFile, err := platform.GetShellConfigFile(shell)
	if err != nil {
		return fmt.Errorf("failed to get shell config file: %w", err)
	}

	if platform.IsRunningAsSudo() {
		realHomeDir, homeErr := platform.GetRealUserHomeDir()
		if homeErr != nil {
			return fmt.Errorf("failed to get real user home directory: %w", homeErr)
		}
		currentHome, _ := os.UserHomeDir()
		rcFile = strings.Replace(rcFile, currentHome, realHomeDir, 1)
	}

	if shell == "fish" {
		if mkdirErr := os.MkdirAll(filepath.Dir(rcFile), 0775); mkdirErr != nil {
			return fmt.Errorf("failed to create fish config directory: %w", mkdirErr)
		}

		if platform.IsRunningAsSudo() {
			uid, gid, userErr := platform.GetRealUser()
			if userErr == nil {
				if chownErr := os.Chown(filepath.Dir(rcFile), uid, gid); chownErr != nil {
					fmt.Printf("Warning: Failed to set fish config directory ownership: %v\n", chownErr)
				}
			}
		}
	}

	var content string
	switch shell {
	case "fish":
		content = `
if test -e '/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.fish'
    source '/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.fish'
else if test -e "$HOME/.nix-profile/etc/profile.d/nix.fish"
    source "$HOME/.nix-profile/etc/profile.d/nix.fish"
end

if not contains $HOME/.local/bin $PATH
    set -x PATH $PATH $HOME/.local/bin
end
`
	case "zsh":
		content = `
if [ -e '/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh' ]; then
    . '/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh'
elif [ -e "$HOME/.nix-profile/etc/profile.d/nix.sh" ]; then
    . "$HOME/.nix-profile/etc/profile.d/nix.sh"
fi

if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    export PATH="$PATH:$HOME/.local/bin"
fi
`
	default:
		content = `
if [ -e '/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh' ]; then
    . '/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh'
elif [ -e "$HOME/.nix-profile/etc/profile.d/nix.sh" ]; then
    . "$HOME/.nix-profile/etc/profile.d/nix.sh"
fi

if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
    export PATH="$PATH:$HOME/.local/bin"
fi
`
	}

	existingContent, readErr := os.ReadFile(rcFile)
	if readErr == nil && len(existingContent) > 0 {
		if strings.Contains(string(existingContent), "nix-daemon") {
			return nil
		}
	}

	f, openErr := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if openErr != nil {
		return fmt.Errorf("failed to open rc file: %w", openErr)
	}
	defer func() { _ = f.Close() }()

	if _, writeErr := f.WriteString(content); writeErr != nil {
		return fmt.Errorf("failed to update rc file: %w", writeErr)
	}

	if platform.IsRunningAsSudo() {
		uid, gid, userErr := platform.GetRealUser()
		if userErr == nil {
			if chownErr := os.Chown(rcFile, uid, gid); chownErr != nil {
				fmt.Printf("Warning: Failed to set rc file ownership: %v\n", chownErr)
			}
		}
	}

	return nil
}
