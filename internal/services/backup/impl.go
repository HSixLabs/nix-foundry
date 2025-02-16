package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (s *ServiceImpl) Create(name string, force bool) error {
	if name == "" {
		name = time.Now().Format("20060102-150405")
	}

	backupPath := filepath.Join(s.backupDir, fmt.Sprintf("%s.tar.gz", name))

	if !force {
		if _, err := os.Stat(backupPath); err == nil {
			return fmt.Errorf("backup %s already exists", name)
		}
	}

	if err := s.configManager.CreateBackup(backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	return nil
}

func (s *ServiceImpl) Delete(name string) error {
	backupPath := filepath.Join(s.backupDir, fmt.Sprintf("%s.tar.gz", name))

	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("backup %s not found", name)
	}

	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	return nil
}

func (s *ServiceImpl) Rotate(maxAge time.Duration) error {
	backups, err := s.List()
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	now := time.Now()
	for _, backup := range backups {
		// Skip safety backups
		if strings.HasPrefix(backup.ID, "pre-restore-") {
			continue
		}

		// Delete if backup is too old
		if now.Sub(backup.Timestamp) > maxAge {
			if err := s.Delete(backup.ID); err != nil {
				return fmt.Errorf("failed to delete old backup %s: %w", backup.ID, err)
			}
		}
	}

	return nil
}
