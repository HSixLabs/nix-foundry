/*
Package config provides configuration management functionality for Nix Foundry.
It handles all aspects of configuration including initialization, saving,
applying, and merging of configurations across different scopes (user, team, project).
*/
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/pkg/filesystem"
	"github.com/shawnkhoffman/nix-foundry/pkg/schema"
	"gopkg.in/yaml.v3"
)

/*
Service provides configuration management functionality for Nix Foundry.
It handles all aspects of configuration including initialization, saving,
applying, and merging of configurations across different scopes (user, team, project).
*/
type Service struct {
	fs filesystem.FileSystem
}

/*
NewService creates a new configuration service with the provided filesystem implementation.
The filesystem abstraction allows for flexible storage backends and easier testing.
*/
func NewService(fs filesystem.FileSystem) *Service {
	return &Service{
		fs: fs,
	}
}

/*
InitConfig initializes a new user configuration with default settings.
It creates the necessary directory structure and configuration file if they don't
exist. Returns an error if the configuration already exists or if there are any
filesystem operations failures.
*/
func (s *Service) InitConfig() error {
	config := schema.NewDefaultConfig()

	configPath, pathErr := schema.GetConfigPath()
	if pathErr != nil {
		return fmt.Errorf("failed to get config path: %w", pathErr)
	}

	configDir := filepath.Dir(configPath)
	if mkdirErr := s.fs.MkdirAll(configDir, 0775); mkdirErr != nil {
		return fmt.Errorf("failed to create config directory: %w", mkdirErr)
	}

	if s.fs.Exists(configPath) {
		return fmt.Errorf("config file already exists at %s", configPath)
	}

	content, marshalErr := yaml.Marshal(config)
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal config: %w", marshalErr)
	}

	if writeErr := s.fs.WriteFile(configPath, content, 0664); writeErr != nil {
		return fmt.Errorf("failed to write config: %w", writeErr)
	}

	return nil
}

/*
SaveConfig persists a configuration to disk in the appropriate location based on its type.
The configuration is saved as YAML and the appropriate directory structure is created
if it doesn't exist. The location depends on the configuration type:
- UserConfig: ~/.config/nix-foundry/config.yaml
- TeamConfig: ~/.config/nix-foundry/teams/<name>.yaml
- ProjectConfig: ./.nix-foundry/config.yaml
*/
func (s *Service) SaveConfig(config *schema.Config) error {
	var configPath string

	switch config.Type {
	case schema.UserConfig:
		var pathErr error
		configPath, pathErr = schema.GetConfigPath()
		if pathErr != nil {
			return fmt.Errorf("failed to get config path: %w", pathErr)
		}

	case schema.TeamConfig:
		userHomeDir, homeDirErr := os.UserHomeDir()
		if homeDirErr != nil {
			return fmt.Errorf("failed to get home directory: %w", homeDirErr)
		}
		configPath = filepath.Join(userHomeDir, ".config", "nix-foundry", "teams", config.Metadata.Name+".yaml")

	case schema.ProjectConfig:
		configPath = filepath.Join(".nix-foundry", "config.yaml")

	default:
		return fmt.Errorf("invalid config type: %s", config.Type)
	}

	configDir := filepath.Dir(configPath)
	if mkdirErr := s.fs.MkdirAll(configDir, 0775); mkdirErr != nil {
		return fmt.Errorf("failed to create config directory: %w", mkdirErr)
	}

	content, marshalErr := yaml.Marshal(config)
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal config: %w", marshalErr)
	}

	if writeErr := s.fs.WriteFile(configPath, content, 0664); writeErr != nil {
		return fmt.Errorf("failed to write config: %w", writeErr)
	}

	return nil
}

