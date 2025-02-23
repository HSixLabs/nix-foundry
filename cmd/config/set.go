// Package config provides configuration management commands for Nix Foundry.
package config

import (
	"fmt"
	"os"

	"github.com/shawnkhoffman/nix-foundry/pkg/schema"
	"github.com/spf13/cobra"
)

// NewSetCmd creates a new set command.
func NewSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set configuration values",
		Long: `Set configuration values.
This command provides subcommands for setting various configuration options.`,
	}

	cmd.AddCommand(
		newSetManagerCmd(),
		newSetScriptCmd(),
	)

	return cmd
}

func newSetManagerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manager [nix-env]",
		Short: "Set package manager",
		Long: `Set the package manager to use.
Currently only supports:
  - nix-env: Use nix-env for package management`,
		Args: cobra.ExactArgs(1),
		RunE: runSetManager,
	}

	return cmd
}

func newSetScriptCmd() *cobra.Command {
	var (
		name string
		desc string
		file string
	)

	cmd := &cobra.Command{
		Use:   "script",
		Short: "Add or update a script",
		Long: `Add a shell script to Nix Foundry configuration.
The script will be stored in the configuration and can be run using the 'run' command.

Examples:
  # Add a script with name and description
  nix-foundry config set script --name "Setup Dev" --desc "Setup development environment" --file setup.sh

  # Add a script from stdin
  cat setup.sh | nix-foundry config set script --name "Setup Dev" --file -`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetScript(name, desc, file)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Name of the script")
	cmd.Flags().StringVar(&desc, "desc", "", "Description of the script")
	cmd.Flags().StringVar(&file, "file", "", "Path to script file")

	for _, flag := range []string{"name", "file"} {
		if err := cmd.MarkFlagRequired(flag); err != nil {
			panic(fmt.Sprintf("failed to mark flag '%s' as required: %v", flag, err))
		}
	}

	return cmd
}

func runSetManager(cmd *cobra.Command, args []string) error {
	manager := args[0]
	if manager != "nix-env" {
		return fmt.Errorf("invalid package manager: %s (only 'nix-env' is supported)", manager)
	}

	configSvc := getConfigService()
	if err := configSvc.SetPackageManager(manager); err != nil {
		return fmt.Errorf("failed to set package manager: %w", err)
	}

	fmt.Printf("✨ Package manager set to '%s'\n", manager)
	return nil
}

func runSetScript(name, desc, file string) error {
	var content []byte
	var err error

	if file == "-" {
		content, err = os.ReadFile(os.Stdin.Name())
	} else {
		content, err = os.ReadFile(file)
	}
	if err != nil {
		return fmt.Errorf("failed to read script: %w", err)
	}

	script := schema.Script{
		Name:        name,
		Description: desc,
		Commands:    schema.MultiLineString(content),
	}

	configSvc := getConfigService()
	if err := configSvc.SetScript(script); err != nil {
		return fmt.Errorf("failed to set script: %w", err)
	}

	fmt.Printf("✨ Script '%s' added to configuration\n", name)
	return nil
}
