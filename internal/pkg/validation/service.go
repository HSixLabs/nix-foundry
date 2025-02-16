package validation

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
)

type Service interface {
	ValidateEnvironment(path string) error
	ValidateConfig(config interface{}) error
	CheckDependencies(deps []string) error
	ValidateBackup(backupPath string) error
}

type Validator struct{}

func NewService() Service {
	return &Validator{}
}

// ValidateEnvironment checks for required environment structure
func (v *Validator) ValidateEnvironment(path string) error {
	requiredFiles := map[string]bool{
		"flake.nix":         true,
		"home.nix":          true,
		"shell.nix":         false,
		"configuration.nix": false,
	}

	var missing []string
	for file, required := range requiredFiles {
		fullPath := filepath.Join(path, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) && required {
			missing = append(missing, file)
		}
	}

	if len(missing) > 0 {
		return errors.E(
			errors.Operation("environment_validation"),
			errors.ErrConfigValidation,
			fmt.Errorf("missing required files: %s", strings.Join(missing, ", ")),
		).WithPath(path)
	}

	// Check directory permissions
	if info, err := os.Stat(path); err == nil {
		if !info.IsDir() {
			return errors.E(
				errors.Operation("environment_validation"),
				errors.ErrConfigValidation,
				fmt.Errorf("path is not a directory"),
			).WithPath(path)
		}

		// Check if directory is writable
		if info.Mode().Perm()&0200 == 0 {
			return errors.E(
				errors.Operation("environment_validation"),
				errors.ErrEnvPermission,
				fmt.Errorf("directory is not writable"),
			).WithPath(path)
		}
	}

	return nil
}

// ValidateConfig checks configuration structural validity
func (v *Validator) ValidateConfig(config interface{}) error {
	if config == nil {
		return errors.E(
			errors.Operation("config_validation"),
			errors.ErrConfigValidation,
			fmt.Errorf("nil configuration provided"),
		)
	}

	// Use reflection to handle different config types
	val := reflect.ValueOf(config)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Handle specific config types
	switch cfg := config.(type) {
	case *ProjectConfig:
		if err := cfg.Validate(); err != nil {
			return errors.E(
				errors.Operation("config_validation"),
				errors.ErrConfigValidation,
				err,
			).WithContext("config_type", "project")
		}
	// Add other config types here
	default:
		return errors.E(
			errors.Operation("config_validation"),
			errors.ErrConfigValidation,
			fmt.Errorf("unsupported config type: %T", config),
		)
	}

	return nil
}

// CheckDependencies verifies system dependencies are available
func (v *Validator) CheckDependencies(deps []string) error {
	var missing []string

	for _, dep := range deps {
		if _, err := exec.LookPath(dep); err != nil {
			missing = append(missing, dep)
		}
	}

	if len(missing) > 0 {
		return errors.E(
			errors.Operation("dependency_check"),
			errors.ErrSystemRequirement,
			fmt.Errorf("missing dependencies: %s", strings.Join(missing, ", ")),
		).WithContext("dependencies", missing)
	}

	return nil
}

// ProjectConfig validation (moved from project service)
type ProjectConfig struct {
	Version      string
	Name         string
	Environment  string
	Dependencies []string
}

func (p *ProjectConfig) Validate() error {
	if p.Version == "" {
		return fmt.Errorf("version is required")
	}
	if p.Name == "" {
		return fmt.Errorf("name is required")
	}
	if p.Environment == "" {
		return fmt.Errorf("environment is required")
	}
	return nil
}

func (v *Validator) ValidateBackup(backupPath string) error {
	if _, err := os.Stat(filepath.Join(backupPath, "manifest.json")); err != nil {
		return fmt.Errorf("invalid backup format: %w", err)
	}
	return nil
}
