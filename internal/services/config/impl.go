package config

import (
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/types"
)

func (s *ServiceImpl) SetValue(key string, value interface{}) error {
	if _, err := s.Load(); err != nil {
		return err
	}

	return s.manager.SetValue(key, value)
}

func (s *ServiceImpl) Reset(section string) error {
	if section == "" {
		s.config = NewDefaultConfig()
		return s.Save(s.config)
	}

	defaultConfig := NewDefaultConfig()
	defaultValue := reflect.ValueOf(defaultConfig).FieldByName(section)
	if !defaultValue.IsValid() {
		return fmt.Errorf("invalid configuration section: %s", section)
	}

	configValue := reflect.ValueOf(s.config).Elem().FieldByName(section)
	if !configValue.IsValid() {
		return fmt.Errorf("invalid configuration section: %s", section)
	}

	configValue.Set(defaultValue)
	return s.Save(s.config)
}

func (s *ServiceImpl) ValidateConfig(cfg *types.Config) error {
	if cfg.Project.Name == "" {
		return fmt.Errorf("project name required")
	}
	return nil
}

func (s *ServiceImpl) Validate() error {
	if s.config == nil {
		if _, err := s.Load(); err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
	}
	return s.ValidateConfig(s.config)
}

func (s *ServiceImpl) Load() (*types.Config, error) {
	cfg, err := s.manager.Load()
	if err != nil {
		return nil, err
	}
	s.config = cfg
	return cfg, nil
}

func (s *ServiceImpl) Save(cfg *types.Config) error {
	return s.manager.Save(cfg)
}

func (s *ServiceImpl) SaveConfig(cfg *types.Config) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}
	s.config = cfg
	return s.manager.Save(cfg)
}

func (s *ServiceImpl) GetConfigDir() string {
	return filepath.Dir(s.path)
}

func (s *ServiceImpl) GetBackupDir() string {
	return filepath.Join(s.GetConfigDir(), "backups")
}

func (s *ServiceImpl) LoadSection(name string, v interface{}) error {
	if _, err := s.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	configValue := reflect.ValueOf(s.config).Elem()
	sectionValue := configValue.FieldByName(name)
	if !sectionValue.IsValid() {
		return fmt.Errorf("invalid configuration section: %s", name)
	}

	// Create a new value to hold the section data
	targetValue := reflect.ValueOf(v).Elem()
	targetValue.Set(sectionValue)

	return nil
}

func (s *ServiceImpl) GetManager() *Manager {
	return s.manager
}
