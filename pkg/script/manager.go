// Package script provides functionality for managing shell scripts.
package script

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/shawnkhoffman/nix-foundry/pkg/schema"
)

// Manager handles script operations.
type Manager struct {
	fs filesystem.FileSystem
}

// NewManager creates a new script manager instance.
func NewManager(fs filesystem.FileSystem) *Manager {
	return &Manager{fs: fs}
}

// AddScript adds a script to the configuration.
func (m *Manager) AddScript(script schema.Script, config *schema.Config) error {
	// Check for duplicate script names
	for _, s := range config.Nix.Scripts {
		if s.Name == script.Name {
			return fmt.Errorf("script with name '%s' already exists", script.Name)
		}
	}

	config.Nix.Scripts = append(config.Nix.Scripts, script)
	return nil
}

// RemoveScript removes a script from the configuration.
func (m *Manager) RemoveScript(name string, config *schema.Config) error {
	found := false
	var scripts []schema.Script
	for _, s := range config.Nix.Scripts {
		if s.Name != name {
			scripts = append(scripts, s)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("script '%s' not found", name)
	}

	config.Nix.Scripts = scripts
	return nil
}

// ListScripts returns all scripts from the configuration.
func (m *Manager) ListScripts(config *schema.Config) []schema.Script {
	return config.Nix.Scripts
}

// RunScript executes a script from the configuration.
func (m *Manager) RunScript(name string, config *schema.Config) error {
	var script *schema.Script
	for _, s := range config.Nix.Scripts {
		if s.Name == name {
			script = &s
			break
		}
	}

	if script == nil {
		return fmt.Errorf("script '%s' not found", name)
	}

	// Create a temporary script file
	tmpDir, err := os.MkdirTemp("", "nix-foundry-script-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	scriptPath := filepath.Join(tmpDir, "script.sh")
	if err := os.WriteFile(scriptPath, []byte(script.Commands), 0755); err != nil {
		return fmt.Errorf("failed to write script file: %w", err)
	}

	// Execute the script using the configured shell
	shell := config.Settings.Shell
	cmd := exec.Command(shell, scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("script execution failed: %w", err)
	}

	return nil
}
