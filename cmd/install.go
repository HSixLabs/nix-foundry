// Package cmd provides the command-line interface for Nix Foundry.
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/shawnkhoffman/nix-foundry/pkg/nix"
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

func getCurrentShell() string {
	shell := os.Getenv("SHELL")
	return filepath.Base(shell)
}

func installShell(shell string) error {
	currentShell := getCurrentShell()
	if shell == currentShell {
		return nil
	}

	fmt.Printf("Installing %s shell...\n", shell)

	// Install shell using nix-env
	cmd := exec.Command("nix-env", "-iA", "nixpkgs."+shell)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install %s: %w", shell, err)
	}

	// Add shell to /etc/shells if it's not already there
	shellPath := fmt.Sprintf("/run/current-system/sw/bin/%s", shell)
	cmd = exec.Command("sudo", "sh", "-c", fmt.Sprintf("echo %s >> /etc/shells", shellPath))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add %s to /etc/shells: %w", shell, err)
	}

	// Change user's shell
	cmd = exec.Command("chsh", "-s", shellPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to change shell to %s: %w", shell, err)
	}

	fmt.Printf("Successfully changed shell to %s. Please log out and back in for changes to take effect.\n", shell)
	return nil
}

func runInstall(cmd *cobra.Command, args []string) error {
	// Check if running as root for multi-user installation
	if multiUser && os.Geteuid() != 0 {
		return fmt.Errorf("multi-user installation requires root privileges. Please run with sudo")
	}

	// Run installation TUI
	manager, shell, packages, confirmed, err := tui.RunInstallTUI()
	if err != nil {
		return err
	}

	if !confirmed {
		return fmt.Errorf("installation cancelled")
	}

	// Create configuration
	config := schema.NewDefaultConfig()
	config.Settings.Shell = shell
	config.Nix.Manager = manager
	config.Nix.Packages.Optional = packages

	// Save configuration
	configPath, err := schema.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	configDir := filepath.Dir(configPath)
	if mkdirErr := os.MkdirAll(configDir, 0755); mkdirErr != nil {
		return fmt.Errorf("failed to create config directory: %w", mkdirErr)
	}

	content, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	// Install Nix
	fs := filesystem.NewOSFileSystem()
	installer := nix.NewInstaller(fs)

	// Check if Nix is already installed
	if installer.IsInstalled() {
		// If installed, check if it's the correct mode
		currentMultiUser, err := installer.IsMultiUser()
		if err != nil {
			return fmt.Errorf("failed to check installation mode: %w", err)
		}

		if currentMultiUser == multiUser {
			fmt.Println("✨ Nix is already installed in the requested mode")
			return nil
		}

		// If mode mismatch, ask to uninstall first
		fmt.Printf("Nix is already installed in %s mode. Please uninstall first to change modes.\n",
			map[bool]string{true: "multi-user", false: "single-user"}[currentMultiUser])
		return nil
	}

	if err := installer.Install(multiUser); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	// Install selected shell if different from current
	if err := installShell(shell); err != nil {
		fmt.Printf("Warning: Failed to install shell: %v\n", err)
	}

	// Add nix-foundry to PATH
	if err := addToPath(shell); err != nil {
		fmt.Printf("Warning: Failed to add nix-foundry to PATH: %v\n", err)
	}

	fmt.Printf("✨ Nix installed successfully in %s mode\n",
		map[bool]string{true: "multi-user", false: "single-user"}[multiUser])
	return nil
}

func addToPath(shell string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
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
		return fmt.Errorf("unsupported shell: %s", shell)
	}

	// Create config directory for fish if needed
	if shell == "fish" {
		if mkdirErr := os.MkdirAll(filepath.Dir(rcFile), 0755); mkdirErr != nil {
			return fmt.Errorf("failed to create fish config directory: %w", mkdirErr)
		}
	}

	// Add PATH update to shell rc file
	var content string
	switch shell {
	case "fish":
		content = "\n# Add nix-foundry to PATH\nset -x PATH $PATH $HOME/.local/bin\n"
	default:
		content = "\n# Add nix-foundry to PATH\nexport PATH=\"$PATH:$HOME/.local/bin\"\n"
	}

	f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open rc file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("failed to update rc file: %w", err)
	}

	return nil
}
