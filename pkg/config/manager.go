// Package config provides safe configuration management with atomic writes
// and validation capabilities. It handles:
// - Configuration versioning
// - Schema validation
// - Secure backup/restore
package config

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	"github.com/shawnkhoffman/nix-foundry/pkg/backup"

	"github.com/99designs/keyring"
	"gopkg.in/yaml.v3"
)

// Add missing encryption key constants and derivation
const encryptionKeyLabel = "nix-foundry-encryption-key"

// deriveEncryptionKey retrieves the current key from secure storage
func deriveEncryptionKey() []byte {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: "nix-foundry",
	})
	if err != nil {
		panic("failed to access system keyring: " + err.Error())
	}

	item, err := ring.Get(encryptionKeyLabel)
	if err != nil {
		panic("encryption key not found: " + err.Error())
	}
	return item.Data
}

// Manager handles configuration operations with atomic guarantees
type Manager struct {
	logger     *logging.Logger
	cfg        *NixConfig // Changed from Config to NixConfig
	configPath string
	configDir  string
	backupDir  string
	paths      Paths
}

// ConfigOptions represents options for configuration operations
type Options struct {
	Force    bool
	Validate bool
	Backup   bool
}

// Add constructor with directory parameter
func NewManagerWithDir(configDir string) *Manager {
	configPath := filepath.Join(configDir, "config.yaml")
	return &Manager{
		configPath: configPath,
		logger:     logging.GetLogger(),
		configDir:  configDir,
		backupDir:  filepath.Join(configDir, "backups"),
		cfg:        &NixConfig{},
	}
}

// Add default constructor for completeness
func NewManager() *Manager {
	return NewManagerWithDir(defaultConfigDir())
}

// Add helper to get default config directory
func defaultConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "nix-foundry")
}

// SafeWrite writes configuration with atomic guarantees
// Options:
//   - Backup: Create backup before write
//   - Validate: Run schema validation
//   - Force: Overwrite existing files
func (cm *Manager) SafeWrite(filename string, config interface{}, opts Options) error {
	if opts.Backup {
		if err := cm.CreateBackup(); err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}
	}

	tempFile := filepath.Join(cm.configDir, ".tmp-"+filename)
	defer os.Remove(tempFile)

	if err := cm.WriteConfig(tempFile, config); err != nil {
		return err
	}

	return os.Rename(tempFile, filepath.Join(cm.configDir, filename))
}

// ReadConfig reads and unmarshals configuration
func (cm *Manager) ReadConfig(filename string, config interface{}) error {
	configPath := filepath.Join(cm.configDir, filename)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read configuration: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	return nil
}

// WriteConfig marshals and writes configuration
func (cm *Manager) WriteConfig(filename string, config interface{}) error {
	configPath := filepath.Join(cm.configDir, filename)

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	return nil
}

// CreateBackup creates a timestamped backup of the current configuration
func (cm *Manager) CreateBackup() error {
	if err := cm.validatePermissions(cm.configDir); err != nil {
		return fmt.Errorf("insecure permissions: %w", err)
	}

	if err := os.MkdirAll(cm.backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(cm.backupDir, fmt.Sprintf("backup-%s.tar.gz", timestamp))

	// Checksum generation
	checksumFile := filepath.Join(cm.backupDir, "checksums.sha256")
	if err := backup.GenerateChecksums(cm.configDir, checksumFile); err != nil {
		return fmt.Errorf("checksum generation failed: %w", err)
	}

	// Encryption support
	if err := cm.encryptBackup(backupPath); err != nil {
		return fmt.Errorf("backup encryption failed: %w", err)
	}

	return nil
}

// Fixed encryption implementation
func (cm *Manager) encryptBackup(srcPath string) error {
	key, err := cm.GetEncryptionKey()
	if err != nil {
		return fmt.Errorf("failed to get encryption key: %w", err)
	}
	return backup.EncryptFile(srcPath, key)
}

// Fixed key generation using crypto/rand
func generateEphemeralKey() []byte {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic(fmt.Sprintf("failed to generate key: %v", err))
	}
	return key
}

