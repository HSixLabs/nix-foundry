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

	if v.config.Shell.Type == "" {
		return fmt.Errorf("shell type is required")
	} else if !Contains(ValidShells, v.config.Shell.Type) {
		return fmt.Errorf("invalid shell type: %s", v.config.Shell.Type)
	}

	if v.config.Editor.Type == "" {
		return fmt.Errorf("editor type is required")
	} else if !Contains(ValidEditors, v.config.Editor.Type) {
		return fmt.Errorf("invalid editor type: %s", v.config.Editor.Type)
	}

	return nil
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

func Contains(slice []string, item string) bool {
	// Check for direct match first
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}

	// Check aliases if no direct match found
	if alias, exists := editorAliases[strings.ToLower(item)]; exists {
		for _, s := range slice {
			if strings.EqualFold(s, alias) {
				return true
			}
		}
	}

	return false
}

func ValidateAutoShell(shell string) error {
	if !Contains(ValidAutoShells, shell) {
		return fmt.Errorf("invalid shell '%s': must be one of %v", shell, ValidAutoShells)
	}
	return nil
}

func ValidateEditor(editor string) error {
	// Direct validation against ValidEditors only, no aliases
	for _, validEditor := range ValidEditors {
		if strings.EqualFold(editor, validEditor) {
			return nil
		}
	}
	return fmt.Errorf("invalid editor '%s': must be one of %v", editor, ValidEditors)
}
