package validator

import (
	"time"

	"github.com/shawnkhoffman/nix-foundry/pkg/schema"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v2"
)

var validate = validator.New()

func ValidateConfig(cfg *schema.Config) error {
	return validate.Struct(cfg)
}

func ValidateYAMLContent(content []byte) (*schema.Config, error) {
	var cfg schema.Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, err
	}
	return &cfg, ValidateConfig(&cfg)
}

func RegisterDurationValidation() error {
	return validate.RegisterValidation("duration", func(fl validator.FieldLevel) bool {
		_, err := time.ParseDuration(fl.Field().String())
		return err == nil
	})
}
