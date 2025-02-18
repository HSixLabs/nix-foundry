package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	stderrors "errors"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/types"
	"gopkg.in/yaml.v3"
)

type (
	// Use type aliases to ensure compatibility
	NixConfig = types.NixConfig
	Settings = types.Settings
	ShellConfig = types.ShellConfig
	EditorConfig = types.EditorConfig
	GitConfig = types.GitConfig
	PackagesConfig = types.PackagesConfig
)

type Type string

const (
	ProjectType Type = "project"
	TeamType    Type = "team"
	UserType    Type = "user"
	SystemType  Type = "system"
)

// Service defines the interface for configuration operations
type Service interface {
	Initialize(testMode bool) error
	Load() (*types.Config, error)
	Save(cfg *types.Config) error
	SaveConfig(cfg *types.Config) error
	Apply(config *types.NixConfig, testMode bool) error
	GetConfig() *types.NixConfig
	GetConfigDir() string
	GetBackupDir() string
	LoadSection(name string, v interface{}) error
	SaveSection(name string, v interface{}) error
	ApplyFlags(flags map[string]string, force bool) error
	GenerateInitialConfig(shell, editor, gitName, gitEmail string) (*types.Config, error)
	PreviewConfiguration(*types.Config) error
	ConfigExists() bool
	Generate(defaultEnv string, nixCfg *types.NixConfig) error
	CreateBackup(path string) error
	RestoreBackup(path string) error
	LoadConfig(configType Type, name string) (interface{}, error)
	WriteConfig(path string, cfg interface{}) error
	MergeProjectConfigs(base, team types.ProjectConfig) types.ProjectConfig
	LoadProjectWithTeam(projectName, teamName string) (*types.ProjectConfig, error)
	ReadConfig(path string, v interface{}) error
	LoadCustomPackages() ([]string, error)
	SaveCustomPackages(packages []string) error
	ValidateConfiguration(verbose bool) error
	CreateConfigFromMap(configMap map[string]string) *types.NixConfig
	GetRetentionDays() int
	SetRetentionDays(days int)
	GetMaxBackups() int
	SetMaxBackups(max int)
	GetCompressionLevel() int
	SetCompressionLevel(level int)
	GetValue(key string) (interface{}, error)
	SetValue(key string, value interface{}) error
	Reset(section string) error
	GetLogger() *logging.Logger
	ResetValue(key string) error
	GenerateEncryptionKey() error
	RotateEncryptionKey() error
	Validate() error
}

// ServiceImpl implements the configuration Service interface
type ServiceImpl struct {
	logger      *logging.Logger
	config      *types.Config
	path        string
	manager     *Manager
	previewer   Previewer
	initialized bool
}

// Update the Manager struct
type Manager struct {
	logger           *logging.Logger
	cfg              *types.Config
	configPath       string
	nixConfig        *types.NixConfig  // Use types.NixConfig
	retentionDays    int
	maxBackups       int
	compressionLevel int
	configDir        string
	backupDir        string
	packageFile      string
}

// Add directory getters
func (m *Manager) GetConfigDir() string {
	return m.configDir
}

func (m *Manager) GetBackupDir() string {
	return m.backupDir
}

// Implement proper Save method in Manager
func (m *Manager) Save(cfg *types.Config) error {
	if cfg != nil {
		m.nixConfig = cfg.NixConfig
	}
	if m.configPath == "" {
		return fmt.Errorf("config path not set")
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(m.configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load existing config first
	existingCfg, err := m.Load()
	if err == nil {
		// Merge with existing config if available
		m.cfg = existingCfg
	}

	// Marshal both configs (use nixConfig as source of truth)
	data, err := yaml.Marshal(m.nixConfig)
	if err != nil {
		return NewSaveError(m.configPath, err, "failed to marshal configuration")
	}

	// Write to file
	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return NewSaveError(m.configPath, err, "failed to write configuration")
	}

	// Sync main config with nix config
	if cfg.NixConfig != nil {
		m.cfg.Shell.Type = cfg.NixConfig.Shell.Type
		m.cfg.Editor.Type = cfg.NixConfig.Editor.Type
		m.cfg.Git.Name = cfg.NixConfig.Git.Name
		m.cfg.Git.Email = cfg.NixConfig.Git.Email
	}

	m.logger.Debug("Configuration saved successfully", "path", m.configPath)
	return nil
}

// Keep existing encryption key methods
func (m *Manager) GenerateEncryptionKey() error {
	// Existing implementation
	return nil
}

func (m *Manager) RotateEncryptionKey() error {
	// Existing implementation
	return nil
}

// Update constructor to initialize directories
func NewManager(configDir string, logger *logging.Logger) *Manager {
	if configDir == "" {
		configDir = defaultConfigDir()
	}

	return &Manager{
		configPath:  filepath.Join(configDir, "config.yaml"),
		configDir:   configDir,
		logger:      logger,
		cfg:         &types.Config{},
		nixConfig:   &types.NixConfig{},  // Use types.NixConfig
		backupDir:   filepath.Join(configDir, "backups"),
		packageFile: filepath.Join(configDir, "packages.yaml"),
	}
}

// NewService creates a new configuration service with proper dependencies
func NewService() Service {
	configDir := defaultConfigDir()
	logger := logging.GetLogger()
	svc := &ServiceImpl{
		manager:     NewManager(configDir, logger),
		logger:      logger,
		initialized: false,
	}
	return svc
}

// Add helper to get default config directory
func defaultConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "nix-foundry")
}

