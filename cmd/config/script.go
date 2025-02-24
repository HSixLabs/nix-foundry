// Package config provides configuration management commands for Nix Foundry.
package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/pkg/schema"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	scriptName     string
	scriptDesc     string
	scriptFile     string
	scriptCommands string
)

// NewScriptCmd creates a new script command.
func NewScriptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "script",
		Short: "Manage shell scripts in the configuration",
		Long: `Manage shell scripts in the configuration.
This command provides subcommands for managing shell scripts.`,
	}

	cmd.AddCommand(
		newScriptSetCmd(),
		newScriptListCmd(),
		newScriptRunCmd(),
	)

	return cmd
}

func newScriptSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set a script in the configuration",
		Long: `Set a script in the configuration.
The script can be provided either from a file or inline.`,
		RunE: runScriptSet,
	}

	cmd.Flags().StringVarP(&scriptName, "name", "n", "", "Name of the script (required)")
	cmd.Flags().StringVarP(&scriptDesc, "desc", "d", "", "Description of the script")
	cmd.Flags().StringVarP(&scriptFile, "file", "f", "", "Path to script file")
	cmd.Flags().StringVarP(&scriptCommands, "commands", "c", "", "Script commands (inline)")

	if err := cmd.MarkFlagRequired("name"); err != nil {
		panic(fmt.Sprintf("failed to mark flag as required: %v", err))
	}

	return cmd
}

func newScriptListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List scripts",
		Long:  `List all scripts in the configuration.`,
		RunE:  runScriptList,
	}
}

func newScriptRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [script-name]",
		Short: "Run a script",
		Long:  `Run a script from the configuration.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runScriptRun,
	}

	return cmd
}

func runScriptSet(cmd *cobra.Command, args []string) error {
	configPath, err := schema.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config schema.Config
	if unmarshalErr := yaml.Unmarshal(content, &config); unmarshalErr != nil {
		return fmt.Errorf("failed to parse config file: %w", unmarshalErr)
	}

	var scriptContent string
	if scriptFile != "" {
		fileContent, readErr := os.ReadFile(scriptFile)
		if readErr != nil {
			return fmt.Errorf("failed to read script file: %w", readErr)
		}
		scriptContent = string(fileContent)
	} else if scriptCommands != "" {
		scriptContent = scriptCommands
	} else {
		return fmt.Errorf("either --file or --commands must be provided")
	}

	script := schema.Script{
		Name:        scriptName,
		Description: scriptDesc,
		Commands:    schema.MultiLineString(scriptContent),
	}

	var found bool
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

	node := &yaml.Node{}
	if encodeErr := node.Encode(config); encodeErr != nil {
		return fmt.Errorf("failed to encode config: %w", encodeErr)
	}

	var setScriptStyle func(*yaml.Node)
	setScriptStyle = func(n *yaml.Node) {
		if n.Kind == yaml.MappingNode {
			for i := 0; i < len(n.Content); i += 2 {
				key := n.Content[i]
				value := n.Content[i+1]
				if key.Value == "commands" {
					value.Style = yaml.LiteralStyle
				}
				setScriptStyle(value)
			}
		} else if n.Kind == yaml.SequenceNode {
			for _, item := range n.Content {
				setScriptStyle(item)
			}
		}
	}
	setScriptStyle(node)

	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if encodeErr := encoder.Encode(node); encodeErr != nil {
		return fmt.Errorf("failed to marshal config: %w", encodeErr)
	}

	if writeErr := os.WriteFile(configPath, []byte(buf.String()), 0644); writeErr != nil {
		return fmt.Errorf("failed to write config: %w", writeErr)
	}

	fmt.Printf("✨ Script '%s' saved successfully\n", scriptName)
	return nil
}

func runScriptList(cmd *cobra.Command, args []string) error {
	configPath, err := schema.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config schema.Config
	if unmarshalErr := yaml.Unmarshal(content, &config); unmarshalErr != nil {
		return fmt.Errorf("failed to parse config file: %w", unmarshalErr)
	}

	if len(config.Nix.Scripts) == 0 {
		fmt.Println("No scripts found")
		return nil
	}

	fmt.Println("Available scripts:")
	for _, script := range config.Nix.Scripts {
		if script.Description != "" {
			fmt.Printf("  %s - %s\n", script.Name, script.Description)
		} else {
			fmt.Printf("  %s\n", script.Name)
		}
	}

	return nil
}

func runScriptRun(cmd *cobra.Command, args []string) error {
	scriptName := args[0]

	configPath, err := schema.GetConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get config path: %w", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config schema.Config
	if unmarshalErr := yaml.Unmarshal(content, &config); unmarshalErr != nil {
		return fmt.Errorf("failed to parse config file: %w", unmarshalErr)
	}

	var script *schema.Script
	for _, s := range config.Nix.Scripts {
		if s.Name == scriptName {
			script = &s
			break
		}
	}

	if script == nil {
		return fmt.Errorf("script '%s' not found", scriptName)
	}

	tmpDir, err := os.MkdirTemp("", "nix-foundry-script-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	scriptPath := filepath.Join(tmpDir, "script.sh")
	if writeErr := os.WriteFile(scriptPath, []byte(script.Commands), 0755); writeErr != nil {
		return fmt.Errorf("failed to write script file: %w", writeErr)
	}

	execCmd := exec.Command(config.Settings.Shell, scriptPath)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	if runErr := execCmd.Run(); runErr != nil {
		return fmt.Errorf("failed to run script: %w", runErr)
	}

	return nil
}