/*
ApplyConfig applies the active configuration to the system. This includes:
1. Configuring the shell environment if specified in user config
2. Managing packages (installing new ones, removing old ones)
3. Running any configured scripts

Returns an error if any step of the application process fails.
*/
func (s *Service) ApplyConfig() error {
	activeConfig, configErr := s.GetActiveConfig()
	if configErr != nil {
		return fmt.Errorf("failed to get active config: %w", configErr)
	}

	if activeConfig.Type == schema.UserConfig && activeConfig.Settings.Shell != "" {
		if shellErr := s.configureShell(activeConfig.Settings.Shell); shellErr != nil {
			return fmt.Errorf("failed to configure shell: %w", shellErr)
		}
	}

	if pkgErr := s.managePackages(activeConfig); pkgErr != nil {
		return fmt.Errorf("failed to manage packages: %w", pkgErr)
	}

	if scriptErr := s.runScripts(activeConfig); scriptErr != nil {
		return fmt.Errorf("failed to run scripts: %w", scriptErr)
	}

	return nil
}

/*
configureShell configures the specified shell with Nix environment settings.
It creates the appropriate shell configuration file (.bashrc, .zshrc, or config.fish)
and adds the necessary Nix initialization commands. If the shell configuration
already contains Nix initialization, it skips the modification.
*/
func (s *Service) configureShell(shell string) error {
	userHomeDir, homeDirErr := os.UserHomeDir()
	if homeDirErr != nil {
		return fmt.Errorf("failed to get home directory: %w", homeDirErr)
	}

	var rcFile string
	switch shell {
	case "bash":
		rcFile = filepath.Join(userHomeDir, ".bashrc")
	case "zsh":
		rcFile = filepath.Join(userHomeDir, ".zshrc")
	case "fish":
		rcFile = filepath.Join(userHomeDir, ".config", "fish", "config.fish")
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}

	if shell == "fish" {
		if mkdirErr := s.fs.MkdirAll(filepath.Dir(rcFile), 0775); mkdirErr != nil {
			return fmt.Errorf("failed to create fish config directory: %w", mkdirErr)
		}
	}

	var content string
	switch shell {
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

	existingContent, readErr := s.fs.ReadFile(rcFile)
	if readErr == nil && len(existingContent) > 0 {
		if strings.Contains(string(existingContent), "# Nix") {
			return nil
		}
	}

	if writeErr := s.fs.WriteFile(rcFile, []byte(content), 0664); writeErr != nil {
		return fmt.Errorf("failed to write shell config: %w", writeErr)
	}

	return nil
}

/*
installPackage installs a single package using nix-env.
It configures the environment to allow unfree and unsupported system packages,
and streams the installation output to the user.
*/
func (s *Service) installPackage(pkg string) error {
	cmd := exec.Command("bash", "-c", fmt.Sprintf(
		". /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh && "+
			"NIXPKGS_ALLOW_UNFREE=1 NIXPKGS_ALLOW_UNSUPPORTED_SYSTEM=1 "+
			"/nix/var/nix/profiles/default/bin/nix-env -iA nixpkgs.%s -Q",
		pkg))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err := cmd.Run()
	if err != nil && s.isPermissionError(err) {
		fmt.Println("\n⚠️  INSTALLATION FAILED - PERMISSION DENIED!")
		fmt.Println("This is likely because Nix doesn't have Full Disk Access permission on macOS.")
		fmt.Println("To fix this:")
		fmt.Println("1. Open System Preferences → Privacy & Security → Full Disk Access")
		fmt.Println("2. Click the '+' button and add 'nix'")
		fmt.Println("3. Re-run 'nix-foundry config apply' after granting permission")
		fmt.Println()
	} else if err == nil && runtime.GOOS == "darwin" {
		if symlinkErr := s.symlinkMacOSApps(pkg); symlinkErr != nil {
			fmt.Printf("Warning: Failed to symlink %s to Applications: %v\n", pkg, symlinkErr)
		}
	}
	return err
}

/*
runScripts executes all scripts defined in the configuration.
Each script is run using bash, with stdout and stderr connected to
allow the user to see the execution output.
*/
func (s *Service) runScripts(config *schema.Config) error {
	for _, script := range config.Nix.Scripts {
		cmd := exec.Command("bash", "-c", string(script.Commands))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if execErr := cmd.Run(); execErr != nil {
			return fmt.Errorf("failed to run script %s: %w", script.Name, execErr)
		}
	}
	return nil
}

/*
managePackages handles the complete package management lifecycle.
It queries currently installed packages using nix-env -q, compares with the desired
configuration, and installs/removes packages as needed.
*/
func (s *Service) managePackages(config *schema.Config) error {
	installedPackages, queryErr := s.getInstalledPackages()
	if queryErr != nil {
		return fmt.Errorf("failed to query installed packages: %w", queryErr)
	}

	diff := schema.DiffPackages(installedPackages, config.Nix.Packages)

	if len(diff.ToRemove) > 0 {
		fmt.Printf("Removing %d packages...\n", len(diff.ToRemove))
		for _, pkg := range diff.ToRemove {
			if removeErr := s.removePackage(pkg); removeErr != nil {
				return fmt.Errorf("failed to remove package %s: %w", pkg, removeErr)
			}
		}
	}

	if len(diff.ToInstall) > 0 {
		fmt.Printf("Installing %d packages...\n", len(diff.ToInstall))
		for _, pkg := range diff.ToInstall {
			if installErr := s.installPackage(pkg); installErr != nil {
				s.handlePackageInstallationFailure(pkg, installErr)
				fmt.Printf("⚠️  Skipping %s due to installation failure\n", pkg)
				continue
			}
		}
	}

	if len(diff.ToInstall) == 0 && len(diff.ToRemove) == 0 {
		fmt.Println("No package changes needed")
	}

	return nil
}

/*
handlePackageInstallationFailure provides helpful error messages and suggestions
for common package installation failures, without hard-coding package-specific logic.
*/
func (s *Service) handlePackageInstallationFailure(pkg string, err error) {
	errorStr := err.Error()

	fmt.Printf("❌ Failed to install %s: %v\n", pkg, err)

	if strings.Contains(errorStr, "Operation not permitted") {
		fmt.Println("💡 Suggestions:")
		fmt.Println("   • This package may have build issues on your system")
		fmt.Println("   • Consider installing manually or using an alternative package manager")
		fmt.Println("   • Check if the package is available through other sources")
	}

	if strings.Contains(errorStr, "unfree license") {
		fmt.Println("💡 Suggestion: This package requires unfree license acceptance")
		fmt.Println("   • The installer already sets NIXPKGS_ALLOW_UNFREE=1")
		fmt.Println("   • You may need to accept the license manually")
	}

	fmt.Println()
}

/*
removePackage removes a single package using nix-env.
It configures the environment and streams the removal output to the user.
*/
func (s *Service) removePackage(pkg string) error {
	fmt.Printf("Removing package: %s\n", pkg)
	cmd := exec.Command("bash", "-c", fmt.Sprintf(
		". /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh && "+
			"/nix/var/nix/profiles/default/bin/nix-env -e %s",
		pkg))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err := cmd.Run()
	if err == nil && runtime.GOOS == "darwin" {
		if cleanupErr := s.cleanupMacOSAppSymlinks(pkg); cleanupErr != nil {
			fmt.Printf("Warning: Failed to cleanup symlinks for %s: %v\n", pkg, cleanupErr)
		}
	}
	return err
}

/*
getInstalledPackages queries nix-env to get a list of currently installed packages.
Uses JSON output for accurate package name parsing, avoiding issues with compound package names.
*/
func (s *Service) getInstalledPackages() ([]string, error) {
	cmd := exec.Command("bash", "-c",
		". /nix/var/nix/profiles/default/etc/profile.d/nix-daemon.sh && "+
			"/nix/var/nix/profiles/default/bin/nix-env -q --json")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to query installed packages: %w", err)
	}

	var packages map[string]map[string]interface{}
	if jsonErr := json.Unmarshal(output, &packages); jsonErr != nil {
		return nil, fmt.Errorf("failed to parse package JSON: %w", jsonErr)
	}

	var packageNames []string
	for _, pkg := range packages {
		if pname, ok := pkg["pname"].(string); ok {
			packageNames = append(packageNames, pname)
		}
	}

	return packageNames, nil
}