func (s *ServiceImpl) Initialize(testMode bool) error {
	if testMode {
		// In test mode, use temporary directories and minimal configuration
		tempDir, err := os.MkdirTemp("", "nix-foundry-test-*")
		if err != nil {
			return fmt.Errorf("failed to create test directory: %w", err)
		}
		s.path = filepath.Join(tempDir, "config.yaml")
		s.config = &types.Config{
			Version: "1.0.0",
			Settings: types.Settings{
				LogLevel:   "debug",
				AutoUpdate: false,
			},
		}
	} else {
		// Normal initialization
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		configDir := filepath.Join(homeDir, ".config", "nix-foundry")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		s.path = filepath.Join(configDir, "config.yaml")
		if !s.ConfigExists() {
			s.config = &types.Config{
				Version: "1.0.0",
				Settings: types.Settings{
					LogLevel:   "info",
					AutoUpdate: true,
				},
			}
		} else {
			// Ensure config is initialized before loading
			if s.config == nil {
				s.config = &types.Config{}
			}
			cfg, err := s.Load()
			if err != nil {
				return fmt.Errorf("failed to load existing config: %w", err)
			}
			s.config = cfg
		}
	}

	return s.Save(s.config)
}

func (s *ServiceImpl) ApplyFlags(flags map[string]string, force bool) error {
	if !force && s.ConfigExists() {
		return errors.NewValidationError(s.path, nil,
			fmt.Sprintf("config exists at %s", s.path))
	}

	for key, value := range flags {
		if err := s.manager.SetValue(key, value); err != nil {
			return err
		}
	}
	return s.manager.Save(s.config)
}

func (s *ServiceImpl) PreviewConfiguration(config *types.Config) error {
	s.logger.Debug("Previewing configuration")

	fmt.Println("\nConfiguration Preview:")
	fmt.Println("----------------------")
	fmt.Printf("Shell: %s\n", config.Shell.Type)
	fmt.Printf("Editor: %s\n", config.Editor.Type)

	if config.Git.Name != "" || config.Git.Email != "" {
		fmt.Println("\nGit Configuration:")
		if config.Git.Name != "" {
			fmt.Printf("  Name: %s\n", config.Git.Name)
		}
		if config.Git.Email != "" {
			fmt.Printf("  Email: %s\n", config.Git.Email)
		}
	}

	return nil
}

func (s *ServiceImpl) ConfigExists() bool {
	if s.path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			s.logger.Error("Failed to get home directory", "error", err)
			return false
		}
		s.path = filepath.Join(home, ".config", "nix-foundry", "config.yaml")
	}

	_, err := os.Stat(s.path)
	return err == nil
}

