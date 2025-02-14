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
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
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
	return fmt.Sprintf("platform setup failed: %v\nplease ensure your system meets the requirements", e.Err)
}

func (e NixError) Error() string {
	return fmt.Sprintf("nix setup failed: %v\ntry running 'curl -L https://nixos.org/nix/install | sh' manually", e.Err)
}

func (e ConfigError) Error() string {
	return fmt.Sprintf("configuration failed: %v\ntry removing ~/.config/nix-foundry and running init again", e.Err)
}

func setupNixFoundry(sys *platform.System, configDir string) error {
	// Create config directory and required subdirectories first
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

	// Skip remaining setup in test mode
	if testMode {
		// Create dummy files that would normally be created by nix setup
		files := []string{
			filepath.Join(configDir, "flake.nix"),
			filepath.Join(configDir, "home.nix"),
		}
		for _, file := range files {
			if err := os.WriteFile(file, []byte("test configuration"), 0644); err != nil {
				return fmt.Errorf("failed to create test file %s: %w", file, err)
			}
		}
		return nil
	}

	// Check if Nix is installed
	if _, err := exec.LookPath("nix-channel"); err != nil {
		return fmt.Errorf("nix setup failed: %w\ntry running 'curl -L https://nixos.org/nix/install | sh' manually", err)
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
		fmt.Printf("🔧  Tools:      %s\n", strings.Join(tools, ", "))
	}

	// Shell plugins (only if any are configured)
	if len(nixCfg.Shell.Plugins) > 0 {
		fmt.Printf("🔌 Plugins:    %s\n", strings.Join(nixCfg.Shell.Plugins, ", "))
	}

	fmt.Println()
}

func setupEnvironmentIsolation() error {
	// Skip environment isolation setup in test mode
	if testMode {
		return nil
	}

	// Check if home-manager is installed
	if _, err := exec.LookPath("home-manager"); err != nil {
		return fmt.Errorf("home-manager not found: please install it first using 'nix-env -iA nixpkgs.home-manager'")
	}

	return nil
}

