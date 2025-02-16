package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
)

func NewInitCommand() *cobra.Command {
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
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get config directory
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			configDir := filepath.Join(home, ".config", "nix-foundry")

			// Initialize environment service
			envService := environment.NewService(
				configDir,
				config.NewService(),
				nil, // validation service (placeholder)
				nil, // platform service (placeholder)
			)
			if cfgErr := envService.Initialize(testMode); cfgErr != nil {
				return fmt.Errorf("failed to initialize environment: %w", cfgErr)
			}

			// Initialize configuration service
			configService := config.NewService()

			// Check if config exists
			configExists := configService.ConfigExists()

			// Generate initial configuration
			nixConfig, err := configService.GenerateInitialConfig(shell, editor, gitName, gitEmail)
			if err != nil {
				return fmt.Errorf("failed to generate configuration: %w", err)
			}

			// Show configuration preview
			if configExists {
				fmt.Println("\n⚠️  Warning: This will overwrite your existing configuration!")
				fmt.Println("\nThe following changes will be made:")
				fmt.Printf("1. Update configuration in: %s\n", configService.GetConfigDir())
				fmt.Println("2. Regenerate home-manager configuration")
				fmt.Println("3. Update development environment")
			} else {
				fmt.Println("\nThe following will be configured:")
				fmt.Printf("1. Create configuration in: %s\n", configService.GetConfigDir())
				fmt.Printf("2. Configure shell: %s\n", shell)
				fmt.Printf("3. Configure editor: %s\n", editor)
				if gitName != "" || gitEmail != "" {
					fmt.Println("4. Set up Git configuration")
				}
			}

			// Show configuration preview
			if err := configService.PreviewConfiguration(nixConfig); err != nil {
				return fmt.Errorf("failed to preview configuration: %w", err)
			}

			// Interactive confirmation
			if !autoConfig && !testMode {
				if !confirmApply() {
					fmt.Println("Initialization cancelled.")
					return nil
				}
			}

			// Apply configuration
			spin := progress.NewSpinner("Applying configuration...")
			spin.Start()
			defer spin.Stop()

			if err := configService.Apply(nixConfig, testMode); err != nil {
				spin.Fail("Failed to apply configuration")
				return fmt.Errorf("failed to apply configuration: %w", err)
			}
			spin.Success("Configuration applied")

			fmt.Println("✅ Configuration initialized successfully")
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

func newInitCommand(cfgSvc config.Service) *cobra.Command {
	var testMode bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfgSvc.Initialize(testMode); err != nil {
				return fmt.Errorf("failed to initialize configuration: %w", err)
			}
			fmt.Println("✅ Configuration initialized")
			return nil
		},
	}

	cmd.Flags().BoolVar(&testMode, "test", false, "Initialize with test configuration")
	return cmd
}