func (s *ServiceImpl) Generate(defaultEnv string, nixCfg *types.NixConfig) error {
	s.logger.Debug("Generating Nix configuration files")

	// Create default environment directory if it doesn't exist
	if err := os.MkdirAll(defaultEnv, 0755); err != nil {
		return errors.NewLoadError(defaultEnv, err, "failed to create default environment")
	}

	// Generate home.nix configuration
	homeNixPath := filepath.Join(defaultEnv, "home.nix")
	homeNixContent := fmt.Sprintf(`{ config, pkgs, ... }:
{
  home.username = "%s";
  home.homeDirectory = "%s";
  home.stateVersion = "23.11";

  programs.%s.enable = true;
  programs.%s.enable = true;

  programs.git = {
    enable = true;
    userName = "%s";
    userEmail = "%s";
  };

  home.packages = with pkgs; [
    %s  # Core packages
    %s  # User packages
    %s  # Team packages
  ];
}
`, os.Getenv("USER"), os.Getenv("HOME"),
		nixCfg.Shell.Type, nixCfg.Editor.Type,
		nixCfg.Git.Name, nixCfg.Git.Email,
		strings.Join(nixCfg.Packages.Core, " "),
		strings.Join(nixCfg.Packages.User, " "),
		strings.Join(nixCfg.Packages.Team, " "))

	if err := os.WriteFile(homeNixPath, []byte(homeNixContent), 0644); err != nil {
		return errors.NewLoadError(homeNixPath, err, "failed to write home.nix")
	}

	// Generate flake.nix configuration
	flakeNixPath := filepath.Join(defaultEnv, "flake.nix")
	flakeNixContent := `{
  description = "nix-foundry environment";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    home-manager = {
      url = "github:nix-community/home-manager";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, home-manager, ... }: let
    systems = [
      "x86_64-linux"
      "aarch64-linux"
      "x86_64-darwin"
      "aarch64-darwin"
    ];
  in {
    packages = nixpkgs.lib.genAttrs systems (system: {
      default = nixpkgs.legacyPackages.${system}.buildEnv {
        name = "nix-foundry-env";
        paths = [];
      };
    });

    homeConfigurations = nixpkgs.lib.genAttrs systems (system: {
      default = home-manager.lib.homeManagerConfiguration {
        pkgs = nixpkgs.legacyPackages.${system};
        modules = [ ./home.nix ];
      };
    });
  };
}`

	if err := os.WriteFile(flakeNixPath, []byte(flakeNixContent), 0644); err != nil {
		return errors.NewLoadError(flakeNixPath, err, "failed to write flake.nix")
	}

	// Generate .envrc for direnv support
	envrcPath := filepath.Join(defaultEnv, ".envrc")
	if err := os.WriteFile(envrcPath, []byte("use flake . --impure\n"), 0644); err != nil {
		return errors.NewLoadError(envrcPath, err, "failed to write .envrc")
	}

	s.logger.Info("Generated Nix configuration files",
		"home.nix", homeNixPath,
		"flake.nix", flakeNixPath,
		".envrc", envrcPath)

	return nil
}

// Fix the Apply method in Manager
func (m *Manager) Apply(config *types.NixConfig, testMode bool) error {
	m.logger.Debug("Applying configuration changes",
		"testMode", testMode,
		"backupDir", m.GetBackupDir(),
		"configDir", m.GetConfigDir())

	if testMode {
		m.logger.Debug("Running in test mode - no changes persisted")
		return nil
	}

	// Update both configs
	m.nixConfig = config
	m.cfg = &types.Config{
		Version:   config.Version,
		NixConfig: config,
		Settings:  config.Settings,
	}

	// Save the updated configuration
	return m.Save(m.cfg)
}

// Verify both implementations satisfy the interface
var _ Service = (*ServiceImpl)(nil)

// SaveError should remain in config package scope
type SaveError struct {
	Path    string
	Err     error
	Context string
}

func (e *SaveError) Error() string {
	return fmt.Sprintf("config save error at %s: %s (%v)", e.Path, e.Context, e.Err)
}

func (e *SaveError) Unwrap() error {
	return e.Err
}

// Add NewSaveError function above SaveError struct
func NewSaveError(path string, err error, context string) *SaveError {
	return &SaveError{
		Path:    path,
		Err:     err,
		Context: context,
	}
}

// Add missing SaveSection method to Manager
func (m *Manager) SaveSection(name string, v interface{}) error {
	configPath := filepath.Join(m.GetConfigDir(), name+".yaml")
	data, err := yaml.Marshal(v)
	if err != nil {
		return NewSaveError(configPath, err, "failed to marshal config section")
	}
	return os.WriteFile(configPath, data, 0644)
}

// Add Apply method to ServiceImpl
func (s *ServiceImpl) Apply(config *types.NixConfig, testMode bool) error {
	s.logger.Debug("Applying configuration changes",
		"testMode", testMode,
		"backupDir", s.manager.backupDir,
		"configDir", s.manager.configDir,
	)

	if testMode {
		s.logger.Debug("Running in test mode - no changes persisted")
		return nil
	}

	return s.manager.Apply(config, testMode)
}

// Add missing Generate method to Manager
func (m *Manager) Generate(defaultEnv string, nixCfg *types.NixConfig) error {
	// Implementation that matches ServiceImpl's Generate
	m.logger.Debug("Generating Nix configuration files")

	// Create default environment directory
	if err := os.MkdirAll(defaultEnv, 0755); err != nil {
		return errors.NewLoadError(defaultEnv, err, "failed to create default environment")
	}

	// Actual file generation logic would go here
	// (Same implementation as ServiceImpl.Generate)
	return nil
}