var initCmd = &cobra.Command{
	Use:   "init",
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
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate requirements before any execution
		if len(args) == 0 && !autoConfig {
			return fmt.Errorf("either --auto flag or config file path is required")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("👋 Welcome to nix-foundry! Let's get you set up.")

		// Rest of the validation and initialization logic...
		if shell != "" {
			if err := config.ValidateAutoShell(shell); err != nil {
				return err
			}
		}

		if editor != "" {
			if err := config.ValidateEditor(editor); err != nil {
				return err
			}
		}

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

		// If auto flag is set, generate default config
		if autoConfig {
			// Create temporary config for validation
			tempConfig := &config.NixConfig{
				Version: "1.0",
				Shell: config.ShellConfig{
					Type: shell,
				},
				Editor: config.EditorConfig{
					Type: editor,
				},
			}
			validator := config.NewValidator(tempConfig)
			if err := validator.ValidateConfig(); err != nil {
				return err
			}

			nixCfg = &config.NixConfig{
				Version: "1.0",
				Shell: config.ShellConfig{
					Type:    shell,
					Plugins: []string{"zsh-autosuggestions"},
				},
				Editor: config.EditorConfig{
					Type:       editor,
					Extensions: make([]string, 0),
				},
				Git: config.GitConfig{
					Enable: true,
					User: struct {
						Name  string `yaml:"name"`
						Email string `yaml:"email"`
					}{
						Name:  gitName,
						Email: gitEmail,
					},
				},
				Platform: config.PlatformConfig{
					OS:   sys.OS,
					Arch: sys.Arch,
				},
				Packages: config.PackagesConfig{
					Additional: []string{"git", "ripgrep"},
					PlatformSpecific: map[string][]string{
						"darwin": {"mas"},
						"linux":  {"inotify-tools"},
					},
					Development: []string{"git", "ripgrep", "fd", "jq"},
					Team:        make(map[string][]string),
				},
				Team: config.TeamConfig{
					Enable:   false,
					Name:     teamName,
					Settings: make(map[string]string),
				},
				Development: config.DevelopmentConfig{
					Languages: struct {
						Go struct {
							Version  string   `yaml:"version,omitempty"`
							Packages []string `yaml:"packages,omitempty"`
						} `yaml:"go"`
						Node struct {
							Version  string   `yaml:"version"`
							Packages []string `yaml:"packages,omitempty"`
						} `yaml:"node"`
						Python struct {
							Version  string   `yaml:"version"`
							Packages []string `yaml:"packages,omitempty"`
						} `yaml:"python"`
					}{},
					Tools: []string{},
				},
			}
		} else if len(args) > 0 {
			// Loading from config file
			configPath := args[0]
			nixCfg = &config.NixConfig{}
			cfg, err := configManager.LoadConfig(config.PersonalConfigType, configPath)
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Type assert the returned config
			var ok bool
			nixCfg, ok = cfg.(*config.NixConfig)
			if !ok {
				return fmt.Errorf("invalid configuration type returned")
			}

			// Use regular validator for config files
			validator := config.NewValidator(nixCfg)
			if err := validator.ValidateConfig(); err != nil {
				return fmt.Errorf("invalid configuration: %w", err)
			}

			// Move platform config update after validation
			nixCfg.Platform = config.PlatformConfig{
				OS:   sys.OS,
				Arch: sys.Arch,
			}
		} else {
			return fmt.Errorf("either --auto flag or config file path is required")
		}

		// Create config directory for the config file
		if mkdirErr := os.MkdirAll(filepath.Dir(configFile), 0755); mkdirErr != nil {
			return fmt.Errorf("failed to create config directory: %w", mkdirErr)
		}

		var err error
		// Marshal config to YAML
		yamlConfig, err = yaml.Marshal(nixCfg)
		if err != nil {
			return fmt.Errorf("failed to marshal configuration: %w", err)
		}

		// Write the configuration file
		if writeErr := os.WriteFile(configFile, yamlConfig, 0644); writeErr != nil {
			return fmt.Errorf("failed to write configuration file: %w", writeErr)
		}

		if configExists {
			fmt.Println("\n⚠️  Warning: This will overwrite your existing configuration!")
			fmt.Println("\nThe following changes will be made:")
			fmt.Printf("1. Update configuration in: %s\n", configDir)
			fmt.Println("2. Regenerate home-manager configuration")
			fmt.Println("3. Update development environment")
		} else {
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

		// Interactive confirmation (skip in test mode)
		if !autoConfig && !testMode {
			fmt.Print("\nWould you like to apply this configuration? [y/N]: ")
			var confirm string
			if _, scanErr := fmt.Scanln(&confirm); scanErr != nil {
				return fmt.Errorf("failed to read user input: %w", scanErr)
			}

			if !strings.EqualFold(confirm, "y") && !strings.EqualFold(confirm, "yes") {
				fmt.Println("Initialization cancelled.")
				return nil
			}
		}

		// After user confirmation
		spin := progress.NewSpinner("Setting up nix-foundry...")
		spin.Start()
		if setupErr := setupNixFoundry(sys, configDir); setupErr != nil {
			spin.Fail("Setup failed")
			return setupErr
		}
		spin.Success("Setup complete")

		// Convert YAML to internal config
		spin = progress.NewSpinner("Processing configuration...")
		spin.Start()
		defaultEnv := filepath.Join(configDir, "environments", "default")
		if mkdirErr := os.MkdirAll(defaultEnv, 0755); mkdirErr != nil {
			spin.Fail("Failed to create default environment")
			return fmt.Errorf("failed to create default environment: %w", mkdirErr)
		}
		// Generate files directly in the default environment
		if genErr := config.Generate(defaultEnv, nixCfg); genErr != nil {
			spin.Fail("Configuration generation failed")
			return fmt.Errorf("failed to generate Nix configuration: %w", genErr)
		}
		spin.Success("Configuration processed")

		// Convert flags to configuration map
		configMap := make(map[string]string)
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Name != "test-mode" && f.Name != "auto" { // Skip non-config flags
				if f.Value.String() != "" {
					configMap[f.Name] = f.Value.String()
				}
			}
		})

		// Apply configuration
		spin = progress.NewSpinner("Applying configuration...")
		spin.Start()
		// Initialize config manager if not already initialized
		if configManager == nil {
			var initErr error
			configManager, initErr = config.NewConfigManager()
			if initErr != nil {
				return fmt.Errorf("failed to initialize config manager: %w", initErr)
			}
		}

		// Get config file path
		configFile = filepath.Join(configManager.GetConfigDir(), "config.yaml")
		if configManager.ConfigExists("config.yaml") && !forceConfig {
			return fmt.Errorf("configuration already exists at %s", configFile)
		}

		// Create config directory for the config file
		if mkdirErr := os.MkdirAll(filepath.Dir(configFile), 0755); mkdirErr != nil {
			return fmt.Errorf("failed to create config directory: %w", mkdirErr)
		}

		if applyErr := configManager.Apply(configMap, testMode); applyErr != nil {
			spin.Fail("Failed to apply configuration")
			return fmt.Errorf("failed to apply configuration: %w", applyErr)
		}
		spin.Success("Configuration applied successfully")

		// After successfully applying configuration...
		spin = progress.NewSpinner("Setting up environment...")
		spin.Start()

		// Create symlink to default environment
		currentEnv := filepath.Join(configDir, "environments", "current")
		defaultEnv = filepath.Join(configDir, "environments", "default")

		// Remove existing symlink if it exists
		if _, lstatErr := os.Lstat(currentEnv); lstatErr == nil {
			if removeErr := os.Remove(currentEnv); removeErr != nil {
				spin.Fail("Failed to update environment link")
				return fmt.Errorf("failed to remove existing environment link: %w", removeErr)
			}
		}

		// Create new symlink
		if symlinkErr := os.Symlink(defaultEnv, currentEnv); symlinkErr != nil {
			spin.Fail("Failed to set up environment")
			return fmt.Errorf("failed to create environment link: %w", symlinkErr)
		}
		spin.Success("Environment ready")

		// After successful configuration application
		spin = progress.NewSpinner("Installing nix-foundry...")
		spin.Start()
		binDir := filepath.Join(os.Getenv("HOME"), ".local", "bin")
		if mkdirErr := os.MkdirAll(binDir, 0755); mkdirErr != nil {
			spin.Fail("Failed to create bin directory")
			return fmt.Errorf("failed to create bin directory: %w", mkdirErr)
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
	initCmd.Flags().StringVar(&shell, "shell", "zsh", "Shell to configure [zsh bash]")
	initCmd.Flags().StringVar(&editor, "editor", "nano", "Text editor [nano vim nvim emacs neovim vscode]")
	initCmd.Flags().StringVar(&gitName, "git-name", "", "Git user name")
	initCmd.Flags().StringVar(&gitEmail, "git-email", "", "Git user email")
	initCmd.Flags().BoolVar(&autoConfig, "auto", false, "Generate minimal configuration if none exists")
	initCmd.Flags().BoolVar(&projectInit, "project", false, "Initialize a project environment")
	initCmd.Flags().StringVar(&teamName, "team", "", "Team configuration to use")
	initCmd.Flags().BoolVar(&forceConfig, "force", false, "Force initialization even if configuration exists")
	initCmd.Flags().BoolVar(&testMode, "test-mode", false, "Run in test mode")
}
