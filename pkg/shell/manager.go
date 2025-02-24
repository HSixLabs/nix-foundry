// Package shell provides functionality for managing shell configurations
// across different platforms and shell types.
package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/shawnkhoffman/nix-foundry/pkg/platform"
)

/*
Manager handles shell configuration operations across different platforms.
It provides functionality for configuring, updating, and managing shell
environments for Nix integration.
*/
type Manager struct {
	fs filesystem.FileSystem
}

// NewManager creates a new shell manager instance with the provided filesystem.
func NewManager(fs filesystem.FileSystem) *Manager {
	return &Manager{fs: fs}
}

/*
ConfigureShell configures the specified shell with Nix environment settings.
It handles reading existing configurations, generating shell-specific settings,
and updating or creating configuration files as needed.
*/
func (m *Manager) ConfigureShell(shell string) error {
	configFile, err := platform.GetShellConfigFile(shell)
	if err != nil {
		return fmt.Errorf("failed to get shell config file: %w", err)
	}

	if createDirErr := m.fs.CreateDir(filepath.Dir(configFile)); createDirErr != nil {
		return fmt.Errorf("failed to create config directory: %w", createDirErr)
	}

	var content []byte
	if m.fs.Exists(configFile) {
		content, err = m.fs.ReadFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}

	config := m.generateShellConfig(shell)

	if strings.Contains(string(content), "# Nix") {
		newContent := m.updateExistingConfig(string(content), config)
		if writeErr := m.fs.WriteFile(configFile, []byte(newContent), 0644); writeErr != nil {
			return fmt.Errorf("failed to update config file: %w", writeErr)
		}
	} else {
		if len(content) > 0 && !strings.HasSuffix(string(content), "\n") {
			content = append(content, '\n')
		}
		content = append(content, []byte(config)...)
		if writeErr := m.fs.WriteFile(configFile, content, 0644); writeErr != nil {
			return fmt.Errorf("failed to write config file: %w", writeErr)
		}
	}

	return nil
}

/*
RemoveShellConfig removes Nix-related configuration from the shell config file.
It preserves all other configuration settings while removing only Nix-specific sections.
*/
func (m *Manager) RemoveShellConfig(shell string) error {
	configFile, err := platform.GetShellConfigFile(shell)
	if err != nil {
		return fmt.Errorf("failed to get shell config file: %w", err)
	}

	if !m.fs.Exists(configFile) {
		return nil
	}

	content, err := m.fs.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	inNixBlock := false

	for _, line := range lines {
		if strings.Contains(line, "# Nix") {
			inNixBlock = true
			continue
		}
		if inNixBlock {
			if strings.TrimSpace(line) == "" {
				inNixBlock = false
			}
			continue
		}
		newLines = append(newLines, line)
	}

	newContent := strings.Join(newLines, "\n")
	if err := m.fs.WriteFile(configFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

/*
generateShellConfig creates shell-specific Nix configuration content.
It generates different configurations based on the shell type (bash, zsh, fish)
and includes necessary environment variables and path configurations.
*/
func (m *Manager) generateShellConfig(shell string) string {
	var config strings.Builder

	config.WriteString("\n# Nix\n")

	switch shell {
	case "/bin/zsh", "/bin/bash":
		config.WriteString("if [ -e '/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh' ]; then\n")
		config.WriteString("  . '/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh'\n")
		config.WriteString("fi\n")
		config.WriteString("if [ -e '$HOME/.nix-profile/etc/profile.d/nix.sh' ]; then\n")
		config.WriteString("  . '$HOME/.nix-profile/etc/profile.d/nix.sh'\n")
		config.WriteString("fi\n")
		config.WriteString("export PATH=\"$HOME/.nix-profile/bin:$PATH\"\n")
	case "/bin/fish":
		config.WriteString("if test -e '/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.fish'\n")
		config.WriteString("  source '/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.fish'\n")
		config.WriteString("end\n")
		config.WriteString("if test -e '$HOME/.nix-profile/etc/profile.d/nix.fish'\n")
		config.WriteString("  source '$HOME/.nix-profile/etc/profile.d/nix.fish'\n")
		config.WriteString("end\n")
		config.WriteString("set -x PATH $HOME/.nix-profile/bin $PATH\n")
	}

	return config.String()
}

/*
updateExistingConfig updates the Nix configuration block in an existing shell config file.
It preserves all other configuration while replacing only the Nix-specific section.
*/
func (m *Manager) updateExistingConfig(content, newConfig string) string {
	lines := strings.Split(content, "\n")
	var newLines []string
	inNixBlock := false
	nixBlockFound := false

	for _, line := range lines {
		if strings.Contains(line, "# Nix") {
			if !nixBlockFound {
				newLines = append(newLines, strings.TrimSuffix(newConfig, "\n"))
				nixBlockFound = true
			}
			inNixBlock = true
			continue
		}
		if inNixBlock {
			if strings.TrimSpace(line) == "" {
				inNixBlock = false
			}
			continue
		}
		newLines = append(newLines, line)
	}

	if !nixBlockFound {
		newLines = append(newLines, newConfig)
	}

	return strings.Join(newLines, "\n")
}

// GetDefaultShell returns the default shell for the current platform.
func (m *Manager) GetDefaultShell() string {
	return platform.GetDefaultShell()
}

/*
IsValidShell checks if the given shell is supported by Nix Foundry.
Currently supports bash, zsh, and fish shells.
*/
func (m *Manager) IsValidShell(shell string) bool {
	validShells := []string{"/bin/bash", "/bin/zsh", "/bin/fish"}
	for _, s := range validShells {
		if s == shell {
			return true
		}
	}
	return false
}

// GetShellConfigFile returns the configuration file path for the given shell.
func (m *Manager) GetShellConfigFile(shell string) (string, error) {
	return platform.GetShellConfigFile(shell)
}

/*
BackupShellConfig creates a backup of the shell configuration file.
The backup is created with a .backup extension in the same directory.
*/
func (m *Manager) BackupShellConfig(shell string) error {
	configFile, err := platform.GetShellConfigFile(shell)
	if err != nil {
		return fmt.Errorf("failed to get shell config file: %w", err)
	}

	if !m.fs.Exists(configFile) {
		return nil
	}

	backupFile := configFile + ".backup"
	return m.fs.Copy(configFile, backupFile)
}

/*
RestoreShellConfig restores the shell configuration from a backup file.
Returns an error if the backup file doesn't exist or if restoration fails.
*/
func (m *Manager) RestoreShellConfig(shell string) error {
	configFile, err := platform.GetShellConfigFile(shell)
	if err != nil {
		return fmt.Errorf("failed to get shell config file: %w", err)
	}

	backupFile := configFile + ".backup"
	if !m.fs.Exists(backupFile) {
		return fmt.Errorf("backup file not found: %s", backupFile)
	}

	if err := m.fs.Remove(configFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove current config: %w", err)
	}

	return m.fs.Copy(backupFile, configFile)
}
