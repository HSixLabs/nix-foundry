package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	"gopkg.in/yaml.v3"
)

// Service defines the interface for configuration operations
type Service interface {
	Initialize(testMode bool) error
	Load() error
	Save() error
	Apply(config *NixConfig, testMode bool) error
	GetConfig() *NixConfig
	GetConfigDir() string
	GetBackupDir() string
	LoadSection(name string, v interface{}) error
	SaveSection(name string, v interface{}) error
	ApplyFlags(flags map[string]string, force bool) error
	GenerateInitialConfig(shell, editor, gitName, gitEmail string) (*NixConfig, error)
	PreviewConfiguration(config *NixConfig) error
	ConfigExists() bool
	Generate(defaultEnv string, nixCfg *NixConfig) error
	CreateBackup(path string) error
	RestoreBackup(path string) error
	LoadConfig(configType ConfigType, name string) (interface{}, error)
	WriteConfig(path string, cfg interface{}) error
	MergeProjectConfigs(base, team ProjectConfig) ProjectConfig
	LoadProjectWithTeam(projectName, teamName string) (*ProjectConfig, error)
	ReadConfig(path string, v interface{}) error
	LoadCustomPackages() ([]string, error)
	SaveCustomPackages(packages []string) error
	ValidateConfiguration() error
	CreateConfigFromMap(configMap map[string]string) *NixConfig
	GetRetentionDays() int
	SetRetentionDays(days int)
	GetMaxBackups() int
	SetMaxBackups(max int)
	GetCompressionLevel() int
	SetCompressionLevel(level int)
	GetValue(key string) (interface{}, error)
	SetValue(key string, value interface{}) error
	Reset(section string) error
}

// ServiceImpl implements the configuration Service interface
type ServiceImpl struct {
	logger  *logging.Logger
	config  *Config
	path    string
	manager *Manager
}

// NixConfig represents the Nix configuration structure
type NixConfig struct {
	Shell struct {
		Type string
	}
	Editor struct {
		Type string
	}
	Git struct {
		Name  string
		Email string
	}
	Packages struct {
		Core []string `yaml:"core"`
		User []string `yaml:"user"`
		Team []string `yaml:"team"`
	} `yaml:"packages"`
}

// Update Manager struct definition
type Manager struct {
	logger           *logging.Logger
	cfg              *Config
	configPath       string
	nixConfig        *NixConfig
	retentionDays    int
	maxBackups       int
	compressionLevel int
}

// Add required methods to Manager
func (m *Manager) GetBackupDir() string {
	return filepath.Join(m.GetConfigDir(), "backups")
}

func (m *Manager) GetConfigDir() string {
	return filepath.Dir(m.configPath)
}

func (m *Manager) Save() error {
	data, err := yaml.Marshal(m.cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	return os.WriteFile(m.configPath, data, 0644)
}

// Implement missing Service interface methods
func (m *Manager) ApplyFlags(flags map[string]string, force bool) error {
	// Implementation logic for applying flags
	return nil
}

// Add ConfigExists method to Manager
func (m *Manager) ConfigExists() bool {
	if m.configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			m.logger.Error("Failed to get home directory", "error", err)
			return false
		}
		m.configPath = filepath.Join(home, ".config", "nix-foundry", "config.yaml")
	}

	_, err := os.Stat(m.configPath)
	return err == nil
}

// Update Manager struct initialization to set configPath
func NewManager() *Manager {
	return &Manager{
		logger:     logging.GetLogger(),
		cfg:        &Config{Version: "1.0.0"},
		configPath: filepath.Join(os.Getenv("HOME"), ".config", "nix-foundry", "config.yaml"),
	}
}

// NewService creates a new configuration service
func NewService() *ServiceImpl {
	mgr := NewManager() // Creates the Manager instance
	return &ServiceImpl{
		logger:  logging.GetLogger(),
		manager: mgr,
	}
}

func (s *ServiceImpl) Initialize(testMode bool) error {
	if testMode {
		// Test mode initialization logic
	} else {
		// Normal initialization
	}
	return nil
}

