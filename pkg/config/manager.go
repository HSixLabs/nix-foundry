/*
Package config provides configuration management functionality for Nix Foundry.
It handles configuration loading, composition, and inheritance.
*/
package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/pkg/schema"
	"gopkg.in/yaml.v3"
)

/*
Manager handles configuration loading, composition, and inheritance.
It manages user, team, and project configurations, handling their relationships
and merging them according to inheritance rules.
*/
type Manager struct {
	userConfig    *schema.Config
	teamConfigs   map[string]*schema.Config
	projectConfig *schema.Config
	activeConfig  *schema.Config
}

/*
GetUserConfig returns the current user configuration.
*/
func (m *Manager) GetUserConfig() *schema.Config {
	return m.userConfig
}

/*
GetTeamConfig returns a team configuration by name.
*/
func (m *Manager) GetTeamConfig(name string) *schema.Config {
	return m.teamConfigs[name]
}

/*
GetProjectConfig returns the current project configuration.
*/
func (m *Manager) GetProjectConfig() *schema.Config {
	return m.projectConfig
}

/*
InstallPackage installs a package using the configured package manager.
It uses the active configuration to determine which package manager to use
and how to install the package.
*/
func (m *Manager) InstallPackage(pkg string) error {
	if m.activeConfig == nil {
		return fmt.Errorf("no active configuration")
	}

	var cmd *exec.Cmd
	switch m.activeConfig.Nix.Manager {
	case "nix-env":
		cmd = exec.Command("bash", "-c", fmt.Sprintf(
			". /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh && "+
				"NIXPKGS_ALLOW_UNFREE=1 NIXPKGS_ALLOW_UNSUPPORTED_SYSTEM=1 "+
				"/nix/var/nix/profiles/default/bin/nix-env -iA nixpkgs.%s -Q",
			pkg))
	default:
		return fmt.Errorf("unsupported package manager: %s", m.activeConfig.Nix.Manager)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install package %s: %w", pkg, err)
	}

	return nil
}

/*
ConfigureShell configures the shell based on the active configuration.
It sets up the shell environment with Nix-specific configurations and
ensures the shell's configuration directory exists.
*/
func (m *Manager) ConfigureShell() error {
	if m.activeConfig == nil {
		return fmt.Errorf("no active configuration")
	}

	if m.activeConfig.Settings.Shell == "" {
		return nil
	}

	homeDir, homeDirErr := os.UserHomeDir()
	if homeDirErr != nil {
		return fmt.Errorf("failed to get home directory: %w", homeDirErr)
	}

	var rcFile string
	switch m.activeConfig.Settings.Shell {
	case "bash":
		rcFile = filepath.Join(homeDir, ".bashrc")
	case "zsh":
		rcFile = filepath.Join(homeDir, ".zshrc")
	case "fish":
		rcFile = filepath.Join(homeDir, ".config", "fish", "config.fish")
	default:
		return fmt.Errorf("unsupported shell: %s", m.activeConfig.Settings.Shell)
	}

	if m.activeConfig.Settings.Shell == "fish" {
		if mkdirErr := os.MkdirAll(filepath.Dir(rcFile), 0755); mkdirErr != nil {
			return fmt.Errorf("failed to create fish config directory: %w", mkdirErr)
		}
	}

	var content string
	switch m.activeConfig.Settings.Shell {
	case "fish":
		content = `
# Nix
if test -e '/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.fish'
    source '/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.fish'
else if test -e "$HOME/.nix-profile/etc/profile.d/nix.fish"
    source "$HOME/.nix-profile/etc/profile.d/nix.fish"
end
`
	default:
		content = `
# Nix
if [ -e '/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh' ]; then
    . '/nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh'
elif [ -e "$HOME/.nix-profile/etc/profile.d/nix.sh" ]; then
    . "$HOME/.nix-profile/etc/profile.d/nix.sh"
fi
`
	}

	existingContent, readErr := os.ReadFile(rcFile)
	if readErr == nil && len(existingContent) > 0 {
		if strings.Contains(string(existingContent), "# Nix") {
			return nil
		}
	}

	f, openErr := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if openErr != nil {
		return fmt.Errorf("failed to open rc file: %w", openErr)
	}
	defer f.Close()

	if _, writeErr := f.WriteString(content); writeErr != nil {
		return fmt.Errorf("failed to update rc file: %w", writeErr)
	}

	return nil
}

/*
NewManager creates a new configuration manager with initialized maps.
*/
func NewManager() *Manager {
	return &Manager{
		teamConfigs: make(map[string]*schema.Config),
	}
}

/*
LoadUserConfig loads the user's configuration from disk.
If no configuration exists, it creates a default one.
*/
func (m *Manager) LoadUserConfig() error {
	configPath, pathErr := schema.GetConfigPath()
	if pathErr != nil {
		return fmt.Errorf("failed to get config path: %w", pathErr)
	}

	content, readErr := os.ReadFile(configPath)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			m.userConfig = schema.NewDefaultConfig()
			return nil
		}
		return fmt.Errorf("failed to read config: %w", readErr)
	}

	config := &schema.Config{}
	if unmarshalErr := yaml.Unmarshal(content, config); unmarshalErr != nil {
		return fmt.Errorf("failed to parse config: %w", unmarshalErr)
	}

	if config.Type != schema.UserConfig {
		return fmt.Errorf("invalid config type: expected user config, got %s", config.Type)
	}

	m.userConfig = config
	return nil
}

