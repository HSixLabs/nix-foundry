package environment

import (
	"os"
	"path/filepath"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
)

func (s *ServiceImpl) Rollback(targetTime time.Time, force bool) error {
	backupPath := filepath.Join(s.configDir, "backups", targetTime.Format("20060102-150405"))

	// Use struct fields
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		s.logger.Error("Backup not found", "path", backupPath)
		return errors.NewValidationError(backupPath, err, "no backup found")
	}

	return s.RestoreEnvironment(backupPath)
}

func (s *ServiceImpl) ListBackups() ([]time.Time, error) {
	backupDir := filepath.Join(s.configDir, "backups")
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, errors.NewLoadError(backupDir, err, "failed to read backups")
	}

	var backups []time.Time
	for _, entry := range entries {
		if entry.IsDir() {
			if t, err := time.Parse("20060102-150405", entry.Name()); err == nil {
				backups = append(backups, t)
			}
		}
	}
	return backups, nil
}
