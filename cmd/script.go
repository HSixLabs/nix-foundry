// Package cmd provides the command-line interface for Nix Foundry.
package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/shawnkhoffman/nix-foundry/pkg/schema"
	scriptmgr "github.com/shawnkhoffman/nix-foundry/pkg/script"
	"github.com/shawnkhoffman/nix-foundry/service/config"
	"github.com/spf13/cobra"
)

var scriptCmd = &cobra.Command{
	Use:   "script",
	Short: "Manage shell scripts",
	Long: `Manage shell scripts with platform-specific optimizations.
Supports adding, removing, listing, and running scripts.`,
}

var addScriptCmd = &cobra.Command{
	Use:   "add [file]",
	Short: "Add a script to configuration",
	Long: `Add a shell script to Nix Foundry configuration.
The script will be stored in the configuration and can be run using the 'run' command.

Examples:
  # Add a script with name and description
  nix-foundry script add setup.sh --name "Setup Dev" --desc "Setup development environment"

  # Add a script from stdin
  cat setup.sh | nix-foundry script add - --name "Setup Dev"`,
	RunE: runAddScript,
}

var removeScriptCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a script from configuration",
	Long: `Remove a shell script from Nix Foundry configuration.

Example:
  nix-foundry script remove "Setup Dev"`,
	RunE: runRemoveScript,
}

var listScriptsCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured scripts",
	Long: `List all scripts in Nix Foundry configuration.

Example:
  nix-foundry script list`,
	RunE: runListScripts,
}

var runScriptCmd = &cobra.Command{
	Use:   "run [name]",
	Short: "Execute a script",
	Long: `Execute a script from Nix Foundry configuration.

Example:
  nix-foundry script run "Setup Dev"`,
	RunE: runScript,
}

var (
	scriptName string
	scriptDesc string
)

func init() {
	rootCmd.AddCommand(scriptCmd)
	scriptCmd.AddCommand(addScriptCmd)
	scriptCmd.AddCommand(removeScriptCmd)
	scriptCmd.AddCommand(listScriptsCmd)
	scriptCmd.AddCommand(runScriptCmd)

	addScriptCmd.Flags().StringVar(&scriptName, "name", "", "Name of the script")
	addScriptCmd.Flags().StringVar(&scriptDesc, "desc", "", "Description of the script")
	if err := addScriptCmd.MarkFlagRequired("name"); err != nil {
		panic(fmt.Sprintf("failed to mark flag as required: %v", err))
	}
}

func runAddScript(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("script file path is required")
	}

	var content []byte
	var err error

	if args[0] == "-" {
		content, err = io.ReadAll(os.Stdin)
	} else {
		content, err = os.ReadFile(args[0])
	}
	if err != nil {
		return fmt.Errorf("failed to read script: %w", err)
	}

	script := schema.Script{
		Name:        scriptName,
		Description: scriptDesc,
		Commands:    schema.MultiLineString(content),
	}

	fs := filesystem.NewOSFileSystem()
	configSvc := config.NewService(fs)
	cfg, err := configSvc.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	scriptMgr := scriptmgr.NewManager(fs)
	if err := scriptMgr.AddScript(script, cfg); err != nil {
		return fmt.Errorf("failed to add script: %w", err)
	}

	if err := configSvc.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✨ Added script '%s' to configuration\n", script.Name)
	return nil
}

func runRemoveScript(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("script name is required")
	}

	name := args[0]
	fs := filesystem.NewOSFileSystem()
	configSvc := config.NewService(fs)
	cfg, err := configSvc.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	scriptMgr := scriptmgr.NewManager(fs)
	if err := scriptMgr.RemoveScript(name, cfg); err != nil {
		return fmt.Errorf("failed to remove script: %w", err)
	}

	if err := configSvc.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✨ Removed script '%s' from configuration\n", name)
	return nil
}

func runListScripts(cmd *cobra.Command, args []string) error {
	fs := filesystem.NewOSFileSystem()
	configSvc := config.NewService(fs)
	cfg, err := configSvc.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	scriptMgr := scriptmgr.NewManager(fs)
	scripts := scriptMgr.ListScripts(cfg)

	if len(scripts) == 0 {
		fmt.Println("No scripts configured")
		return nil
	}

	fmt.Println("Configured scripts:")
	for _, script := range scripts {
		if script.Description != "" {
			fmt.Printf("  %s: %s\n", script.Name, script.Description)
		} else {
			fmt.Printf("  %s\n", script.Name)
		}
	}

	return nil
}

func runScript(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("script name is required")
	}

	name := args[0]
	fs := filesystem.NewOSFileSystem()
	configSvc := config.NewService(fs)
	cfg, err := configSvc.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	scriptMgr := scriptmgr.NewManager(fs)
	if err := scriptMgr.RunScript(name, cfg); err != nil {
		return fmt.Errorf("failed to run script: %w", err)
	}

	fmt.Printf("✨ Script '%s' executed successfully\n", name)
	return nil
}
