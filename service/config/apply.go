// Package config provides configuration management functionality.
package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/shawnkhoffman/nix-foundry/pkg/schema"
)

// ApplyService handles configuration application operations.
type ApplyService struct {
	fs filesystem.FileSystem
}

// NewApplyService creates a new configuration apply service.
func NewApplyService(fs filesystem.FileSystem) *ApplyService {
	return &ApplyService{fs: fs}
}

// Apply applies the configuration by installing packages and running scripts.
func (s *ApplyService) Apply(config *schema.Config) error {
	if err := s.applyPackages(config); err != nil {
		return fmt.Errorf("failed to apply packages: %w", err)
	}

	if err := s.applyScripts(config); err != nil {
		return fmt.Errorf("failed to apply scripts: %w", err)
	}

	return nil
}

// applyPackages installs packages specified in the configuration.
func (s *ApplyService) applyPackages(config *schema.Config) error {
	tmpDir, err := os.MkdirTemp("", "nix-foundry-packages-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	scriptPath := filepath.Join(tmpDir, "install.sh")
	var script string

	for _, pkg := range config.Nix.Packages.Core {
		script += fmt.Sprintf("nix-env -iA nixpkgs.%s\n", pkg)
	}
	for _, pkg := range config.Nix.Packages.Optional {
		script += fmt.Sprintf("nix-env -iA nixpkgs.%s\n", pkg)
	}

	if err := s.fs.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return fmt.Errorf("failed to write package installation script: %w", err)
	}

	cmd := exec.Command(config.Settings.Shell, scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install packages: %w", err)
	}

	return nil
}

// applyScripts runs scripts specified in the configuration.
func (s *ApplyService) applyScripts(config *schema.Config) error {
	for _, script := range config.Nix.Scripts {
		tmpDir, err := os.MkdirTemp("", "nix-foundry-script-*")
		if err != nil {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		scriptPath := filepath.Join(tmpDir, "script.sh")
		if err := s.fs.WriteFile(scriptPath, []byte(script.Commands), 0755); err != nil {
			return fmt.Errorf("failed to write script file: %w", err)
		}

		cmd := exec.Command(config.Settings.Shell, scriptPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run script '%s': %w", script.Name, err)
		}
	}

	return nil
}
