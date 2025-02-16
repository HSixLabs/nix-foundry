package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
)

type Service struct {
	configDir string
}

func NewService(configDir string) *Service {
	return &Service{configDir: configDir}
}

// Create with force flag support
func (s *Service) Create(name string, packages []string, force bool) error {
	path := filepath.Join(s.configDir, name+".yaml")

	if !force {
		if _, err := os.Stat(path); err == nil {
			return errors.NewConflictError(nil, "profile already exists, use --force to overwrite")
		}
	}

	// Actual creation logic
	return os.WriteFile(path, []byte(generateProfile(packages)), 0644)
}

func generateProfile(packages []string) string {
	return fmt.Sprintf("packages:\n  - %s", strings.Join(packages, "\n  - "))
}

// Added list functionality
func (s *Service) List() ([]string, error) {
	files, err := os.ReadDir(s.configDir)
	if err != nil {
		return nil, errors.NewLoadError(s.configDir, err, "failed to read profiles directory")
	}

	var profiles []string
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".yaml" {
			profiles = append(profiles, strings.TrimSuffix(f.Name(), ".yaml"))
		}
	}
	return profiles, nil
}

// Added delete functionality
func (s *Service) Delete(name string) error {
	return os.Remove(filepath.Join(s.configDir, name+".yaml"))
}