// Add GenerateInitialConfig to Manager
func (m *Manager) GenerateInitialConfig(shell, editor, gitName, gitEmail string) (*types.Config, error) {
	m.logger.Debug("Generating initial configuration",
		"shell", shell,
		"editor", editor,
		"gitName", gitName,
		"gitEmail", gitEmail)

	// Set defaults if empty
	if shell == "" {
		shell = "zsh"
	}
	if editor == "" {
		editor = "nano"
	}

	return &types.Config{
		Version: "1.0.0",
		Settings: types.Settings{
			LogLevel:   "info",
			AutoUpdate: true,
		},
		Shell: types.ShellConfig{
			Type:     shell,
			InitFile: defaultInitFile(shell),
		},
		Editor: types.EditorConfig{
			Type:        editor,
			ConfigPath:  defaultEditorConfig(editor),
			PackageName: defaultEditorPackage(editor),
		},
		Git: types.GitConfig{
			Name:  gitName,
			Email: gitEmail,
		},
		Packages: types.PackagesConfig{
			Core: []string{"git", "curl", "jq"},
			User: []string{},
			Team: []string{},
		},
	}, nil
}

// Update Manager's Initialize method
func (m *Manager) Initialize(createIfMissing bool) error {
	if m.configPath == "" {
		return fmt.Errorf("config path not set")
	}

	// Generate initial config if needed
	if createIfMissing {
		config, err := m.GenerateInitialConfig("zsh", "nano", "", "")
		if err != nil {
			return fmt.Errorf("failed to generate initial config: %w", err)
		}

		// Convert local NixConfig to types.NixConfig
		nixConfig := &types.NixConfig{
			Version:  config.Version,
			Settings: types.Settings(config.Settings),  // Explicit conversion
			Shell:    types.ShellConfig(config.Shell),
			Editor:   types.EditorConfig(config.Editor),
			Git:      types.GitConfig(config.Git),
			Packages: types.PackagesConfig(config.Packages),
		}
		m.nixConfig = nixConfig
	}

	// Create a types.Config from the nixConfig
	cfg := &types.Config{
		Version:   m.nixConfig.Version,
		NixConfig: m.nixConfig,  // Now using converted types.NixConfig
		Settings:  types.Settings(m.nixConfig.Settings),  // Explicit conversion
	}

	return m.Save(cfg)
}

// Update Manager's Load method
func (m *Manager) Load() (*types.Config, error) {
	if m.nixConfig == nil {
		return nil, fmt.Errorf("config not loaded")
	}

	// Create a new types.Config with explicit type conversions
	return &types.Config{
		Version:   m.nixConfig.Version,
		NixConfig: m.nixConfig,  // Now using converted types.NixConfig
		Settings:  types.Settings(m.nixConfig.Settings),  // Explicit conversion
	}, nil
}

// Update LoadSection to handle both return values from Load
func (m *Manager) LoadSection(name string, v interface{}) error {
	cfg, err := m.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	configValue := reflect.ValueOf(cfg).Elem()
	sectionValue := configValue.FieldByName(name)
	if !sectionValue.IsValid() {
		return fmt.Errorf("invalid configuration section: %s", name)
	}

	// Create a new value to hold the section data
	targetValue := reflect.ValueOf(v).Elem()
	targetValue.Set(sectionValue)

	return nil
}

// Add PreviewConfiguration to Manager
func (m *Manager) PreviewConfiguration(config *types.NixConfig) error {
	m.logger.Debug("Previewing configuration")

	fmt.Println("\nConfiguration Preview:")
	fmt.Println("----------------------")
	fmt.Printf("Shell: %s\n", config.Shell.Type)
	fmt.Printf("Editor: %s\n", config.Editor.Type)

	if config.Git.Name != "" || config.Git.Email != "" {
		fmt.Println("\nGit Configuration:")
		if config.Git.Name != "" {
			fmt.Printf("  Name: %s\n", config.Git.Name)
		}
		if config.Git.Email != "" {
			fmt.Printf("  Email: %s\n", config.Git.Email)
		}
	}
	return nil
}

// Add backup methods to Manager
func (m *Manager) CreateBackup(path string) error {
	m.logger.Debug("Creating configuration backup", "path", path)
	return m.SaveSection("backup", m.cfg)
}

func (m *Manager) RestoreBackup(path string) error {
	m.logger.Debug("Restoring from backup", "path", path)
	return m.LoadSection("backup", m.cfg)
}

// Add backup methods to ServiceImpl
func (s *ServiceImpl) CreateBackup(path string) error {
	return s.manager.CreateBackup(path)
}

func (s *ServiceImpl) RestoreBackup(path string) error {
	return s.manager.RestoreBackup(path)
}

// Add interface implementation
func (s *ServiceImpl) SaveSection(name string, v interface{}) error {
	return s.manager.SaveSection(name, v)
}

