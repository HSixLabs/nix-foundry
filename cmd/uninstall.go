// Package cmd provides the command-line interface for Nix Foundry.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/shawnkhoffman/nix-foundry/pkg/nix"
	"github.com/shawnkhoffman/nix-foundry/pkg/schema"
	"github.com/shawnkhoffman/nix-foundry/pkg/tui"
	"github.com/spf13/cobra"
)

var (
	force bool
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall Nix Foundry",
	Long: `Uninstall Nix Foundry.
This command will remove Nix Foundry and optionally uninstall Nix.`,
	RunE: runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().BoolVar(&force, "force", false, "Force uninstallation even if errors occur")
}

func runUninstall(_ *cobra.Command, _ []string) error {
	uninstallNix, confirmed, err := tui.RunUninstallTUI()
	if err != nil {
		return err
	}

	if !confirmed {
		return fmt.Errorf("uninstallation cancelled")
	}

	if uninstallNix {
		fs := filesystem.NewOSFileSystem()
		installer := nix.NewInstaller(fs)

		if uninstallErr := installer.Uninstall(force); uninstallErr != nil {
			return fmt.Errorf("failed to uninstall Nix: %w", uninstallErr)
		}
	}

	configPath, err := schema.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	configDir := filepath.Dir(configPath)
	if removeErr := os.RemoveAll(configDir); removeErr != nil {
		return fmt.Errorf("failed to remove config directory: %w", removeErr)
	}

	if pathErr := removeFromPath(); pathErr != nil {
		fmt.Printf("Warning: Failed to remove nix-foundry from PATH: %v\n", pathErr)
	}

	fmt.Println("✨ Nix Foundry uninstalled successfully")
	if uninstallNix {
		fmt.Println("✨ Nix uninstalled successfully")
	}
	return nil
}

func removeFromPath() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	rcFiles := []string{
		filepath.Join(homeDir, ".bashrc"),
		filepath.Join(homeDir, ".zshrc"),
	}

	for _, rcFile := range rcFiles {
		if _, statErr := os.Stat(rcFile); statErr == nil {
			content, readErr := os.ReadFile(rcFile)
			if readErr != nil {
				continue
			}

			lines := strings.Split(string(content), "\n")
			var newLines []string
			for _, line := range lines {
				if !strings.Contains(line, "# Add nix-foundry to PATH") &&
					!strings.Contains(line, "export PATH=\"$PATH:$HOME/.local/bin\"") {
					newLines = append(newLines, line)
				}
			}

			newContent := strings.Join(newLines, "\n")
			if writeErr := os.WriteFile(rcFile, []byte(newContent), 0644); writeErr != nil {
				continue
			}
		}
	}

	return nil
}
