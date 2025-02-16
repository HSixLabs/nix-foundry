package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func (s *ServiceImpl) GetValue(key string) (interface{}, error) {
	if err := s.Load(); err != nil {
		return nil, err
	}

	value, err := getNestedValue(s.config, strings.Split(key, "."))
	if err != nil {
		return nil, fmt.Errorf("failed to get value: %w", err)
	}

	return value, nil
}

func (s *ServiceImpl) SetValue(key string, value interface{}) error {
	if err := s.Load(); err != nil {
		return err
	}

	return s.manager.SetValue(key, value)
}

func (s *ServiceImpl) Reset(section string) error {
	if section == "" {
		s.config = NewDefaultConfig()
		return s.Save()
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
	return s.Save()
}

func (s *ServiceImpl) Validate() error {
	if s.config == nil {
		return fmt.Errorf("configuration is nil")
	}

	// Validate backup settings
	if err := s.config.Backup.Validate(); err != nil {
		return fmt.Errorf("backup configuration invalid: %w", err)
	}

	// Validate environment settings
	if err := s.config.Environment.Validate(); err != nil {
		return fmt.Errorf("environment configuration invalid: %w", err)
	}

	return nil
}

func (s *ServiceImpl) Load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			s.config = NewDefaultConfig()
			return nil
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &s.config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

func (s *ServiceImpl) Save() error {
	s.config.LastUpdated = time.Now()

	data, err := yaml.Marshal(s.config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (s *ServiceImpl) GetConfigDir() string {
	return filepath.Dir(s.path)
}

func (s *ServiceImpl) GetBackupDir() string {
	return filepath.Join(s.GetConfigDir(), "backups")
}

func (s *ServiceImpl) LoadSection(name string, v interface{}) error {
	if err := s.Load(); err != nil {
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