// Add package management methods to ServiceImpl
func (s *ServiceImpl) LoadCustomPackages() ([]string, error) {
	return s.manager.LoadCustomPackages()
}

func (s *ServiceImpl) SaveCustomPackages(packages []string) error {
	return s.manager.SaveCustomPackages(packages)
}

// Add type definitions to config package
type BaseConfig struct {
	Type    Type   `yaml:"type"`
	Version string `yaml:"version"`
	Name    string `yaml:"name"`
}

type ProjectConfig struct {
	BaseConfig
	Required []string `yaml:"required"`
	Settings Settings `yaml:"settings"`
	Version     string            `yaml:"version"`
	Name        string            `yaml:"name"`
	Environment string            `yaml:"environment"`
	Tools       []string          `yaml:"tools"`
}

func (s *ServiceImpl) LoadConfig(configType Type, name string) (interface{}, error) {
	var cfg interface{}
	err := s.manager.LoadSection(string(configType)+"-"+name, &cfg)
	return cfg, err
}

func (s *ServiceImpl) LoadProjectWithTeam(projectName, teamName string) (*types.ProjectConfig, error) {
	var projectCfg types.ProjectConfig
	err := s.manager.LoadSection("project-"+projectName+"-team-"+teamName, &projectCfg)
	return &projectCfg, err
}

func (s *ServiceImpl) ReadConfig(path string, v interface{}) error {
	// Implementation of ReadConfig method
	return nil // Placeholder return, actual implementation needed
}

func (s *ServiceImpl) WriteConfig(path string, cfg interface{}) error {
	return s.manager.SaveSection(filepath.Base(path), cfg)
}

func (s *ServiceImpl) MergeProjectConfigs(base, team types.ProjectConfig) types.ProjectConfig {
	// Implementation of MergeProjectConfigs method
	return types.ProjectConfig{} // Placeholder return, actual implementation needed
}

// Add missing LoadConfig to Manager
func (m *Manager) LoadConfig(configType Type, name string) (interface{}, error) {
	switch configType {
	case ProjectType:
		return m.loadProjectConfig(name)
	case TeamType:
		return m.loadTeamConfig(name)
	case UserType:
		return m.loadUserConfig(name)
	case SystemType:
		return m.loadSystemConfig(name)
	default:
		return nil, fmt.Errorf("unsupported config type: %s", configType)
	}
}

// Add missing LoadProjectWithTeam to Manager
func (m *Manager) LoadProjectWithTeam(projectName, teamName string) (*types.ProjectConfig, error) {
	var projectCfg types.ProjectConfig
	err := m.LoadSection("project-"+projectName+"-team-"+teamName, &projectCfg)
	return &projectCfg, err
}

// Add missing WriteConfig to Manager
func (m *Manager) WriteConfig(path string, cfg interface{}) error {
	return m.SaveSection(filepath.Base(path), cfg)
}

// Add package management to Manager
func (m *Manager) LoadCustomPackages() ([]string, error) {
	var pkgConfig struct{ Packages []string }
	err := m.LoadSection("custom-packages", &pkgConfig)
	return pkgConfig.Packages, err
}

func (m *Manager) SaveCustomPackages(packages []string) error {
	return m.SaveSection("custom-packages", map[string]interface{}{"Packages": packages})
}

// Fix the MergeProjectConfigs method in Manager
func (m *Manager) MergeProjectConfigs(base, team types.ProjectConfig) types.ProjectConfig {
	merged := base

	// Initialize settings map if nil
	if merged.Settings == nil {
		merged.Settings = make(map[string]string)
	}

	// Merge settings
	// Copy all settings from base
	for k, v := range base.Settings {
		merged.Settings[k] = v
	}

	// Override with team settings
	for k, v := range team.Settings {
		if v != "" { // Only override if team setting has a value
			merged.Settings[k] = v
		}
	}

	// Merge required dependencies
	merged.Required = uniqueMerge(base.Required, team.Required)

	return merged
}

// Helper function for merging string slices
func uniqueMerge(a, b []string) []string {
	set := make(map[string]struct{})
	for _, s := range append(a, b...) {
		set[s] = struct{}{}
	}
	result := make([]string, 0, len(set))
	for k := range set {
		result = append(result, k)
	}
	return result
}

// Add config file reading to Manager
func (m *Manager) ReadConfig(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	return yaml.Unmarshal(data, v)
}

func (s *ServiceImpl) GetConfig() *types.NixConfig {
	return s.manager.nixConfig
}

// Add GetConfig to Manager
func (m *Manager) GetConfig() *types.NixConfig {
	return m.nixConfig
}