/*
LoadTeamConfig loads a team configuration by name from disk.
*/
func (m *Manager) LoadTeamConfig(name string) error {
	homeDir, homeDirErr := os.UserHomeDir()
	if homeDirErr != nil {
		return fmt.Errorf("failed to get home directory: %w", homeDirErr)
	}

	configPath := filepath.Join(homeDir, ".config", "nix-foundry", "teams", name+".yaml")
	content, readErr := os.ReadFile(configPath)
	if readErr != nil {
		return fmt.Errorf("failed to read team config: %w", readErr)
	}

	config := &schema.Config{}
	if unmarshalErr := yaml.Unmarshal(content, config); unmarshalErr != nil {
		return fmt.Errorf("failed to parse team config: %w", unmarshalErr)
	}

	if config.Type != schema.TeamConfig {
		return fmt.Errorf("invalid config type: expected team config, got %s", config.Type)
	}

	m.teamConfigs[name] = config
	return nil
}

/*
LoadProjectConfig loads a project configuration from the current directory.
*/
func (m *Manager) LoadProjectConfig() error {
	configPath := ".nix-foundry/config.yaml"
	content, readErr := os.ReadFile(configPath)
	if readErr != nil {
		if os.IsNotExist(readErr) {
			return nil
		}
		return fmt.Errorf("failed to read project config: %w", readErr)
	}

	config := &schema.Config{}
	if unmarshalErr := yaml.Unmarshal(content, config); unmarshalErr != nil {
		return fmt.Errorf("failed to parse project config: %w", unmarshalErr)
	}

	if config.Type != schema.ProjectConfig {
		return fmt.Errorf("invalid config type: expected project config, got %s", config.Type)
	}

	m.projectConfig = config
	return nil
}

/*
ComposeConfig creates the active configuration by merging configs based on inheritance.
It follows the inheritance chain from project to team to user config, with each
level overriding settings from the previous level.
*/
func (m *Manager) ComposeConfig() error {
	if m.userConfig == nil {
		return fmt.Errorf("no user config loaded")
	}

	result := m.userConfig

	if result.Base != "" {
		teamConfig, ok := m.teamConfigs[result.Base]
		if !ok {
			return fmt.Errorf("base team config %q not found", result.Base)
		}
		result = mergeConfigs(teamConfig, result)
	}

	if m.projectConfig != nil && (result.Base == m.projectConfig.Metadata.Name) {
		result = mergeConfigs(m.projectConfig, result)
	}

	m.activeConfig = result
	return nil
}

/*
GetActiveConfig returns the current active configuration.
*/
func (m *Manager) GetActiveConfig() *schema.Config {
	return m.activeConfig
}

/*
mergeConfigs merges two configurations, with the override config taking precedence.
*/
func mergeConfigs(base, override *schema.Config) *schema.Config {
	result := &schema.Config{
		Version:  override.Version,
		Kind:     override.Kind,
		Type:     override.Type,
		Base:     override.Base,
		Metadata: override.Metadata,
		Settings: mergeSettings(base.Settings, override.Settings),
		Nix:      mergeNix(base.Nix, override.Nix),
	}

	return result
}

/*
mergeSettings merges two settings objects, with override settings taking precedence.
*/
func mergeSettings(base, override schema.Settings) schema.Settings {
	result := base

	if override.Shell != "" {
		result.Shell = override.Shell
	}
	if override.LogLevel != "" {
		result.LogLevel = override.LogLevel
	}
	if override.UpdateInterval != 0 {
		result.UpdateInterval = override.UpdateInterval
	}
	result.AutoUpdate = override.AutoUpdate

	return result
}

/*
mergeNix merges two Nix configurations, combining their package lists and scripts.
*/
func mergeNix(base, override schema.Nix) schema.Nix {
	result := base

	if override.Manager != "" {
		result.Manager = override.Manager
	}

	result.Packages = mergePackages(base.Packages, override.Packages)
	result.Scripts = append(base.Scripts, override.Scripts...)

	return result
}

/*
mergePackages merges two package lists while maintaining uniqueness.
*/
func mergePackages(base, override schema.Packages) schema.Packages {
	result := schema.Packages{
		Core:     make([]string, 0),
		Optional: make([]string, 0),
	}

	coreMap := make(map[string]bool)
	optionalMap := make(map[string]bool)

	for _, pkg := range base.Core {
		coreMap[pkg] = true
	}
	for _, pkg := range base.Optional {
		optionalMap[pkg] = true
	}

	for _, pkg := range override.Core {
		coreMap[pkg] = true
	}
	for _, pkg := range override.Optional {
		optionalMap[pkg] = true
	}

	for pkg := range coreMap {
		result.Core = append(result.Core, pkg)
	}
	for pkg := range optionalMap {
		result.Optional = append(result.Optional, pkg)
	}

	return result
}