/*
ListConfigs returns a list of all available configurations across all scopes.
It searches for and loads:
1. User configuration from ~/.config/nix-foundry/config.yaml
2. Team configurations from ~/.config/nix-foundry/teams/
3. Project configuration from ./.nix-foundry/config.yaml

Returns the list of found configurations and any error encountered during the search.
*/
func (s *Service) ListConfigs() ([]*schema.Config, error) {
	var configs []*schema.Config

	configPath, pathErr := schema.GetConfigPath()
	if pathErr != nil {
		return nil, fmt.Errorf("failed to get config path: %w", pathErr)
	}

	if s.fs.Exists(configPath) {
		fileContent, readErr := s.fs.ReadFile(configPath)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read user config: %w", readErr)
		}

		userConfig := &schema.Config{}
		if unmarshalErr := yaml.Unmarshal(fileContent, userConfig); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to parse user config: %w", unmarshalErr)
		}

		configs = append(configs, userConfig)
	}

	userHomeDir, homeDirErr := os.UserHomeDir()
	if homeDirErr != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", homeDirErr)
	}

	teamsDir := filepath.Join(userHomeDir, ".config", "nix-foundry", "teams")
	if s.fs.Exists(teamsDir) {
		entries, readDirErr := os.ReadDir(teamsDir)
		if readDirErr != nil {
			return nil, fmt.Errorf("failed to read teams directory: %w", readDirErr)
		}

		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
				fileContent, readErr := s.fs.ReadFile(filepath.Join(teamsDir, entry.Name()))
				if readErr != nil {
					continue
				}

				teamConfig := &schema.Config{}
				if unmarshalErr := yaml.Unmarshal(fileContent, teamConfig); unmarshalErr != nil {
					continue
				}

				configs = append(configs, teamConfig)
			}
		}
	}

	projectConfigPath := ".nix-foundry/config.yaml"
	if s.fs.Exists(projectConfigPath) {
		fileContent, readErr := s.fs.ReadFile(projectConfigPath)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read project config: %w", readErr)
		}

		projectConfig := &schema.Config{}
		if unmarshalErr := yaml.Unmarshal(fileContent, projectConfig); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to parse project config: %w", unmarshalErr)
		}

		configs = append(configs, projectConfig)
	}

	return configs, nil
}

