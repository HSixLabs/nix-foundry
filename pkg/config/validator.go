package config

import (
	"fmt"
	"strings"
)

var (
	ValidShells     = []string{"zsh", "bash", "fish"}
	ValidAutoShells = []string{"zsh", "bash"}
	ValidEditors    = []string{"nano", "vim", "nvim", "emacs", "neovim", "vscode", "code"}
	editorAliases   = map[string]string{
		"code": "vscode",
	}
)

type Validator struct {
	config *NixConfig
}

func NewValidator(config *NixConfig) *Validator {
	return &Validator{config: config}
}

func (v *Validator) ValidateConfig() error {
	if v.config.Version == "" {
		return fmt.Errorf("version is required")
	}

	// Keep core validation flow here
	validationSteps := []func() error{
		v.validateShell,
		v.validateEditor,
		v.validatePackages,
		v.validateGit,
		v.validateTeam,
	}

	for _, step := range validationSteps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

// Keep core validation methods here
func (v *Validator) validateShell() error {
	if v.config.Shell.Type == "" {
		return fmt.Errorf("shell type is required")
	}
	if !Contains(ValidShells, v.config.Shell.Type) {
		return fmt.Errorf("invalid shell type: %s", v.config.Shell.Type)
	}
	return nil
}

func (v *Validator) validateEditor() error {
	if v.config.Editor.Type == "" {
		return fmt.Errorf("editor type is required")
	}
	return ValidateEditor(v.config.Editor.Type)
}

// Keep helper functions
func Contains(slice []string, target string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, target) {
			return true
		}
	}
	return false
}

func ValidateEditor(editor string) error {
	for _, validEditor := range ValidEditors {
		if strings.EqualFold(editor, validEditor) {
			return nil
		}
	}
	return fmt.Errorf("invalid editor '%s'", editor)
}

func (v *Validator) ValidateConflicts(other *NixConfig) error {
	var conflicts []string

	// Check shell conflicts
	if v.config.Shell.Type != other.Shell.Type {
		conflicts = append(conflicts, fmt.Sprintf("shell type mismatch: personal=%s, project=%s",
			v.config.Shell.Type, other.Shell.Type))
	}

	// Check editor conflicts
	if v.config.Editor.Type != other.Editor.Type {
		conflicts = append(conflicts, fmt.Sprintf("editor type mismatch: personal=%s, project=%s",
			v.config.Editor.Type, other.Editor.Type))
	}

	// Check environment variable conflicts
	for env, value := range other.Team.Settings {
		if personalValue, exists := v.config.Team.Settings[env]; exists {
			if value != personalValue {
				conflicts = append(conflicts, fmt.Sprintf("environment %s has conflicting values", env))
			}
		}
	}

	if len(conflicts) > 0 {
		return fmt.Errorf("- %s", strings.Join(conflicts, "\n- "))
	}

	return nil
}

func ValidateAutoShell(shell string) error {
	if !Contains(ValidAutoShells, shell) {
		return fmt.Errorf("invalid shell '%s': must be one of %v", shell, ValidAutoShells)
	}
	return nil
}

func (v *Validator) validatePackages() error {
	seen := make(map[string]bool)

	for _, pkg := range v.config.Packages.Additional {
		if seen[pkg] {
			return fmt.Errorf("duplicate package: %s", pkg)
		}
		seen[pkg] = true
	}

	for platform, pkgs := range v.config.Packages.PlatformSpecific {
		if !isValidPlatform(platform) {
			return fmt.Errorf("invalid platform: %s", platform)
		}
		for _, pkg := range pkgs {
			if seen[pkg] {
				return fmt.Errorf("duplicate package: %s", pkg)
			}
			seen[pkg] = true
		}
	}

	return nil
}

func (v *Validator) validateGit() error {
	if v.config.Git.Enable {
		if v.config.Git.User.Name == "" {
			return fmt.Errorf("git user name is required when git is enabled")
		}
		if v.config.Git.User.Email == "" {
			return fmt.Errorf("git user email is required when git is enabled")
		}
	}
	return nil
}

func (v *Validator) validateTeam() error {
	if v.config.Team.Enable {
		if v.config.Team.Name == "" {
			return fmt.Errorf("team name is required when team is enabled")
		}
	}
	return nil
}

func isValidPlatform(platform string) bool {
	validPlatforms := map[string]bool{
		"linux":   true,
		"darwin":  true,
		"windows": true,
	}
	return validPlatforms[strings.ToLower(platform)]
}
