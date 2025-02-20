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

func NewConfigService() *ConfigService {
	return &ConfigService{
		fs: filesystem.NewOSFileSystem(),
	}
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
func (s *ConfigService) UpdateActiveConfigWithPackages(activeConfigPath string, packages []string) error {
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

	// Remove duplicates
	cfg.Nix.Packages.Core = unique(cfg.Nix.Packages.Core)

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

// unique removes duplicate strings from a slice.
func unique(slice []string) []string {
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
		Message: "Welcome to Nix Foundry! Would you like to run the interactive setup?",
		Default: true,
	}

	var response bool
	err := survey.AskOne(prompt, &response)
	if err != nil {
		return false, fmt.Errorf("failed to get user input: %w", err)
	}

	return response, nil
}
