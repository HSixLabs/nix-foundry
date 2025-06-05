/*
Package script provides functionality for managing shell scripts.
It handles script operations including adding, removing, and executing scripts
within the Nix Foundry configuration system.
*/
package script

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/shawnkhoffman/nix-foundry/pkg/schema"
)

/*
Manager handles script operations.
It provides functionality for managing and executing shell scripts
using the provided filesystem abstraction.
*/
type Manager struct {
	fs filesystem.FileSystem
}

/*
NewManager creates a new script manager instance with the provided filesystem.
*/
func NewManager(fs filesystem.FileSystem) *Manager {
	return &Manager{fs: fs}
}

/*
AddScript adds a script to the configuration.
It validates the script name for uniqueness and adds it to the configuration's
script list.
*/
func (m *Manager) AddScript(script schema.Script, config *schema.Config) error {
	for _, s := range config.Nix.Scripts {
		if s.Name == script.Name {
			return fmt.Errorf("script %s already exists", script.Name)
		}
	}

	config.Nix.Scripts = append(config.Nix.Scripts, script)
	return nil
}

/*
RemoveScript removes a script from the configuration.
It searches for the script by name and removes it from the configuration's
script list if found.
*/
func (m *Manager) RemoveScript(name string, config *schema.Config) error {
	for i, s := range config.Nix.Scripts {
		if s.Name == name {
			config.Nix.Scripts = append(config.Nix.Scripts[:i], config.Nix.Scripts[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("script %s not found", name)
}

/*
ListScripts returns all scripts from the configuration.
*/
func (m *Manager) ListScripts(config *schema.Config) []schema.Script {
	return config.Nix.Scripts
}

/*
RunScript executes a script from the configuration.
It creates a temporary script file with the script content and executes it
using the configured shell.
*/
func (m *Manager) RunScript(name string, config *schema.Config) error {
	var script schema.Script
	for _, s := range config.Nix.Scripts {
		if s.Name == name {
			script = s
			break
		}
	}
	if script.Name == "" {
		return fmt.Errorf("script %s not found", name)
	}

	tmpDir, err := os.MkdirTemp("", "nix-foundry-script-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	scriptPath := filepath.Join(tmpDir, script.Name)
	if err := m.fs.WriteFile(scriptPath, []byte(script.Commands), 0755); err != nil {
		return fmt.Errorf("failed to write script file: %w", err)
	}

	shell := config.Settings.Shell
	if shell == "" {
		shell = "bash"
	}

	cmd := exec.Command(shell, scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run script: %w", err)
	}

	return nil
}
