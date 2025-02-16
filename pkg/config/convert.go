package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// ConvertYAMLToNix converts YAML configuration to Nix configuration
func ConvertYAMLToNix(yamlConfig []byte) (*NixConfig, error) {
	var cfg NixConfig
	if err := yaml.Unmarshal(yamlConfig, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate package names
	allPackages := make(map[string]bool)
	for _, pkg := range cfg.Packages.Additional {
		allPackages[pkg] = true
	}
	for _, pkgs := range cfg.Packages.PlatformSpecific {
		for _, pkg := range pkgs {
			allPackages[pkg] = true
		}
	}
	for _, pkg := range cfg.Packages.Development {
		allPackages[pkg] = true
	}
	for _, pkgs := range cfg.Packages.Team {
		for _, pkg := range pkgs {
			allPackages[pkg] = true
		}
	}

	// For now, skip package validation as it requires nixpkgs integration
	// TODO: Implement proper package validation against nixpkgs

	return &cfg, nil
}

// loadBaseConfig loads a base configuration file
// func loadBaseConfig(path string) (UserConfig, error) {
// 	var baseCfg UserConfig

// 	// First try absolute path
// 	data, err := os.ReadFile(path)
// 	if err != nil {
// 		// Then try relative to config directory
// 		configDir := filepath.Join(os.Getenv("HOME"), ".config", "nix-foundry", "bases")
// 		data, err = os.ReadFile(filepath.Join(configDir, path))
// 		if err != nil {
// 			return baseCfg, err
// 		}
// 	}

// 	if err := yaml.Unmarshal(data, &baseCfg); err != nil {
// 		return baseCfg, fmt.Errorf("failed to parse base config: %w", err)
// 	}

// 	// Recursively load base configurations
// 	if baseCfg.Extends != "" {
// 		parentCfg, err := loadBaseConfig(baseCfg.Extends)
// 		if err != nil {
// 			return baseCfg, fmt.Errorf("failed to load parent config %s: %w", baseCfg.Extends, err)
// 		}
// 		baseCfg = mergeConfigs(parentCfg, baseCfg)
// 	}

// 	return baseCfg, nil
// }

// mergeConfigs merges a base configuration with an extending configuration.
// Values from the extending config take precedence over the base config.
// func mergeConfigs(base, extend UserConfig) UserConfig {
// 	result := extend

// 	// If shell is not set in extending config, use base
// 	if result.Shell == "" {
// 		result.Shell = base.Shell
// 	}

// 	// Merge editors
// 	if len(result.Editors) == 0 {
// 		result.Editors = base.Editors
// 	} else {
// 		// Create a map of editors by type for efficient lookup
// 		baseEditors := make(map[string]EditorConfig)
// 		for _, editor := range base.Editors {
// 			baseEditors[editor.Type] = editor
// 		}

// 		// For each editor in the extending config
// 		for i, editor := range result.Editors {
// 			if baseEditor, exists := baseEditors[editor.Type]; exists {
// 				// Merge extensions
// 				extensions := make(map[string]bool)
// 				for _, ext := range baseEditor.Extensions {
// 					extensions[ext] = true
// 				}
// 				for _, ext := range editor.Extensions {
// 					extensions[ext] = true
// 				}

// 				merged := make([]string, 0, len(extensions))
// 				for ext := range extensions {
// 					merged = append(merged, ext)
// 				}
// 				result.Editors[i].Extensions = merged

// 				// Merge settings
// 				if result.Editors[i].Settings == nil {
// 					result.Editors[i].Settings = make(map[string]interface{})
// 				}
// 				for k, v := range baseEditor.Settings {
// 					if _, exists := result.Editors[i].Settings[k]; !exists {
// 						result.Editors[i].Settings[k] = v
// 					}
// 				}
// 			}
// 		}
// 	}

// 	// Merge packages
// 	packages := make(map[string]bool)
// 	for _, pkg := range base.Packages {
// 		packages[pkg] = true
// 	}
// 	for _, pkg := range result.Packages {
// 		packages[pkg] = true
// 	}
// 	result.Packages = make([]string, 0, len(packages))
// 	for pkg := range packages {
// 		result.Packages = append(result.Packages, pkg)
// 	}

// 	// Merge Git config
// 	if !result.Git.Enable {
// 		result.Git = base.Git
// 	} else if result.Git.Enable && base.Git.Enable {
// 		// Merge Git user info if not set
// 		if result.Git.User.Name == "" {
// 			result.Git.User.Name = base.Git.User.Name
// 		}
// 		if result.Git.User.Email == "" {
// 			result.Git.User.Email = base.Git.User.Email
// 		}

// 		// Merge Git config settings
// 		if result.Git.Config == nil {
// 			result.Git.Config = make(map[string]string)
// 		}
// 		for k, v := range base.Git.Config {
// 			if _, exists := result.Git.Config[k]; !exists {
// 				result.Git.Config[k] = v
// 			}
// 		}
// 	}

// 	return result
// }