func (s *ServiceImpl) ApplyFlags(flags map[string]string, force bool) error {
	if !force && s.ConfigExists() {
		return errors.NewValidationError(s.path, nil,
			fmt.Sprintf("configuration already exists at %s", s.path))
	}

	// Convert flags to configuration
	for key, value := range flags {
		switch key {
		case "shell":
			s.config.Shell.Type = value
		case "editor":
			s.config.Editor.Type = value
		case "git-name":
			s.config.Git.Name = value
		case "git-email":
			s.config.Git.Email = value
		}
	}

	return s.Save()
}

func (s *ServiceImpl) PreviewConfiguration(config *NixConfig) error {
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

func (s *ServiceImpl) Generate(defaultEnv string, nixCfg *NixConfig) error {
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
    nixpkgs.url = "github:nixos/nixpkgs/nixos-23.11";
    home-manager.url = "github:nix-community/home-manager/release-23.11";
  };

  outputs = { nixpkgs, home-manager, ... }@inputs: {
    homeConfigurations = {
      "nix-foundry" = home-manager.lib.homeManagerConfiguration {
        pkgs = nixpkgs.legacyPackages.x86_64-linux;
        modules = [ ./home.nix ];
      };
    };
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

// Add the missing Apply method to the Manager type
func (m *Manager) Apply(config *NixConfig, testMode bool) error {
	m.logger.Info("Applying configuration changes",
		"testMode", testMode,
		"backupDir", m.GetBackupDir(),
		"configDir", m.GetConfigDir())

	if testMode {
		m.logger.Debug("Running in test mode - no changes persisted")
		return nil
	}

	// Actual implementation would generate files and run nix commands
	return m.Save()
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
func (s *ServiceImpl) Apply(config *NixConfig, testMode bool) error {
	return s.manager.Apply(config, testMode)
}

// Add missing Generate method to Manager
func (m *Manager) Generate(defaultEnv string, nixCfg *NixConfig) error {
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
func (m *Manager) GenerateInitialConfig(shell, editor, gitName, gitEmail string) (*NixConfig, error) {
	m.logger.Debug("Generating initial configuration",
		"shell", shell,
		"editor", editor,
		"gitName", gitName,
		"gitEmail", gitEmail)

	config := &NixConfig{
		Shell:  struct{ Type string }{Type: shell},
		Editor: struct{ Type string }{Type: editor},
		Git: struct {
			Name  string
			Email string
		}{
			Name:  gitName,
			Email: gitEmail,
		},
	}
	return config, nil
}

// Add Initialize method to Manager
func (m *Manager) Initialize(configDir string) error {
	m.configPath = filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		m.logger.Debug("Creating initial config.yaml in manager")
		return m.Save()
	}
	return nil
}

// Add Load method to Manager
func (m *Manager) Load() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	return yaml.Unmarshal(data, m.cfg)
}

// Add LoadSection to Manager
func (m *Manager) LoadSection(name string, v interface{}) error {
	if err := m.Load(); err != nil {
		return err
	}
	// Implementation similar to ServiceImpl's LoadSection
	configValue := reflect.ValueOf(m.cfg).Elem()
	sectionValue := configValue.FieldByName(name)
	if !sectionValue.IsValid() {
		return fmt.Errorf("invalid configuration section: %s", name)
	}
	targetValue := reflect.ValueOf(v).Elem()
	targetValue.Set(sectionValue)
	return nil
}

// Add PreviewConfiguration to Manager
func (m *Manager) PreviewConfiguration(config *NixConfig) error {
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
type ConfigType string

const (
	ProjectConfigType ConfigType = "project"
	TeamConfigType    ConfigType = "team"
)

type BaseConfig struct {
	Type    ConfigType `yaml:"type"`
	Version string     `yaml:"version"`
	Name    string     `yaml:"name"`
}

type ProjectConfig struct {
	BaseConfig
	Required []string `yaml:"required"`
	Settings Settings `yaml:"settings"`
}

func (s *ServiceImpl) LoadConfig(configType ConfigType, name string) (interface{}, error) {
	var cfg interface{}
	err := s.manager.LoadSection(string(configType)+"-"+name, &cfg)
	return cfg, err
}

func (s *ServiceImpl) LoadProjectWithTeam(projectName, teamName string) (*ProjectConfig, error) {
	var projectCfg ProjectConfig
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

func (s *ServiceImpl) MergeProjectConfigs(base, team ProjectConfig) ProjectConfig {
	// Implementation of MergeProjectConfigs method
	return ProjectConfig{} // Placeholder return, actual implementation needed
}

// Add missing LoadConfig to Manager
func (m *Manager) LoadConfig(configType ConfigType, name string) (interface{}, error) {
	var cfg interface{}
	err := m.LoadSection(string(configType)+"-"+name, &cfg)
	return cfg, err
}

// Add missing LoadProjectWithTeam to Manager
func (m *Manager) LoadProjectWithTeam(projectName, teamName string) (*ProjectConfig, error) {
	var projectCfg ProjectConfig
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

// Add config merging capability to Manager
func (m *Manager) MergeProjectConfigs(base, team ProjectConfig) ProjectConfig {
	merged := base

	// Merge settings
	merged.Settings.AutoUpdate = base.Settings.AutoUpdate || team.Settings.AutoUpdate
	if team.Settings.UpdateInterval != "" {
		merged.Settings.UpdateInterval = team.Settings.UpdateInterval
	} else {
		merged.Settings.UpdateInterval = base.Settings.UpdateInterval
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

func (s *ServiceImpl) GetConfig() *NixConfig {
	return s.manager.GetConfig()
}

// Add GetConfig to Manager
func (m *Manager) GetConfig() *NixConfig {
	return m.nixConfig
}

// Add encryption key methods to Manager
func (m *Manager) GenerateEncryptionKey() error {
	m.logger.Info("Generating new encryption key")
	// Implementation would generate and store encryption key
	return nil
}

func (m *Manager) RotateEncryptionKey() error {
	m.logger.Info("Rotating encryption key")
	// Implementation would rotate existing encryption key
	return nil
}

// Add ValidateConfiguration implementation to Manager
func (m *Manager) ValidateConfiguration() error {
	m.logger.Debug("Validating configuration")

	// Check required fields
	if m.nixConfig.Shell.Type == "" {
		return fmt.Errorf("missing required shell type")
	}

	if m.nixConfig.Git.Name == "" || m.nixConfig.Git.Email == "" {
		return fmt.Errorf("git user identity must be configured")
	}

	// Validate package lists
	if err := validatePackages(m.nixConfig.Packages.Core); err != nil {
		return fmt.Errorf("core packages: %w", err)
	}
	if err := validatePackages(m.nixConfig.Packages.User); err != nil {
		return fmt.Errorf("user packages: %w", err)
	}

	return nil
}

func validatePackages(packages []string) error {
	for _, pkg := range packages {
		if strings.Contains(pkg, " ") {
			return fmt.Errorf("invalid package name '%s' - contains spaces", pkg)
		}
	}
	return nil
}

// Add ValidateConfiguration to ServiceImpl
func (s *ServiceImpl) ValidateConfiguration() error {
	return s.manager.ValidateConfiguration()
}

func (s *ServiceImpl) CreateConfigFromMap(configMap map[string]string) *NixConfig {
	// Implementation to convert map to NixConfig
	return nil // Placeholder return, actual implementation needed
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
func (m *Manager) CreateConfigFromMap(configMap map[string]string) *NixConfig {
	// Implementation that converts map to NixConfig
	return &NixConfig{
		Shell:  struct{ Type string }{Type: configMap["shell"]},
		Editor: struct{ Type string }{Type: configMap["editor"]},
		Git: struct {
			Name  string
			Email string
		}{
			Name:  configMap["git.name"],
			Email: configMap["git.email"],
		},
	}
}

// Add missing GenerateInitialConfig to ServiceImpl
func (s *ServiceImpl) GenerateInitialConfig(shell, editor, gitName, gitEmail string) (*NixConfig, error) {
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

// Update string handler to use the helper
func (m *Manager) setStringValue(key string, value string) error {
	path := strings.Split(key, ".")
	if len(path) < 2 {
		return fmt.Errorf("invalid key format, use dot notation: %s", key)
	}

	// Convert path elements to PascalCase for struct fields
	for i := range path {
		path[i] = strings.Title(path[i])
	}

	return setNestedValue(&m.cfg, path, value)
}