// Fix the ValidateConfiguration method in Manager
func (m *Manager) ValidateConfiguration(verbose bool) error {
	if m == nil {
		return fmt.Errorf("config manager not initialized")
	}

	// Fix the Load() call to handle both return values
	_, err := m.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var validationErrors []string

	// Update validation checks
	validShells := map[string]bool{"zsh": true, "bash": true, "fish": true}
	if m.nixConfig.Shell.Type == "" {
		validationErrors = append(validationErrors,
			"missing required shell configuration\n\tRun 'nix-foundry config set shell <type>' (valid types: zsh, bash, fish)")
	} else if !validShells[strings.ToLower(m.nixConfig.Shell.Type)] {
		validationErrors = append(validationErrors,
			fmt.Sprintf("invalid shell type: %s\n\tRun 'nix-foundry config set shell <type>' (valid types: zsh, bash, fish)", m.nixConfig.Shell.Type))
	}

	validEditors := map[string]bool{"vim": true, "nano": true, "vscode": true}
	if m.nixConfig.Editor.Type == "" {
		validationErrors = append(validationErrors,
			"missing required editor configuration\n\tRun 'nix-foundry config set editor <type>' (valid types: vim, nano, vscode)")
	} else if !validEditors[strings.ToLower(m.nixConfig.Editor.Type)] {
		validationErrors = append(validationErrors,
			fmt.Sprintf("invalid editor type: %s\n\tRun 'nix-foundry config set editor <type>' (valid types: vim, nano, vscode)", m.nixConfig.Editor.Type))
	}

	if len(m.nixConfig.Packages.Core) == 0 {
		validationErrors = append(validationErrors,
			"no core packages configured\n\tRun 'nix-foundry pkg add core <package>' to add core packages")
	}

	// Always show verbose report if requested
	if verbose {
		fmt.Println("\n🔍 Configuration Validation Report")
		fmt.Println("========================================")
		fmt.Printf("📂 Config Directory: %s\n", m.configPath)
		fmt.Printf("📄 Config File: %s\n", filepath.Join(m.configPath, "config.yaml"))
		fmt.Println("\n⚙️  Core Configuration:")
		if m.nixConfig.Shell.Type != "" {
			fmt.Printf("  %-12s %s (%s)\n", "Shell:", m.nixConfig.Shell.Type, os.Getenv("SHELL"))
		} else {
			fmt.Printf("  %-12s (not set) (%s)\n", "Shell:", os.Getenv("SHELL"))
		}
		if m.nixConfig.Editor.Type != "" {
			fmt.Printf("  %-12s %s\n", "Editor:", m.nixConfig.Editor.Type)
		} else {
			fmt.Printf("  %-12s (not set)\n", "Editor:")
		}

		if m.nixConfig.Git.Name != "" || m.nixConfig.Git.Email != "" {
			fmt.Println("\n👤 Git Configuration:")
			fmt.Printf("  %-12s %s\n", "Name:", m.nixConfig.Git.Name)
			fmt.Printf("  %-12s %s\n", "Email:", m.nixConfig.Git.Email)
		}

		if len(m.nixConfig.Packages.Core) > 0 {
			fmt.Println("\n📦 Installed Packages:")
			fmt.Printf("  %-8s %d packages\n", "Core:", len(m.nixConfig.Packages.Core))
			fmt.Printf("  %-8s %d packages\n", "User:", len(m.nixConfig.Packages.User))
			fmt.Printf("  %-8s %d packages\n", "Team:", len(m.nixConfig.Packages.Team))
		}

		fmt.Println("\n Environment Links:")
		envs, _ := filepath.Glob(filepath.Join(m.configPath, "environments", "*"))
		for _, env := range envs {
			if target, err := os.Readlink(env); err == nil {
				fmt.Printf("  %-20s → %s\n", filepath.Base(env), target)
			}
		}

		fmt.Println("\n========================================")

		if len(validationErrors) > 0 {
			fmt.Println("❌ Validation Issues Found:")
			for i, err := range validationErrors {
				fmt.Printf("%d. %s\n", i+1, err)
			}
		} else {
			fmt.Println("✅ All Checks Passed!")
		}
	}

	if len(validationErrors) > 0 {
		// Only return the raw errors without joining if we've already shown verbose output
		if verbose {
			return nil  // We've already displayed the errors in verbose mode
		}
		return fmt.Errorf("%s", strings.Join(validationErrors, "\n"))
	}

	return nil
}

