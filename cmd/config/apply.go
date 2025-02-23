// Package config provides configuration management commands for Nix Foundry.
package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/pkg/schema"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	configFile string
	force      bool
)

// NewApplyCmd creates a new apply command.
func NewApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply Nix configuration",
		Long: `Apply Nix configuration from a YAML file.
This command will read the configuration file and apply it to your system.
By default, it uses ~/.config/nix-foundry/config.yaml.`,
		RunE: runApply,
	}

	cmd.Flags().StringVarP(&configFile, "file", "f", "", "Path to configuration file (default: ~/.config/nix-foundry/config.yaml)")
	cmd.Flags().BoolVarP(&force, "force", "", false, "Force apply configuration even if another package manager is installed")
	return cmd
}

func checkPackageManager(name string) bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	nixProfileBin := filepath.Join(homeDir, ".nix-profile", "bin", name)
	if stat, statErr := os.Stat(nixProfileBin); statErr == nil {
		return !stat.IsDir()
	}

	_, lookErr := exec.LookPath(name)
	return lookErr == nil
}

func runApply(cmd *cobra.Command, args []string) error {
	if configFile == "" {
		var err error
		configFile, err = schema.GetConfigPath()
		if err != nil {
			return fmt.Errorf("failed to get default config path: %w", err)
		}
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config schema.Config
	if err := yaml.Unmarshal(content, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := schema.ValidateConfig(&config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	hasNixEnv := checkPackageManager("nix-env")
	if !hasNixEnv {
		return fmt.Errorf("nix-env is not installed. Please install it first")
	}

	if config.Nix.Manager != "nix-env" {
		return fmt.Errorf("unsupported package manager: %s (only nix-env is supported)", config.Nix.Manager)
	}

	return applyNixEnv(&config)
}

func applyNixEnv(config *schema.Config) error {
	for _, pkg := range config.Nix.Packages.Core {
		fmt.Printf("Installing core package: %s\n", pkg)
		cmd := exec.Command("nix-env", "-iA", "nixpkgs."+pkg)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install package %s: %w", pkg, err)
		}
	}

	for _, pkg := range config.Nix.Packages.Optional {
		fmt.Printf("Installing optional package: %s\n", pkg)
		cmd := exec.Command("nix-env", "-iA", "nixpkgs."+pkg)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install package %s: %w", pkg, err)
		}
	}

	for _, script := range config.Nix.Scripts {
		fmt.Printf("Running script: %s\n", script.Name)

		tmpDir, err := os.MkdirTemp("", "nix-foundry-script-*")
		if err != nil {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
		defer os.RemoveAll(tmpDir)

		scriptPath := filepath.Join(tmpDir, "script.sh")
		if err := os.WriteFile(scriptPath, []byte(script.Commands), 0755); err != nil {
			return fmt.Errorf("failed to write script file: %w", err)
		}

		cmd := exec.Command(config.Settings.Shell, scriptPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run script %s: %w", script.Name, err)
		}
	}

	fmt.Println("✨ Configuration applied successfully")
	return nil
}