/*
GetActiveConfig returns the active configuration for the current context.
It performs the following steps:
1. Loads the user configuration
2. If the user config extends a team config, merges them
3. If the resulting config extends a project config, merges that as well

The merging follows the override principle where later configs take precedence
over earlier ones in the chain.
*/
func (s *Service) GetActiveConfig() (*schema.Config, error) {
	userConfig := schema.NewDefaultConfig()
	configPath, pathErr := schema.GetConfigPath()
	if pathErr != nil {
		return nil, fmt.Errorf("failed to get config path: %w", pathErr)
	}

	if s.fs.Exists(configPath) {
		fileContent, readErr := s.fs.ReadFile(configPath)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read user config: %w", readErr)
		}

		if unmarshalErr := yaml.Unmarshal(fileContent, userConfig); unmarshalErr != nil {
			return nil, fmt.Errorf("failed to parse user config: %w", unmarshalErr)
		}
	}

	if userConfig.Base != "" {
		teamConfig, teamErr := s.GetConfig(schema.TeamConfig, userConfig.Base)
		if teamErr != nil {
			return nil, fmt.Errorf("failed to get team config: %w", teamErr)
		}
		userConfig = s.mergeConfigs(teamConfig, userConfig)
	}

	projectConfig, projectErr := s.GetConfig(schema.ProjectConfig, "")
	if projectErr == nil && userConfig.Base == projectConfig.Metadata.Name {
		userConfig = s.mergeConfigs(projectConfig, userConfig)
	}

	return userConfig, nil
}

/*
mergeConfigs merges two configurations, with the override configuration taking precedence.
It handles merging of all configuration aspects including metadata, settings,
and Nix-specific configurations while preserving the override hierarchy.
*/
func (s *Service) mergeConfigs(base, override *schema.Config) *schema.Config {
	result := &schema.Config{
		Version:  override.Version,
		Kind:     override.Kind,
		Type:     override.Type,
		Base:     override.Base,
		Metadata: override.Metadata,
		Settings: s.mergeSettings(base.Settings, override.Settings),
		Nix:      s.mergeNix(base.Nix, override.Nix),
	}
	return result
}

