package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/config"
	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/platform"
	"github.com/shawnkhoffman/nix-foundry/cmd/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
)

type (
	PlatformError struct {
		Err error
	}
	NixError struct {
		Err error
	}
	ConfigError struct {
		Err error
	}
)

func (e PlatformError) Error() string {
	return fmt.Sprintf("Platform setup failed: %v\nPlease ensure your system meets the requirements.", e.Err)
}

func (e NixError) Error() string {
	return fmt.Sprintf("Nix setup failed: %v\nTry running 'curl -L https://nixos.org/nix/install | sh' manually.", e.Err)
}

func (e ConfigError) Error() string {
	return fmt.Sprintf("Configuration failed: %v\nTry removing ~/.config/nix-foundry and running init again.", e.Err)
}

func setupNixFoundry(sys *platform.System, configDir string) error {
	// Create config directory and required subdirectories
	dirs := []string{
		configDir,
		filepath.Join(configDir, "environments"),
		filepath.Join(configDir, "backups"),
		filepath.Join(configDir, "logs"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Platform-specific setup
	if sys.OS == "darwin" {
		// Ensure Homebrew is installed on macOS
		if _, err := exec.LookPath("brew"); err != nil {
			spin := progress.NewSpinner("Installing Homebrew...")
			spin.Start()
			if err := installHomebrew(); err != nil {
				spin.Fail("Failed to install Homebrew")
				return PlatformError{Err: err}
			}
			spin.Success("Homebrew installed")
		}
	}

	// Install home-manager if not present
	if _, err := exec.LookPath("home-manager"); err != nil {
		spin := progress.NewSpinner("Installing home-manager...")
		spin.Start()
		if err := installHomeManager(); err != nil {
			spin.Fail("Failed to install home-manager")
			return NixError{Err: err}
		}
		spin.Success("home-manager installed")
	}

	// Initialize environment isolation
	if err := setupEnvironmentIsolation(); err != nil {
		return fmt.Errorf("failed to setup environment isolation: %w", err)
	}

	return nil
}

func installHomebrew() error {
	cmd := exec.Command("/bin/bash", "-c",
		`$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)`)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("homebrew installation failed: %w", err)
	}
	return nil
}

func installHomeManager() error {
	cmd := exec.Command("nix-channel", "--add", "https://github.com/nix-community/home-manager/archive/master.tar.gz", "home-manager")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("nix-channel", "--update")
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("nix-shell", "<home-manager>", "-A", "installer", "--run", "home-manager init")
	return cmd.Run()
}

func previewConfiguration(nixCfg *config.NixConfig) {
	fmt.Println("\n📋 Configuration Summary")
	fmt.Println("---------------------")

	// Core Configuration
	fmt.Printf("🖥️  System:     %s (%s)\n", nixCfg.Platform.OS, nixCfg.Platform.Arch)
	fmt.Printf("🐚 Shell:      %s\n", nixCfg.Shell.Type)
	fmt.Printf("📝 Editor:     %s\n", nixCfg.Editor.Type)

	// Git (only if configured)
	if nixCfg.Git.User.Name != "" || nixCfg.Git.User.Email != "" {
		fmt.Printf("🔀 Git:        %s", nixCfg.Git.User.Name)
		if nixCfg.Git.User.Email != "" {
			fmt.Printf(" <%s>", nixCfg.Git.User.Email)
		}
		fmt.Println()
	}

	// Development Tools (only if any are configured)
	var tools []string
	if v := nixCfg.Development.Languages.Go.Version; v != "" {
		tools = append(tools, fmt.Sprintf("Go %s", v))
	}
	if v := nixCfg.Development.Languages.Node.Version; v != "" {
		tools = append(tools, fmt.Sprintf("Node %s", v))
	}
	if v := nixCfg.Development.Languages.Python.Version; v != "" {
		tools = append(tools, fmt.Sprintf("Python %s", v))
	}
	if len(tools) > 0 {
		fmt.Printf("🛠️  Tools:      %s\n", strings.Join(tools, ", "))
	}

	// Shell plugins (only if any are configured)
	if len(nixCfg.Shell.Plugins) > 0 {
		fmt.Printf("🔌 Plugins:    %s\n", strings.Join(nixCfg.Shell.Plugins, ", "))
	}

	fmt.Println()
}

var initCmd = &cobra.Command{
	Use:   "init [config.yaml]",
	Short: "Initialize a new development environment",
	Long: `Initialize a new development environment using a YAML configuration file.
If no configuration file is provided and --auto is set, a default configuration will be generated.

Examples:
  # Generate default configuration automatically
  nix-foundry init --auto

  # Use custom configuration file
  nix-foundry init my-config.yaml

  # Customize shell and editor
  nix-foundry init --auto --shell zsh --editor nvim

  # Configure git settings
  nix-foundry init --auto --git-name "Your Name" --git-email "you@example.com"

Configuration will be stored in ~/.config/nix-foundry/`,
	RunE: func(cmd *cobra.Command, args []string) error {
		home, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return fmt.Errorf("failed to get home directory: %w", homeErr)
		}
		configDir := filepath.Join(home, ".config", "nix-foundry")

		sys, detectErr := platform.Detect()
		if detectErr != nil {
			return fmt.Errorf("platform detection failed: %w", detectErr)
		}

		// Check if configuration file exists (not just the directory)
		configFile := filepath.Join(configDir, "config.yaml")
		configExists := false
		if _, err := os.Stat(configFile); err == nil {
			configExists = true
		}

		// Handle configuration
		var yamlConfig []byte
		var nixCfg *config.NixConfig
		if len(args) > 0 {
			configFile = args[0]
			var err error
			yamlConfig, err = os.ReadFile(configFile)
			if err != nil {
				return fmt.Errorf("failed to read config file: %w", err)
			}
			nixCfg, err = config.ConvertYAMLToNix(yamlConfig)
			if err != nil {
				return fmt.Errorf("failed to convert config: %w", err)
			}
			nixCfg.Platform = config.PlatformConfig{
				OS:   sys.OS,
				Arch: sys.Arch,
			}
		} else if autoConfig {
			var err error
			yamlConfig, err = config.GenerateDefaultConfig(*sys)
			if err != nil {
				return fmt.Errorf("failed to generate default config: %w", err)
			}
			nixCfg, err = config.ConvertYAMLToNix(yamlConfig)
			if err != nil {
				return fmt.Errorf("failed to convert config: %w", err)
			}
			nixCfg.Platform = config.PlatformConfig{
				OS:   sys.OS,
				Arch: sys.Arch,
			}
		} else {
			return fmt.Errorf("no configuration file provided. Use --auto to generate a default configuration")
		}

		// Create config directory for the config file
		if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		// Write the configuration file
		if err := os.WriteFile(configFile, yamlConfig, 0644); err != nil {
			return fmt.Errorf("failed to write configuration file: %w", err)
		}

		if configExists {
			fmt.Println("\n⚠️  Warning: This will overwrite your existing configuration!")
			fmt.Println("\nThe following changes will be made:")
			fmt.Printf("1. Update configuration in: %s\n", configDir)
			fmt.Println("2. Regenerate home-manager configuration")
			fmt.Println("3. Update development environment")
		} else {
			fmt.Println("👋 Welcome to nix-foundry! Let's get you set up.")
			fmt.Println("\nThe following will be configured:")
			fmt.Printf("1. Create configuration in: %s\n", configDir)
			fmt.Printf("2. Configure shell: %s\n", nixCfg.Shell.Type)
			fmt.Printf("3. Configure editor: %s\n", nixCfg.Editor.Type)
			if gitName != "" || gitEmail != "" {
				fmt.Println("4. Set up Git configuration")
			}
		}

		// Show configuration preview
		previewConfiguration(nixCfg)

		fmt.Println()

		// Ask for confirmation
		fmt.Print("Would you like to apply this configuration? [y/N]: ")
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			return fmt.Errorf("failed to read user input: %w", err)
		}

		if !strings.EqualFold(response, "y") && !strings.EqualFold(response, "yes") {
			fmt.Println("Initialization cancelled.")
			return nil
		}

		// After user confirmation
		spin := progress.NewSpinner("Setting up nix-foundry...")
		spin.Start()
		if err := setupNixFoundry(sys, configDir); err != nil {
			spin.Fail("Setup failed")
			return err
		}
		spin.Success("Setup complete")

		// Convert YAML to internal config
		spin = progress.NewSpinner("Processing configuration...")
		spin.Start()
		defaultEnv := filepath.Join(configDir, "environments", "default")
		if err := os.MkdirAll(defaultEnv, 0755); err != nil {
			spin.Fail("Failed to create default environment")
			return fmt.Errorf("failed to create default environment: %w", err)
		}
		// Generate files directly in the default environment
		if err := config.Generate(defaultEnv, nixCfg); err != nil {
			spin.Fail("Configuration generation failed")
			return fmt.Errorf("failed to generate Nix configuration: %w", err)
		}
		spin.Success("Configuration processed")

		// Apply configuration (this will create flake.lock)
		spin = progress.NewSpinner("Applying configuration...")
		spin.Start()
		if err := config.Apply(configDir); err != nil {
			spin.Fail("Configuration application failed")
			return fmt.Errorf("failed to apply configuration: %w", err)
		}
		spin.Success("Configuration applied successfully")

		// After successfully applying configuration...
		spin = progress.NewSpinner("Setting up environment...")
		spin.Start()

		// Create symlink to default environment
		currentEnv := filepath.Join(configDir, "environments", config.CurrentEnv)
		defaultEnv = filepath.Join(configDir, "environments", config.DefaultEnv)

		// Remove existing symlink if it exists
		if _, err := os.Lstat(currentEnv); err == nil {
			if err := os.Remove(currentEnv); err != nil {
				spin.Fail("Failed to update environment link")
				return fmt.Errorf("failed to remove existing environment link: %w", err)
			}
		}

		// Create new symlink
		if err := os.Symlink(defaultEnv, currentEnv); err != nil {
			spin.Fail("Failed to set up environment")
			return fmt.Errorf("failed to create environment link: %w", err)
		}
		spin.Success("Environment ready")

		// After successful configuration application
		spin = progress.NewSpinner("Installing nix-foundry...")
		spin.Start()
		binDir := filepath.Join(os.Getenv("HOME"), ".local", "bin")
		if err := os.MkdirAll(binDir, 0755); err != nil {
			spin.Fail("Failed to create bin directory")
			return fmt.Errorf("failed to create bin directory: %w", err)
		}

		executable, err := os.Executable()
		if err != nil {
			spin.Fail("Failed to get executable path")
			return fmt.Errorf("failed to get executable path: %w", err)
		}

		symlink := filepath.Join(binDir, "nix-foundry")
		if err := os.Symlink(executable, symlink); err != nil && !os.IsExist(err) {
			spin.Fail("Failed to create symlink")
			return fmt.Errorf("failed to create symlink: %w", err)
		}
		spin.Success("nix-foundry installed")

		fmt.Printf("\nℹ️  Add %s to your PATH to use nix-foundry from anywhere\n", binDir)

		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&shell, "shell", "zsh", fmt.Sprintf("Shell to configure %v", validShells))
	initCmd.Flags().StringVar(&editor, "editor", "nano", fmt.Sprintf("Text editor %v", validEditors))
	initCmd.Flags().StringVar(&gitName, "git-name", "", "Git user name")
	initCmd.Flags().StringVar(&gitEmail, "git-email", "", "Git user email")
	initCmd.Flags().StringVar(&teamName, "team", "", "Team configuration to use")
	initCmd.Flags().BoolVar(&autoConfig, "auto", false, "Generate minimal configuration if none exists")
	initCmd.Flags().BoolVar(&projectInit, "project", false, "Initialize a project environment")
	initCmd.Flags().BoolVar(&forceInit, "force", false, "Force initialization even if configuration exists")
}
