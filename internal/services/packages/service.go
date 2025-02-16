package packages

import (
	"os"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	"gopkg.in/yaml.v3"
)

type Service interface {
	Add(pkgs []string, pkgType string) error
	Remove(pkgs []string, pkgType string) error
	List() (map[string][]string, error)
	Sync() error
}

type ServiceImpl struct {
	logger     *logging.Logger
	configPath string
}

func NewService(configDir string) Service {
	return &ServiceImpl{
		logger:     logging.GetLogger(),
		configPath: filepath.Join(configDir, "packages.yaml"),
	}
}

func (s *ServiceImpl) loadPackages() (map[string][]string, error) {
	data := make(map[string][]string)

	if _, err := os.Stat(s.configPath); err == nil {
		file, err := os.ReadFile(s.configPath)
		if err != nil {
			return nil, errors.NewLoadError(s.configPath, err, "failed to read packages config")
		}

		if err := yaml.Unmarshal(file, &data); err != nil {
			return nil, errors.NewLoadError(s.configPath, err, "failed to parse packages config")
		}
	}

	return data, nil
}

func (s *ServiceImpl) savePackages(pkgs map[string][]string) error {
	data, err := yaml.Marshal(pkgs)
	if err != nil {
		return errors.NewLoadError(s.configPath, err, "failed to serialize packages")
	}

	if err := os.WriteFile(s.configPath, data, 0644); err != nil {
		return errors.NewLoadError(s.configPath, err, "failed to save packages configuration")
	}

	return nil
}

func (s *ServiceImpl) Add(pkgs []string, pkgType string) error {
	existing, err := s.loadPackages()
	if err != nil {
		return err
	}

	existing[pkgType] = unique(append(existing[pkgType], pkgs...))
	return s.savePackages(existing)
}

func (s *ServiceImpl) Remove(pkgs []string, pkgType string) error {
	existing, err := s.loadPackages()
	if err != nil {
		return err
	}

	existing[pkgType] = filter(existing[pkgType], pkgs)
	return s.savePackages(existing)
}

func (s *ServiceImpl) List() (map[string][]string, error) {
	return s.loadPackages()
}

func (s *ServiceImpl) Sync() error {
	// Implementation that syncs packages with Nix configuration
	// Would call config service to regenerate files
	return nil
}

// Helper functions
func unique(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, item := range slice {
		if _, value := keys[item]; !value {
			keys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func filter(source, remove []string) []string {
	m := make(map[string]bool)
	for _, item := range remove {
		m[item] = true
	}

	var result []string
	for _, item := range source {
		if !m[item] {
			result = append(result, item)
		}
	}
	return result
}
