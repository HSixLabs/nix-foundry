package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage environment profiles",
	Long:  `Create and manage different environment profiles for different contexts.`,
}

var profileCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create new profile",
	Long: `Create a new environment profile.

Example:
  nix-foundry profile create work
  nix-foundry profile create personal`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return createProfile(args[0])
	},
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List profiles",
	Long:  `List all available environment profiles.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return listProfiles()
	},
}

var profileDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete profile",
	Long:  `Delete an environment profile.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteProfile(args[0])
	},
}

func init() {
	profileCmd.Flags().StringVarP(&profileName, "name", "n", "", "Profile name")
	// Update flags to use forceProfile
	profileCreateCmd.Flags().BoolVar(&forceProfile, "force", false, "Force creation even if profile exists")
	profileDeleteCmd.Flags().BoolVar(&forceProfile, "force", false, "Force deletion without confirmation")

	// Add commands
	profileCmd.AddCommand(profileCreateCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileDeleteCmd)
}

func createProfile(name string) error {
	profilesDir := filepath.Join(getConfigDir(), "profiles")
	profilePath := filepath.Join(profilesDir, name)

	if _, err := os.Stat(profilePath); err == nil && !forceProfile {
		return fmt.Errorf("profile '%s' already exists. Use --force to override", name)
	}

	// Create profile directory
	if err := os.MkdirAll(profilePath, 0755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}

	// Create default config for profile
	config := map[string]interface{}{
		"name":    name,
		"version": "1.0",
		"shell": map[string]interface{}{
			"type": "zsh",
		},
		"editor": map[string]interface{}{
			"type": "neovim",
		},
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to create profile configuration: %w", err)
	}

	configPath := filepath.Join(profilePath, "config.yaml")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write profile configuration: %w", err)
	}

	fmt.Printf("Profile '%s' created successfully\n", name)
	return nil
}

func listProfiles() error {
	profilesDir := filepath.Join(getConfigDir(), "profiles")
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No profiles found")
			return nil
		}
		return fmt.Errorf("failed to read profiles: %w", err)
	}

	if len(entries) == 0 {
		fmt.Println("No profiles found")
		return nil
	}

	fmt.Println("Available profiles:")
	for _, entry := range entries {
		if entry.IsDir() {
			fmt.Printf("- %s\n", entry.Name())
		}
	}

	return nil
}

func deleteProfile(name string) error {
	if name == "default" {
		return fmt.Errorf("cannot delete default profile")
	}

	profilePath := filepath.Join(getConfigDir(), "profiles", name)
	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		return fmt.Errorf("profile '%s' does not exist", name)
	}

	if !forceProfile {
		fmt.Printf("Are you sure you want to delete profile '%s'? [y/N]: ", name)
		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			return fmt.Errorf("failed to read user input: %w", err)
		}
		if !strings.EqualFold(response, "y") && !strings.EqualFold(response, "yes") {
			fmt.Println("Profile deletion cancelled")
			return nil
		}
	}

	if err := os.RemoveAll(profilePath); err != nil {
		return fmt.Errorf("failed to delete profile: %w", err)
	}

	fmt.Printf("Profile '%s' deleted successfully\n", name)
	return nil
}
