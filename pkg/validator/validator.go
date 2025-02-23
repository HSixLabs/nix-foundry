/*
Package validator provides configuration validation functionality for Nix Foundry.
It handles validation of configuration structures, applies default values,
and provides custom validation rules for specific types.
*/
package validator

import (
	"fmt"
	"time"

	"github.com/shawnkhoffman/nix-foundry/pkg/schema"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

var validate = validator.New()

func init() {
	if err := RegisterDurationValidation(); err != nil {
		panic(fmt.Sprintf("failed to register duration validation: %v", err))
	}
}

/*
ValidateConfig validates a Config struct and applies default values where needed.
It ensures all required fields are present and valid according to their validation rules.
Returns an error if validation fails or if the config is nil.
*/
func ValidateConfig(cfg *schema.Config) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	applyDefaults(cfg)

	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	return nil
}

/*
ValidateYAMLContent validates YAML content and returns a Config struct.
If the content is empty, returns a new default configuration.
Otherwise, unmarshals the YAML content, applies defaults, and validates the resulting config.
*/
func ValidateYAMLContent(content []byte) (*schema.Config, error) {
	if len(content) == 0 {
		cfg := schema.NewDefaultConfig()
		return cfg, nil
	}

	var cfg schema.Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	applyDefaults(&cfg)
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

/*
applyDefaults applies default values to a Config struct.
Sets sensible defaults for various configuration fields when they are empty
or uninitialized.
*/
func applyDefaults(cfg *schema.Config) {
	if cfg.Version == "" {
		cfg.Version = "1.0"
	}
	if cfg.Kind == "" {
		cfg.Kind = "user"
	}
	if cfg.Metadata.Name == "" {
		cfg.Metadata.Name = "default"
	}
	if cfg.Settings.LogLevel == "" {
		cfg.Settings.LogLevel = "info"
	}
	if cfg.Settings.UpdateInterval == 0 {
		cfg.Settings.UpdateInterval = 24 * time.Hour
	}
	if cfg.Nix.Packages.Core == nil {
		cfg.Nix.Packages.Core = make([]string, 0)
	}
	if cfg.Nix.Packages.Optional == nil {
		cfg.Nix.Packages.Optional = make([]string, 0)
	}
	if cfg.Nix.Scripts == nil {
		cfg.Nix.Scripts = make([]schema.Script, 0)
	}
}

/*
RegisterDurationValidation registers a custom validation for time.Duration fields.
This validator ensures that string representations of durations can be properly parsed.
Returns an error if registration fails.
*/
func RegisterDurationValidation() error {
	return validate.RegisterValidation("duration", func(fl validator.FieldLevel) bool {
		if str, ok := fl.Field().Interface().(string); ok {
			_, err := time.ParseDuration(str)
			return err == nil
		}
		return true
	})
}
