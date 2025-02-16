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

func (s *ServiceImpl) Restore(name string) error {
	backupPath := filepath.Join(s.backupDir, fmt.Sprintf("%s.tar.gz", name))

	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("backup %s not found", name)
	}

	if err := s.configManager.RestoreBackup(backupPath); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}

func (s *ServiceImpl) List() ([]BackupInfo, error) {
	entries, err := os.ReadDir(s.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".gz" {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			name := entry.Name()
			name = name[:len(name)-7] // Remove .tar.gz extension

			backups = append(backups, BackupInfo{
				Name:      name,
				CreatedAt: info.ModTime(),
				Size:      info.Size(),
				Path:      filepath.Join(s.backupDir, entry.Name()),
			})
		}
	}

	return backups, nil
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
		if strings.HasPrefix(backup.Name, "pre-restore-") {
			continue
		}

		// Delete if backup is too old
		if now.Sub(backup.CreatedAt) > maxAge {
			if err := s.Delete(backup.Name); err != nil {
				return fmt.Errorf("failed to delete old backup %s: %w", backup.Name, err)
			}
		}
	}

	return nil
}
