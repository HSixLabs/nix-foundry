// Package config provides configuration management functionality.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/shawnkhoffman/nix-foundry/pkg/schema"
	"gopkg.in/yaml.v3"
)

// Service provides configuration management functionality.
type Service struct {
	fs filesystem.FileSystem
}

// NewService creates a new configuration service.
func NewService(fs filesystem.FileSystem) *Service {
	return &Service{fs: fs}
}

// LoadConfig loads the configuration from disk.
func (s *Service) LoadConfig() (*schema.Config, error) {
	configPath := getConfigPath()
	content, err := s.fs.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return schema.NewDefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config schema.Config
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration to disk.
func (s *Service) SaveConfig(config *schema.Config) error {
	configPath := getConfigPath()
	configDir := filepath.Dir(configPath)

	if err := s.fs.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	node := &yaml.Node{}
	if err := node.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
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
	if err := encoder.Encode(node); err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := s.fs.WriteFile(configPath, []byte(buf.String()), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// InitConfig initializes a new configuration file.
func (s *Service) InitConfig() error {
	config := schema.NewDefaultConfig()
	return s.SaveConfig(config)
}

// UninstallConfig removes the configuration file.
func (s *Service) UninstallConfig() error {
	configPath := getConfigPath()
	if err := s.fs.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config: %w", err)
	}
	return nil
}

// SetPackageManager sets the package manager in the configuration.
func (s *Service) SetPackageManager(manager string) error {
	config, err := s.LoadConfig()
	if err != nil {
		return err
	}

	config.Nix.Manager = manager
	return s.SaveConfig(config)
}

// SetScript adds or updates a script in the configuration.
func (s *Service) SetScript(script schema.Script) error {
	config, err := s.LoadConfig()
	if err != nil {
		return err
	}

	var updated bool
	for i, s := range config.Nix.Scripts {
		if s.Name == script.Name {
			config.Nix.Scripts[i] = script
			updated = true
			break
		}
	}

	if !updated {
		config.Nix.Scripts = append(config.Nix.Scripts, script)
	}

	return s.SaveConfig(config)
}

// getConfigPath returns the path to the configuration file.
func getConfigPath() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(configDir, "nix-foundry", "config.yaml")
}