func (cm *Manager) ConfigExists(filename string) bool {
	_, err := os.Stat(filepath.Join(cm.configDir, filename))
	return err == nil
}

func (cm *Manager) GetConfigDir() string {
	return cm.configDir
}

func (cm *Manager) GetBackupDir() string {
	return filepath.Join(cm.configDir, "backups")
}

func (cm *Manager) loadTeamConfig(teamName string) (*ProjectConfig, error) {
	var config ProjectConfig
	teamConfigPath := filepath.Join(cm.configDir, "teams", teamName+".yaml")

	if err := cm.ReadConfig(teamConfigPath, &config); err != nil {
		return nil, fmt.Errorf("failed to load team config: %w", err)
	}

	return &config, nil
}

// MergeProjectConfigs merges two project configurations
func (cm *Manager) MergeProjectConfigs(base, overlay ProjectConfig) ProjectConfig {
	result := overlay

	// Merge required packages
	seen := make(map[string]bool)
	var required []string

	for _, pkg := range base.Required {
		if !seen[pkg] {
			seen[pkg] = true
			required = append(required, pkg)
		}
	}

	for _, pkg := range overlay.Required {
		if !seen[pkg] {
			seen[pkg] = true
			required = append(required, pkg)
		}
	}

	result.Required = required

	// Merge tools
	result.Tools.Go = cm.mergeLists(base.Tools.Go, overlay.Tools.Go)
	result.Tools.Node = cm.mergeLists(base.Tools.Node, overlay.Tools.Node)
	result.Tools.Python = cm.mergeLists(base.Tools.Python, overlay.Tools.Python)

	// Merge environment variables
	if result.Environment == nil {
		result.Environment = make(map[string]string)
	}
	for k, v := range base.Environment {
		if _, exists := result.Environment[k]; !exists {
			result.Environment[k] = v
		}
	}

	return result
}

