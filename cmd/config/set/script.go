/*
Package set provides commands for setting various configuration options in Nix Foundry.
This package specifically handles the configuration of shell scripts, packages,
and other settable properties within the Nix Foundry configuration system.
*/
package set

import (
	"fmt"
	"io"
	"os"

	"github.com/shawnkhoffman/nix-foundry/pkg/config"
	"github.com/shawnkhoffman/nix-foundry/pkg/schema"
	"github.com/spf13/cobra"
)

var (
	scriptName        string
	scriptDesc        string
	scriptFile        string
	scriptInline      string
	scriptInteractive bool
)

var scriptCmd = &cobra.Command{
	Use:   "script [name]",
	Short: "Set a shell script in configuration",
	Long: `Set a shell script in configuration.
This command allows you to add or update a shell script in your configuration.
The script can be provided from a file, inline, or interactively.`,
	RunE: runScript,
}

func init() {
	Cmd.AddCommand(scriptCmd)
	scriptCmd.Flags().StringVarP(&scriptName, "name", "n", "", "Script name (required)")
	scriptCmd.Flags().StringVarP(&scriptDesc, "desc", "d", "", "Script description")
	scriptCmd.Flags().StringVarP(&scriptFile, "file", "f", "", "Path to script file")
	scriptCmd.Flags().StringVarP(&scriptInline, "script", "s", "", "Inline script content")
	scriptCmd.Flags().BoolVarP(&scriptInteractive, "interactive", "i", false, "Enter script content interactively")
	if err := scriptCmd.MarkFlagRequired("name"); err != nil {
		panic(fmt.Sprintf("failed to mark name flag as required: %v", err))
	}
}

/*
runScript handles the creation or update of shell scripts in the Nix Foundry configuration.
It supports three methods of script input:
1. File-based: Reading script content from a specified file
2. Inline: Using directly provided script content
3. Interactive: Reading script content from standard input

The function will:
1. Load the active configuration
2. Acquire script content from the specified source
3. Create or update the script in the configuration
4. Save the updated configuration

Returns an error if any step fails.
*/
func runScript(cmd *cobra.Command, args []string) error {
	configSvc := config.GetConfigService()

	config, err := configSvc.GetActiveConfig()
	if err != nil {
		return fmt.Errorf("failed to get active config: %w", err)
	}

	var content string
	switch {
	case scriptFile != "":
		data, err := os.ReadFile(scriptFile)
		if err != nil {
			return fmt.Errorf("failed to read script file: %w", err)
		}
		content = string(data)

	case scriptInline != "":
		content = scriptInline

	case scriptInteractive:
		fmt.Println("Enter script content (press Ctrl+D when done):")
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read script content: %w", err)
		}
		content = string(data)

	default:
		return fmt.Errorf("must specify one of: --file, --script, or --interactive")
	}

	script := schema.Script{
		Name:        scriptName,
		Description: scriptDesc,
		Commands:    schema.MultiLineString(content),
	}

	found := false
	for i, s := range config.Nix.Scripts {
		if s.Name == scriptName {
			config.Nix.Scripts[i] = script
			found = true
			break
		}
	}

	if !found {
		config.Nix.Scripts = append(config.Nix.Scripts, script)
	}

	if err := configSvc.SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("✨ Script '%s' saved successfully\n", scriptName)
	return nil
}
