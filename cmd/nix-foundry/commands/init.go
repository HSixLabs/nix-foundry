package commands

import (
	"errors"
	"fmt"
	"os"

	customerrors "github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/shawnkhoffman/nix-foundry/pkg/progress"
	"github.com/spf13/cobra"
)

func NewInitCommand() *cobra.Command {
	var (
		shell       string
		editor      string
		gitName     string
		gitEmail    string
		force       bool
		autoConfig  bool
		testMode    bool
		autoInstall bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize nix-foundry configuration",
		Long: `Initialize a new nix-foundry environment and configuration.

This will:
1. Set up the nix-foundry configuration directory
2. Configure your preferred shell and editor
3. Set up Git configuration (if provided)
4. Initialize the Nix environment
5. Enable flake features
6. Install home-manager`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			debug, err := cmd.Root().Flags().GetBool("debug")
			if err != nil {
				return fmt.Errorf("failed to get debug flag: %w", err)
			}
			return logging.InitLogger(debug)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := logging.GetLogger()
			verbose, _ := cmd.Flags().GetBool("verbose")
			// Initialize core services first
			configSvc := config.NewService()
			platformSvc := platform.NewService()

			// Check for Nix installation and handle auto-install
			if !platformSvc.IsNixInstalled() {
				if autoInstall {
					logger.Info("Attempting Nix installation...")
					if err := platformSvc.InstallNix(); err != nil {
						return formatError(fmt.Errorf("nix auto-install failed: %w", err), verbose)
					}
					// Verify installation after attempt
					if !platformSvc.IsNixInstalled() {
						return formatError(
							customerrors.NewValidationError(
								"nix",
								fmt.Errorf("verification failed after installation"),
								"Automatic installation succeeded but Nix not detected.\nPlease restart your shell and try again.",
							),
							verbose,
						)
					}
				} else {
					return formatError(
						customerrors.NewValidationError(
							"nix",
							fmt.Errorf("nix not installed"),
							`Nix package manager is required. Please either:
1. Run with --auto-install flag to attempt automatic installation
2. Install manually using:
   curl -L https://nixos.org/nix/install | sh
   Then restart your shell and try again`,
						),
						verbose,
					)
				}
			}

			// Initialize environment service early
			envSvc := environment.NewService(
				configSvc.GetConfigDir(),
				configSvc,
				platformSvc,
				testMode,
				true,
				autoInstall,
			)

			// Create required directories first
			configDir := configSvc.GetConfigDir()
			if err := os.MkdirAll(configDir, 0755); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}

			spin := progress.NewSpinner("Initializing nix-foundry...")
			spin.Start()
			defer spin.Stop()

			// Full environment initialization including validation
			logger.Info("Initializing core environment")
			if err := envSvc.Initialize(testMode); err != nil {
				spin.Fail("Environment initialization failed")
				return formatError(err, verbose)
			}

			// Enable flake features
			logger.Info("Enabling Nix flake features")
			if err := platformSvc.EnableFlakeFeatures(); err != nil {
				spin.Fail("Failed to enable flake features")
				return formatError(err, verbose)
			}

			// Install home-manager if not present
			logger.Info("Checking home-manager installation")
			if !platformSvc.IsHomeManagerInstalled() {
				if err := platformSvc.InstallHomeManager(); err != nil {
					spin.Fail("Failed to install home-manager")
					return formatError(err, verbose)
				}
			}

			// Generate initial configuration
			logger.Info("Generating configuration")
			nixConfig, err := configSvc.GenerateInitialConfig(shell, editor, gitName, gitEmail)
			if err != nil {
				spin.Fail("Failed to generate configuration")
				return formatError(err, verbose)
			}

			// Show configuration preview
			if !autoConfig {
				spin.Stop()
				fmt.Println("\nConfiguration Preview:")
				fmt.Printf("Shell: %s\n", shell)
				fmt.Printf("Editor: %s\n", editor)
				if gitName != "" || gitEmail != "" {
					fmt.Printf("Git Config:\n  Name: %s\n  Email: %s\n", gitName, gitEmail)
				}

				fmt.Print("\nApply this configuration? [y/N]: ")
				var response string
				if _, err := fmt.Scanln(&response); err != nil {
					response = "n" // Default to no on error
				}
				if response != "y" && response != "Y" {
					fmt.Println("Initialization cancelled")
					return nil
				}
				spin.Start()
			}

			// Apply configuration
			logger.Info("Applying configuration")
			if err := configSvc.Apply(nixConfig.NixConfig, testMode); err != nil {
				spin.Fail("Failed to apply configuration")
				return formatError(err, verbose)
			}

			spin.Success("Initialization complete")
			fmt.Println("\n✅ nix-foundry initialized successfully")
			fmt.Println("\nNext steps:")
			fmt.Println("1. Run 'nix-foundry config packages add' to install packages")
			fmt.Println("2. Run 'nix-foundry project init' to set up a project")
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Add flags
	cmd.Flags().StringVar(&shell, "shell", "zsh", "Preferred shell [zsh, bash]")
	cmd.Flags().StringVar(&editor, "editor", "nano", "Preferred editor [nano, vim, emacs]")
	cmd.Flags().StringVar(&gitName, "git-name", "", "Git user name")
	cmd.Flags().StringVar(&gitEmail, "git-email", "", "Git user email")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force overwrite existing configuration")
	cmd.Flags().BoolVarP(&autoConfig, "yes", "y", false, "Automatic yes to prompts")
	cmd.Flags().BoolVar(&testMode, "test", false, "Run in test mode")
	cmd.Flags().BoolVarP(&autoInstall, "auto-install", "a", false, "Automatically install Nix")

	return cmd
}

func formatError(err error, verbose bool) error {
	if verbose {
		return fmt.Errorf("🚨 Detailed error:\n%+v", err)
	}

	// Extract root error message using standard library
	finalErr := err
	for {
		unwrapped := errors.Unwrap(finalErr)
		if unwrapped == nil {
			break
		}
		finalErr = unwrapped
	}

	// Handle validation errors specially
	var valErr *customerrors.ValidationError
	if errors.As(finalErr, &valErr) {
		return fmt.Errorf("⚠️  %s", valErr.UserMessage)
	}

	// Handle conflict errors specially
	var conflictErr *customerrors.ConflictError
	if errors.As(finalErr, &conflictErr) {
		return fmt.Errorf("🚧 Conflict detected!\n%s\n\nResolution options:\n%s\n\nUse --force to overwrite (creates backup)",
			conflictErr.Message,
			conflictErr.Resolution)
	}

	// Generic error formatting
	return fmt.Errorf("⚠️  %v", finalErr)
}