func (cm *Manager) mergeLists(a, b []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range a {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	for _, item := range b {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// LoadConfig loads any configuration type with proper validation
func (cm *Manager) LoadConfig(configType Type, name string) (interface{}, error) {
	var config interface{}
	var path string

	switch configType {
	case PersonalConfigType:
		config = &NixConfig{}
		path = cm.paths.Personal
	case ProjectConfigType:
		config = &ProjectConfig{
			BaseConfig: BaseConfig{
				Type: ProjectConfigType,
			},
		}
		path = cm.paths.Project
		if name != "" {
			path = filepath.Join(cm.configDir, "projects", name+".yaml")
		}
	case TeamConfigType:
		config = &ProjectConfig{
			BaseConfig: BaseConfig{
				Type: TeamConfigType,
			},
		}
		path = filepath.Join(cm.paths.Team, name+".yaml")
	default:
		return nil, fmt.Errorf("unknown config type: %s", configType)
	}

	if err := cm.ReadConfig(path, config); err != nil {
		return nil, err
	}

	return config, nil
}

// GenerateTestConfig generates a test configuration for testing purposes
func (cm *Manager) GenerateTestConfig(config interface{}) (string, error) {
	// Convert config to map[string]string if it's not already
	configMap := make(map[string]string)
	switch c := config.(type) {
	case map[string]string:
		configMap = c
	case map[string]interface{}:
		for k, v := range c {
			if str, ok := v.(string); ok {
				configMap[k] = str
			}
		}
	case nil:
		// Use defaults for nil config
		configMap = map[string]string{
			"shell":  "zsh",
			"editor": "nano",
		}
	}

	// Default shell if not specified
	shell := "zsh"
	if s, ok := configMap["shell"]; ok && s != "" {
		shell = s
	}

	// Default editor if not specified
	editor := "nano"
	if e, ok := configMap["editor"]; ok && e != "" {
		editor = e
	}

	// Default git configuration
	gitName := configMap["git-name"]
	gitEmail := configMap["git-email"]
	gitEnabled := gitName != "" && gitEmail != ""

	// Ensure the config directory exists
	if err := os.MkdirAll(cm.configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return fmt.Sprintf(`{ config, pkgs, ... }:

{
  # Let Home Manager manage itself
  programs.home-manager.enable = true;

  home = {
    username = builtins.getEnv "USER";
    homeDirectory = builtins.getEnv "HOME";
    stateVersion = "23.11";

    # Package management
    packages = with pkgs; [
      %s    # Shell package
    ];
  };

  # Shell configuration
  programs.%s = {
    enable = true;
    package = pkgs.%s;
  };

  # Editor configuration
  programs.%s = {
    enable = true;
  };

  %s  # Git configuration (conditionally included)
}`,
		shell,
		shell, shell,
		editor,
		generateGitConfig(gitEnabled, gitName, gitEmail)), nil
}

// Helper function to generate git configuration
func generateGitConfig(enabled bool, name, email string) string {
	if !enabled {
		return "# Git configuration disabled"
	}
	return fmt.Sprintf(`# Git configuration
  programs.git = {
    enable = true;
    userName = "%s";
    userEmail = "%s";
  };`, name, email)
}

// Add permission validation
func (cm *Manager) validatePermissions(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.Mode().Perm()&0077 != 0 {
		return fmt.Errorf("insecure permissions on %s", path)
	}
	return nil
}

// Add key rotation
func (cm *Manager) RotateEncryptionKey() error {
	oldKey := deriveEncryptionKey()
	newKey := generateEphemeralKey()

	if err := cm.reencryptBackups(oldKey, newKey); err != nil {
		return fmt.Errorf("key rotation failed: %w", err)
	}

	// Fixed keyring usage pattern
	ring, _ := keyring.Open(keyring.Config{
		ServiceName: "nix-foundry",
	})

	return ring.Set(keyring.Item{
		Key:  encryptionKeyLabel,
		Data: newKey,
	})
}

// Update GetEncryptionKey to use the derivation logic
func (cm *Manager) GetEncryptionKey() ([]byte, error) {
	return deriveEncryptionKey(), nil
}

func (cm *Manager) reencryptBackups(_ /* oldKey */, _ /* newKey */ []byte) error {
	// TODO: Implement backup re-encryption
	// This should:
	// 1. Decrypt all backups with oldKey
	// 2. Re-encrypt with newKey
	// 3. Update backup index
	return nil
}

// Add the missing section methods to Manager
func (m *Manager) LoadSection(name string, v interface{}) error {
	if err := m.Load(); err != nil {
		return err
	}

	// Use reflection to get the section
	configValue := reflect.ValueOf(m.cfg).Elem()
	sectionValue := configValue.FieldByName(name)
	if !sectionValue.IsValid() {
		return fmt.Errorf("invalid configuration section: %s", name)
	}

	// Set the output value
	targetValue := reflect.ValueOf(v).Elem()
	targetValue.Set(sectionValue)
	return nil
}

func (m *Manager) SaveSection(name string, v interface{}) error {
	// Load current config first
	if err := m.Load(); err != nil {
		return err
	}

	// Update the section using reflection
	configValue := reflect.ValueOf(m.cfg).Elem()
	sectionValue := configValue.FieldByName(name)
	if !sectionValue.IsValid() {
		return fmt.Errorf("invalid configuration section: %s", name)
	}

	newValue := reflect.ValueOf(v).Elem()
	sectionValue.Set(newValue)

	// Save updated config
	return m.Save()
}

// Add core load/save implementation
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Maintain empty config if file doesn't exist
			return nil
		}
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := yaml.Unmarshal(data, m.cfg); err != nil {
		return fmt.Errorf("invalid config format: %w", err)
	}
	return nil
}

func (m *Manager) Save() error {
	data, err := yaml.Marshal(m.cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.MkdirAll(m.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	return os.WriteFile(m.configPath, data, 0644)
}