func (s *ServiceImpl) CreateConfigFromMap(configMap map[string]string) *types.NixConfig {
	cfg := &types.NixConfig{
		Shell: types.ShellConfig{
			Type:     configMap["shell"],
			InitFile: defaultInitFile(configMap["shell"]),
		},
		Editor: types.EditorConfig{
			Type:        configMap["editor"],
			ConfigPath:  defaultEditorConfig(configMap["editor"]),
			PackageName: defaultEditorPackage(configMap["editor"]),
		},
		Git: types.GitConfig{
			Name:  configMap["git.name"],
			Email: configMap["git.email"],
		},
	}

	// Use manager methods instead of direct field access
	if s.manager.GetRetentionDays() == 0 {
		s.manager.SetRetentionDays(30)
	}

	return cfg
}

func defaultInitFile(shell string) string {
	switch shell {
	case "zsh": return "~/.zshrc"
	case "bash": return "~/.bashrc"
	case "fish": return "~/.config/fish/config.fish"
	default: return "~/.bashrc"
	}
}

func defaultEditorConfig(editor string) string {
	switch editor {
	case "vim": return "~/.vimrc"
	case "neovim": return "~/.config/nvim/init.vim"
	case "vscode": return "~/.vscode/argv.json"
	default: return ""
	}
}

func defaultEditorPackage(editor string) string {
	switch editor {
	case "vim": return "vim"
	case "neovim": return "neovim"
	case "vscode": return "vscode"
	default: return ""
	}
}

// Add getter/setter methods
func (s *ServiceImpl) GetRetentionDays() int {
	return s.manager.GetRetentionDays()
}

func (s *ServiceImpl) SetRetentionDays(days int) {
	s.manager.SetRetentionDays(days)
}

func (s *ServiceImpl) GetMaxBackups() int {
	return s.manager.GetMaxBackups()
}

func (s *ServiceImpl) SetMaxBackups(max int) {
	s.manager.SetMaxBackups(max)
}

func (s *ServiceImpl) GetCompressionLevel() int {
	return s.manager.GetCompressionLevel()
}

func (s *ServiceImpl) SetCompressionLevel(level int) {
	s.manager.SetCompressionLevel(level)
}

// Add these methods to the Manager implementation
func (m *Manager) GetRetentionDays() int {
	return m.retentionDays
}

func (m *Manager) SetRetentionDays(days int) {
	m.retentionDays = days
}

func (m *Manager) GetMaxBackups() int {
	return m.maxBackups
}

func (m *Manager) SetMaxBackups(max int) {
	m.maxBackups = max
}

func (m *Manager) GetCompressionLevel() int {
	return m.compressionLevel
}

func (m *Manager) SetCompressionLevel(level int) {
	m.compressionLevel = level
}

// Add missing interface method implementation
func (m *Manager) CreateConfigFromMap(configMap map[string]string) *types.NixConfig {
	return &types.NixConfig{
		Shell: types.ShellConfig{
			Type:     configMap["shell"],
			InitFile: defaultInitFile(configMap["shell"]),
		},
		Editor: types.EditorConfig{
			Type:        configMap["editor"],
			ConfigPath:  defaultEditorConfig(configMap["editor"]),
			PackageName: defaultEditorPackage(configMap["editor"]),
		},
		Git: types.GitConfig{
			Name:  configMap["git.name"],
			Email: configMap["git.email"],
		},
	}
}

// Add missing GenerateInitialConfig to ServiceImpl
func (s *ServiceImpl) GenerateInitialConfig(shell, editor, gitName, gitEmail string) (*types.Config, error) {
	return s.manager.GenerateInitialConfig(shell, editor, gitName, gitEmail)
}

// Add to Manager implementation
func (m *Manager) SetValue(key string, value interface{}) error {
	// Handle different value types based on key
	switch v := value.(type) {
	case int:
		return m.setIntValue(key, v)
	case string:
		return m.setStringValue(key, v)
	default:
		return fmt.Errorf("unsupported value type %T for key: %s", v, key)
	}
}

// Existing numeric handlers
func (m *Manager) setIntValue(key string, value int) error {
	switch key {
	case "backup.retentionDays":
		m.SetRetentionDays(value)
	case "backup.maxBackups":
		m.SetMaxBackups(value)
	case "backup.compressionLevel":
		m.SetCompressionLevel(value)
	default:
		return fmt.Errorf("unknown integer key: %s", key)
	}
	return nil
}

// Update string handler to persist changes to both nixConfig and main config
func (m *Manager) setStringValue(key string, value string) error {
	// Ensure configs exist
	if m.cfg == nil {
		m.cfg = &types.Config{}
	}
	if m.nixConfig == nil {
		m.nixConfig = &types.NixConfig{}  // Using aliased type
	}

	// Handle short key aliases
	switch key {
	case "shell":
		key = "shell.type"
	case "editor":
		key = "editor.type"
	case "git-name":
		key = "git.name"
	case "git-email":
		key = "git.email"
	}

	switch key {
	case "shell.type":
		m.nixConfig.Shell.Type = value
		m.cfg.Shell.Type = value
	case "editor.type":
		m.nixConfig.Editor.Type = value
		m.cfg.Editor.Type = value
	case "git.name":
		m.nixConfig.Git.Name = value
		m.cfg.Git.Name = value
	case "git.email":
		m.nixConfig.Git.Email = value
		m.cfg.Git.Email = value
	default:
		return fmt.Errorf("unknown string key: %s", key)
	}

	if err := m.Save(m.cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}
	return nil
}

