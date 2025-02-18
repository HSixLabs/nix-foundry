package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	configservice "github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config/defaults"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/spf13/cobra"
)

func NewInitCommand(cfgSvc configservice.Service) *cobra.Command {
	var (
		forceConfig bool
		autoConfig  bool
		testMode    bool
		shell       string
		editor      string
		gitName     string
		gitEmail    string
		projectInit bool
		teamName    string
		autoInstall bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := logging.GetLogger()

			// Get config directory
			home, err := os.UserHomeDir()
			if err != nil {
				logger.WithError(err).Error("Failed to get home directory")
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			configDir := filepath.Join(home, ".config", "nix-foundry")

			// Add validation checks first
			validShells := map[string]bool{"zsh": true, "bash": true, "fish": true}
			if !validShells[shell] {
				return fmt.Errorf("invalid shell: %s. Valid options: zsh, bash, fish", shell)
			}

			validEditors := map[string]bool{"nano": true, "vim": true, "emacs": true, "neovim": true, "vscode": true}
			if !validEditors[editor] {
				return fmt.Errorf("invalid editor: %s. Valid options: nano, vim, emacs, neovim, vscode", editor)
			}

			// Initialize config service with test mode
			if err := cfgSvc.Initialize(testMode); err != nil {
				return fmt.Errorf("failed to initialize config service: %w", err)
			}

			// Create configuration map with validated values
			configMap := map[string]string{
				"shell.type":  shell,
				"editor.type": editor,
			}
			if gitName != "" {
				configMap["git.name"] = gitName
			}
			if gitEmail != "" {
				configMap["git.email"] = gitEmail
			}

			// Apply the configuration values
			if err := cfgSvc.ApplyFlags(configMap, true); err != nil {
				return fmt.Errorf("failed to apply configuration values: %w", err)
			}

			// Generate configuration with validated values
			cfg := defaults.New()

			// If you need to access NixConfig fields:
			cfg.NixConfig.Settings.LogLevel = "info"

			// Check if config exists
			configExists := cfgSvc.ConfigExists()

			// Show configuration preview
			if configExists {
				fmt.Println("\n⚠️  Warning: This will overwrite your existing configuration!")
				fmt.Println("\nThe following changes will be made:")
				fmt.Printf("1. Update configuration in: %s\n", cfgSvc.GetConfigDir())
				fmt.Println("2. Regenerate home-manager configuration")
				fmt.Println("3. Update development environment")
			} else {
				fmt.Println("\nThe following will be configured:")
				fmt.Printf("1. Create configuration in: %s\n", cfgSvc.GetConfigDir())
				fmt.Printf("2. Configure shell: %s\n", shell)
				fmt.Printf("3. Configure editor: %s\n", editor)
				if gitName != "" || gitEmail != "" {
					fmt.Println("4. Set up Git configuration")
				}
			}

			// Show configuration preview
			if err := cfgSvc.PreviewConfiguration(cfg); err != nil {
				return fmt.Errorf("failed to preview configuration: %w", err)
			}

			// Interactive confirmation
			if !autoConfig && !testMode {
				if !confirmApply() {
					fmt.Println("Initialization cancelled.")
					return nil
				}
			}

			// Apply the configuration
			if err := cfgSvc.Apply(cfg.GetNixConfig(), testMode); err != nil {
				return fmt.Errorf("failed to apply configuration: %w", err)
			}

			// Save the configuration explicitly
			if err := cfgSvc.Save(cfg); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			// Initialize environment service
			envService := environment.NewService(
				configDir,
				cfgSvc,
				platform.NewService(),
				testMode,
				true,
				autoInstall,
			)

			// Initialize the environment
			if err := envService.Initialize(testMode); err != nil {
				return fmt.Errorf("failed to initialize environment: %w", err)
			}

			// Apply the configuration through home-manager
			if err := envService.ApplyConfiguration(); err != nil {
				return fmt.Errorf("failed to activate configuration: %w", err)
			}

			fmt.Println("\n✅ Configuration initialized successfully")
			fmt.Println("┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓")
			fmt.Println("┃         Configuration Summary       ┃")
			fmt.Println("┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫")
			fmt.Printf("┃ %-20s %-18s ┃\n", "Config Directory:", cfgSvc.GetConfigDir())
			fmt.Printf("┃ %-20s %-18s ┃\n", "Shell:", shell)
			fmt.Printf("┃ %-20s %-18s ┃\n", "Editor:", editor)
			if gitName != "" || gitEmail != "" {
				fmt.Printf("┃ %-20s %-18s ┃\n", "Git Name:", gitName)
				fmt.Printf("┃ %-20s %-18s ┃\n", "Git Email:", gitEmail)
			}
			fmt.Println("┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛")

			// Add next steps guidance
			fmt.Println("\nNext steps:")
			fmt.Println(" 1. Review your configuration:")
			fmt.Println("    cat", filepath.Join(cfgSvc.GetConfigDir(), "config.yaml"))
			fmt.Println(" 2. Add packages:")
			fmt.Println("    nix-foundry packages add [package-name]")

			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&forceConfig, "force", false, "Force initialization even if config exists")
	cmd.Flags().BoolVar(&autoConfig, "auto", false, "Automatically generate default configuration")
	cmd.Flags().BoolVar(&testMode, "test", false, "Run in test mode")
	cmd.Flags().StringVar(&shell, "shell", "zsh", "Shell to configure [zsh bash]")
	cmd.Flags().StringVar(&editor, "editor", "nano", "Text editor [nano vim nvim emacs neovim vscode]")
	cmd.Flags().StringVar(&gitName, "git-name", "", "Git user name")
	cmd.Flags().StringVar(&gitEmail, "git-email", "", "Git user email")
	cmd.Flags().BoolVar(&projectInit, "project", false, "Initialize a project environment")
	cmd.Flags().StringVar(&teamName, "team", "", "Team configuration to use")
	cmd.Flags().BoolVar(&autoInstall, "auto-install", false, "Automatically install dependencies without prompting")

	return cmd
}

func confirmApply() bool {
	fmt.Print("\nWould you like to apply this configuration? [y/N]: ")
	var confirm string
	if _, err := fmt.Scanln(&confirm); err != nil {
		return false
	}
	return strings.EqualFold(confirm, "y") || strings.EqualFold(confirm, "yes")
}
