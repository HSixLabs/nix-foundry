package uninstall

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	"github.com/shawnkhoffman/nix-foundry/internal/services/config"
)

type Service interface {
	Execute(keepBackups bool) error
}

type ServiceImpl struct {
	logger    *logging.Logger
	configSvc config.Service
}

func NewService() Service {
	return &ServiceImpl{
		logger:    logging.GetLogger(),
		configSvc: config.NewService(),
	}
}

func (s *ServiceImpl) Execute(keepBackups bool) error {
	configDir := s.configSvc.GetConfigDir()

	// Run cleanup steps first
	if err := s.removeGlobalSymlinks(); err != nil {
		s.logger.Warn("Failed to remove symlinks", "error", err)
	}

	if err := s.cleanLegacyFiles(); err != nil {
		s.logger.Warn("Failed to clean legacy files", "error", err)
	}

	// Remove configuration directory
	if err := os.RemoveAll(configDir); err != nil {
		return fmt.Errorf("failed to remove config directory: %w", err)
	}

	// Conditionally remove backups
	if !keepBackups {
		backupDir := filepath.Join(configDir, "backups")
		if err := os.RemoveAll(backupDir); err != nil {
			return fmt.Errorf("failed to remove backups: %w", err)
		}
	}

	// Remove cache directory
	cacheDir := filepath.Join(os.Getenv("HOME"), ".cache", "nix-foundry")
	if err := os.RemoveAll(cacheDir); err != nil {
		return fmt.Errorf("failed to remove cache: %w", err)
	}

	return nil
}

func (s *ServiceImpl) removeGlobalSymlinks() error {
	return filepath.Walk(s.configSvc.GetConfigDir(), func(path string, info os.FileInfo, err error) error {
		if info.Mode()&os.ModeSymlink != 0 {
			return os.Remove(path)
		}
		return nil
	})
}

func (s *ServiceImpl) cleanLegacyFiles() error {
	legacyPaths := []string{
		filepath.Join(s.configSvc.GetConfigDir(), "old_config.yaml"),
		filepath.Join(s.configSvc.GetConfigDir(), "deprecated"),
	}

	for _, path := range legacyPaths {
		if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