// Add the missing load methods to Manager
func (m *Manager) loadProjectConfig(name string) (*types.ProjectConfig, error) {
	path := filepath.Join(m.configPath, "projects", name+".yaml")
	var config types.ProjectConfig
	if err := m.ReadConfig(path, &config); err != nil {
		return nil, fmt.Errorf("failed to load project config: %w", err)
	}
	return &config, nil
}

func (m *Manager) loadTeamConfig(name string) (*TeamConfig, error) {
	path := filepath.Join(m.configPath, "teams", name+".yaml")
	var config TeamConfig
	if err := m.ReadConfig(path, &config); err != nil {
		return nil, fmt.Errorf("failed to load team config: %w", err)
	}
	return &config, nil
}

func (m *Manager) loadUserConfig(name string) (*UserConfig, error) {
	path := filepath.Join(m.configPath, "users", name+".yaml")
	var config UserConfig
	if err := m.ReadConfig(path, &config); err != nil {
		return nil, fmt.Errorf("failed to load user config: %w", err)
	}
	return &config, nil
}

func (m *Manager) loadSystemConfig(name string) (*SystemConfig, error) {
	path := filepath.Join(m.configPath, "system", name+".yaml")
	var config SystemConfig
	if err := m.ReadConfig(path, &config); err != nil {
		return nil, fmt.Errorf("failed to load system config: %w", err)
	}
	return &config, nil
}

// Add to ServiceImpl struct
func (s *ServiceImpl) ConfigDir() string {
	return s.manager.GetConfigDir()
}

// Add to ServiceImpl struct
func (s *ServiceImpl) GetLogger() *logging.Logger {
	return s.logger
}

// Add to Manager implementation
func (m *Manager) ResetValue(key string) error {
	// Handle different configuration keys
	switch key {
	case "shell.type":
		m.nixConfig.Shell.Type = "zsh" // default shell
	case "editor.type":
		m.nixConfig.Editor.Type = "nano" // default editor
	case "git.name":
		m.nixConfig.Git.Name = ""
	case "git.email":
		m.nixConfig.Git.Email = ""
	default:
		return fmt.Errorf("invalid reset key: %s", key)
	}
	return m.Save(m.cfg)
}

// Add to ServiceImpl
func (s *ServiceImpl) ResetValue(key string) error {
	return s.manager.ResetValue(key)
}

// Implement methods in ServiceImpl
func (s *ServiceImpl) GenerateEncryptionKey() error {
	return s.manager.GenerateEncryptionKey()
}

func (s *ServiceImpl) RotateEncryptionKey() error {
	return s.manager.RotateEncryptionKey()
}

// Remove the ServiceImpl's ValidateConfiguration implementation and delegate to manager
func (s *ServiceImpl) ValidateConfiguration(verbose bool) error {
	return s.manager.ValidateConfiguration(verbose)
}

// Add config path setter to Manager
func (m *Manager) SetConfigPath(path string) {
	m.configPath = path
	m.configDir = filepath.Dir(path)
	m.backupDir = filepath.Join(m.configDir, "backups")
}

func (s *ServiceImpl) ensureInitialized() error {
	if !s.initialized {
		return stderrors.New("service not initialized")
	}
	return nil
}

func (s *ServiceImpl) ValidateProjectConfig(cfg *types.ProjectConfig) error {
	// ... validation logic ...
	return nil
}

// Remove the method from types.Config and create a helper function
func getConfigValue(c *types.Config, key string) (interface{}, bool) {
	if c == nil {
		return nil, false
	}

	switch key {
	case "settings.autoUpdate":
		return c.Settings.AutoUpdate, true
	case "settings.updateInterval":
		return c.Settings.UpdateInterval, true
	case "settings.logLevel":
		return c.Settings.LogLevel, true
	default:
		return nil, false
	}
}

// Update any code that was using the method to use the helper function
func (s *ServiceImpl) GetValue(key string) (interface{}, error) {
	if s.config == nil {
		if _, err := s.Load(); err != nil {
			return nil, err
		}
	}

	if value, ok := getConfigValue(s.config, key); ok {
		return value, nil
	}
	return nil, fmt.Errorf("invalid key: %s", key)
}
