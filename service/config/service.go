package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/pkg/schema"
	"github.com/shawnkhoffman/nix-foundry/pkg/validator"

	"github.com/AlecAivazis/survey/v2"
	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"gopkg.in/yaml.v2"
)

type ConfigService struct {
	fs filesystem.FileSystem
}

func NewConfigService(fs filesystem.FileSystem) *ConfigService {
	return &ConfigService{fs: fs}
}

func (s *ConfigService) InitConfig(kind, name string, force, newConfig bool, basePath string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".config", "nix-foundry", kind+"s", name)

	if !force && s.fs.Exists(configPath) {
		return fmt.Errorf("configuration already exists at %s (use --force to overwrite)", configPath)
	}

	if err := s.fs.CreateDir(configPath); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	defaultConfig := s.generateDefaultConfig(kind, name, newConfig, basePath)
	configFile := filepath.Join(configPath, "config.yaml")

	if err := s.fs.WriteFile(configFile, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (s *ConfigService) generateDefaultConfig(kind, name string, minimal bool, base string) string {
	baseSection := ""
	if base != "" {
		baseSection = fmt.Sprintf("base: %s\n", base)
	}

	if minimal {
		return fmt.Sprintf(`version: "1.0"
kind: %s
metadata:
  name: %s
  description: "%s configuration"
%s
nix:
  packages:
    core: []
settings:
  autoUpdate: true
`, kind, name, name, baseSection)
	}

	return fmt.Sprintf(`version: "1.0"
kind: %s
metadata:
  name: %s
  description: "%s configuration"
%s
nix:
  channels:
    - name: "nixpkgs"
      url: "github:NixOS/nixpkgs/nixpkgs-unstable"
  packages:
    core: [htop, jq]
settings:
  autoUpdate: true
  updateInterval: "24h"
  logLevel: info
`, kind, name, name, baseSection)
}

// CopyConfig copies the content from source to destination.
func (s *ConfigService) CopyConfig(source, destination string) error {
	data, err := s.fs.ReadFile(source)
	if err != nil {
		return fmt.Errorf("failed to read source config: %w", err)
	}
	if err := s.fs.WriteFile(destination, data, 0644); err != nil {
		return fmt.Errorf("failed to write destination config: %w", err)
	}
	return nil
}

// ValidateConfig validates the configuration file at the given path.
func (s *ConfigService) ValidateConfig(path string) error {
	content, err := s.fs.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	if err := validator.RegisterDurationValidation(); err != nil {
		return err
	}

	_, err = validator.ValidateYAMLContent(content)
	return err
}

// UpdateActiveConfigWithPackages updates the active configuration with selected packages.
func (s *ConfigService) UpdateActiveConfigWithPackages(activeConfigPath string, packages []string, selectedShell string) error {
	// Read existing config
	data, err := s.fs.ReadFile(activeConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read active config: %w", err)
	}

	var cfg schema.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Append selected packages to core packages
	cfg.Nix.Packages.Core = append(cfg.Nix.Packages.Core, packages...)
	cfg.Nix.Shell = selectedShell

	// Remove duplicates
	cfg.Nix.Packages.Core = Unique(cfg.Nix.Packages.Core)

	// Marshal back to YAML
	updatedData, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal updated config: %w", err)
	}

	// Write back to active config
	if err := s.fs.WriteFile(activeConfigPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write updated config: %w", err)
	}

	// Validate the updated config
	if err := s.ValidateConfig(activeConfigPath); err != nil {
		return fmt.Errorf("validation failed after updating config: %w", err)
	}

	return nil
}

// Unique removes duplicate strings from a slice.
func Unique(slice []string) []string {
	uniqueMap := make(map[string]struct{})
	var result []string
	for _, item := range slice {
		if _, exists := uniqueMap[item]; !exists {
			uniqueMap[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func (s *ConfigService) ConfigExists() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	activeConfigPath := filepath.Join(homeDir, ".config", "nix-foundry", "config.yaml")
	return s.fs.Exists(activeConfigPath)
}

func (s *ConfigService) PromptForSetup() (bool, error) {
	prompt := &survey.Confirm{
		Message: "\033[1mWelcome to Nix Foundry!\033[0m First-time setup is required. This will:\n" +
			"- Create a default configuration file\n" +
			"- Choose your initial packages\n" +
			"- Choose your shell\n" +
			"- Generate the required Nix files\n\n" +
			"Note: This will \033[31mnot\033[0m make changes to your system until you run \033[36mnix-foundry apply\033[0m.\n\n" +
			"Would you like to continue with the initial setup?",
		Default: true,
	}

	var response bool
	err := survey.AskOne(prompt, &response)
	return response, err
}

func (s *ConfigService) ActiveConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "nix-foundry", "config.yaml"), nil
}

// UpdateConfig applies modifications to an existing config file
func (s *ConfigService) UpdateConfig(configPath string, updates ...UpdateFunc) error {
	data, err := s.fs.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	var cfg schema.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	for _, update := range updates {
		update(&cfg)
	}

	updatedData, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal updated config: %w", err)
	}

	if err := s.fs.WriteFile(configPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write updated config: %w", err)
	}

	return s.ValidateConfig(configPath)
}

type UpdateFunc func(*schema.Config)