/*
mergeSettings merges two settings objects, with override settings taking precedence.
It handles each setting individually, allowing for granular control over which
settings are inherited and which are overridden.
*/
func (s *Service) mergeSettings(base, override schema.Settings) schema.Settings {
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
It preserves the override's manager setting if specified, and concatenates
script lists from both configurations.
*/
func (s *Service) mergeNix(base, override schema.Nix) schema.Nix {
	result := base
	if override.Manager != "" {
		result.Manager = override.Manager
	}
	result.Packages = s.mergePackages(base.Packages, override.Packages)
	result.Scripts = append(base.Scripts, override.Scripts...)
	return result
}

/*
mergePackages merges two package lists while maintaining uniqueness.
It handles both core and optional packages, ensuring no duplicates exist
in the final package lists. The function uses maps for efficient deduplication
before converting back to slices for the final result.
*/
func (s *Service) mergePackages(base, override schema.Packages) schema.Packages {
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

/*
GetConfig retrieves a specific configuration by type and name.
It handles loading of user, team, and project configurations from their
respective locations in the filesystem. Returns an error if the configuration
cannot be found or cannot be parsed.
*/
func (s *Service) GetConfig(configType schema.ConfigType, name string) (*schema.Config, error) {
	var configPath string

	switch configType {
	case schema.UserConfig:
		var pathErr error
		configPath, pathErr = schema.GetConfigPath()
		if pathErr != nil {
			return nil, fmt.Errorf("failed to get config path: %w", pathErr)
		}

	case schema.TeamConfig:
		userHomeDir, homeDirErr := os.UserHomeDir()
		if homeDirErr != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", homeDirErr)
		}
		configPath = filepath.Join(userHomeDir, ".config", "nix-foundry", "teams", name+".yaml")

	case schema.ProjectConfig:
		configPath = filepath.Join(".nix-foundry", "config.yaml")

	default:
		return nil, fmt.Errorf("invalid config type: %s", configType)
	}

	if !s.fs.Exists(configPath) {
		return nil, fmt.Errorf("config file not found at %s", configPath)
	}

	fileContent, readErr := s.fs.ReadFile(configPath)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read config: %w", readErr)
	}

	config := &schema.Config{}
	if unmarshalErr := yaml.Unmarshal(fileContent, config); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to parse config: %w", unmarshalErr)
	}

	return config, nil
}

/*
UninstallConfig removes all Nix Foundry configuration files and directories.
This includes:
1. User configuration directory (~/.config/nix-foundry)
2. Project configuration directory (./.nix-foundry)

Returns an error if any deletion operation fails.
*/
func (s *Service) UninstallConfig() error {
	userHomeDir, homeDirErr := os.UserHomeDir()
	if homeDirErr != nil {
		return fmt.Errorf("failed to get home directory: %w", homeDirErr)
	}

	configDir := filepath.Join(userHomeDir, ".config", "nix-foundry")
	if removeErr := s.fs.Remove(configDir); removeErr != nil {
		return fmt.Errorf("failed to remove config directory: %w", removeErr)
	}

	projectConfigPath := ".nix-foundry"
	if s.fs.Exists(projectConfigPath) {
		if removeErr := s.fs.Remove(projectConfigPath); removeErr != nil {
			return fmt.Errorf("failed to remove project config: %w", removeErr)
		}
	}

	return nil
}

/*
InitConfigWithType initializes a new configuration of the specified type (team or project).
It creates the necessary directory structure and configuration file with default
settings appropriate for the specified type. Returns an error if the configuration
already exists or if there are any filesystem operation failures.
*/
func (s *Service) InitConfigWithType(configType schema.ConfigType, name string) error {
	var config *schema.Config

	switch configType {
	case schema.TeamConfig:
		config = schema.NewTeamConfig(name)
	case schema.ProjectConfig:
		config = schema.NewProjectConfig(name)
	default:
		return fmt.Errorf("invalid config type: %s", configType)
	}

	var configPath string

	switch configType {
	case schema.TeamConfig:
		userHomeDir, homeDirErr := os.UserHomeDir()
		if homeDirErr != nil {
			return fmt.Errorf("failed to get home directory: %w", homeDirErr)
		}
		configPath = filepath.Join(userHomeDir, ".config", "nix-foundry", "teams", name+".yaml")

	case schema.ProjectConfig:
		configPath = filepath.Join(".nix-foundry", "config.yaml")
	}

	configDir := filepath.Dir(configPath)
	if mkdirErr := s.fs.MkdirAll(configDir, 0775); mkdirErr != nil {
		return fmt.Errorf("failed to create config directory: %w", mkdirErr)
	}

	if s.fs.Exists(configPath) {
		return fmt.Errorf("config file already exists at %s", configPath)
	}

	content, marshalErr := yaml.Marshal(config)
	if marshalErr != nil {
		return fmt.Errorf("failed to marshal config: %w", marshalErr)
	}

	if writeErr := s.fs.WriteFile(configPath, content, 0664); writeErr != nil {
		return fmt.Errorf("failed to write config: %w", writeErr)
	}

	return nil
}

/*
isPermissionError checks if an error is likely caused by macOS permission issues.
*/
func (s *Service) isPermissionError(err error) bool {
	if runtime.GOOS != "darwin" {
		return false
	}

	errorMsg := strings.ToLower(err.Error())
	return strings.Contains(errorMsg, "operation not permitted") ||
		strings.Contains(errorMsg, "permission denied") ||
		strings.Contains(errorMsg, "mkdir: /nix/store") ||
		strings.Contains(errorMsg, "cannot create directory")
}

/*
symlinkMacOSApps automatically creates symlinks for GUI applications in /Applications
so they appear in Launchpad and Finder. This searches for .app bundles in the
individual package store paths and symlinks them to the system Applications folder.
*/
func (s *Service) symlinkMacOSApps(pkg string) error {
	packageName := pkg
	if strings.Contains(pkg, ".") {
		parts := strings.Split(pkg, ".")
		packageName = parts[len(parts)-1]
	}
	
	cmd := exec.Command("find", "/nix/store", "-name", "*"+packageName+"*", "-type", "d", "-maxdepth", "1")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}
	
	storePaths := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, storePath := range storePaths {
		if storePath == "" {
			continue
		}
		
		err := filepath.Walk(storePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			
			if info.IsDir() && strings.HasSuffix(info.Name(), ".app") {
				appName := info.Name()
				targetPath := filepath.Join("/Applications", appName)
				
				if _, statErr := os.Lstat(targetPath); statErr == nil {
					return nil
				}
				
				if symlinkErr := os.Symlink(path, targetPath); symlinkErr != nil {
					fmt.Printf("\n📱 GUI App Installed: %s\n", appName)
					fmt.Printf("   To make it visible in Launchpad, run:\n")
					fmt.Printf("   sudo ln -sf \"%s\" \"%s\"\n", path, targetPath)
					fmt.Printf("   Or manually drag it from Finder to Applications folder\n\n")
				} else {
					fmt.Printf("✨ Symlinked %s to Applications for Launchpad visibility\n", appName)
				}
			}
			return nil
		})
		
		if err != nil {
			continue
		}
	}
	
	return nil
}

/*
cleanupMacOSAppSymlinks cleans up any symlinks for GUI applications after removal on macOS.
It searches for symlinks in /Applications that point to Nix store paths containing the package name.
*/
func (s *Service) cleanupMacOSAppSymlinks(pkg string) error {
	packageName := pkg
	if strings.Contains(pkg, ".") {
		parts := strings.Split(pkg, ".")
		packageName = parts[len(parts)-1]
	}
	
	applicationsDir := "/Applications"
	entries, err := os.ReadDir(applicationsDir)
	if err != nil {
		return nil
	}
	
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".app") {
			continue
		}
		
		entryPath := filepath.Join(applicationsDir, entry.Name())
		
		linkInfo, err := os.Lstat(entryPath)
		if err != nil || linkInfo.Mode()&os.ModeSymlink == 0 {
			continue
		}
		
		target, err := os.Readlink(entryPath)
		if err != nil {
			continue
		}
		
		if strings.Contains(target, "/nix/store/") && strings.Contains(target, packageName) {
			if removeErr := os.Remove(entryPath); removeErr != nil {
				fmt.Printf("Warning: Failed to remove symlink for %s: %v\n", entry.Name(), removeErr)
			} else {
				fmt.Printf("🗑️  Removed symlink for %s\n", entry.Name())
			}
		}
	}
	
	return nil
}
