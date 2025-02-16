package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
)

func (s *ServiceImpl) RestoreFromTime(targetTime time.Time, force bool) error {
	// Find the backup directory for the target time
	backupDir := filepath.Join(s.configDir, "backups")
	targetBackup := filepath.Join(backupDir, targetTime.Format("20060102-150405"))

	// Check if backup exists
	if _, err := os.Stat(targetBackup); err != nil {
		return errors.NewLoadError(targetBackup, err, "backup not found")
	}

	// If not forcing, check for conflicts
	if !force {
		if err := s.checkRestoreConflicts(targetBackup); err != nil {
			return fmt.Errorf("restore conflicts detected: %w", err)
		}
	}

	// Restore the environment from backup
	if err := s.RestoreEnvironment(targetBackup); err != nil {
		return fmt.Errorf("failed to restore environment: %w", err)
	}

	// Validate the restored environment
	currentEnv, err := s.GetCurrentEnvironment()
	if err != nil {
		return fmt.Errorf("failed to get current environment after restore: %w", err)
	}

	if err := s.ValidateRestoredEnvironment(currentEnv); err != nil {
		return fmt.Errorf("restored environment validation failed: %w", err)
	}

	return nil
}

func (s *ServiceImpl) checkRestoreConflicts(backupPath string) error {
	// Check if backup exists and is readable
	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("backup not accessible: %w", err)
	}

	// Check if backup has required structure
	requiredFiles := []string{"config.yaml", "environment.yaml"}
	for _, file := range requiredFiles {
		path := filepath.Join(backupPath, file)
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("backup missing required file %s: %w", file, err)
		}
	}

	// Check if current environment has uncommitted changes
	currentEnv := filepath.Join(s.configDir, "environments", "current")
	if _, err := os.Stat(currentEnv); err == nil {
		// TODO: Add logic to check for uncommitted changes
		// For now, just check if the environment is in use
		if inUse, err := s.isEnvironmentInUse(currentEnv); err != nil {
			return fmt.Errorf("failed to check environment status: %w", err)
		} else if inUse {
			return fmt.Errorf("current environment is in use, cannot restore")
		}
	}

	return nil
}

// Add this helper method to ServiceImpl
func (s *ServiceImpl) isEnvironmentInUse(envPath string) (bool, error) {
	// Basic implementation - check for lock file
	lockFile := filepath.Join(envPath, ".lock")
	if _, err := os.Stat(lockFile); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, fmt.Errorf("failed to check environment lock: %w", err)
	}
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

// Rename this to be a helper method since Restore is in impl.go
func (s *ServiceImpl) restoreFromArchive(archivePath, destPath string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open backup archive: %w", err)
	}
	defer file.Close()

	gr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	// Clear destination directory
	if err := os.RemoveAll(destPath); err != nil {
		return fmt.Errorf("failed to clear destination: %w", err)
	}

	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		target := filepath.Join(destPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return fmt.Errorf("failed to write file contents: %w", err)
			}
			f.Close()
		case tar.TypeSymlink:
			if err := os.Symlink(header.Linkname, target); err != nil {
				return fmt.Errorf("failed to create symlink: %w", err)
			}
		}
	}

	return nil
}

func (s *ServiceImpl) RestoreEnvironment(backupPath string) error {
	// Get the current environment path
	currentEnv := filepath.Join(s.configDir, "environments", "current")

	// Use the existing restoreFromArchive method
	return s.restoreFromArchive(backupPath, currentEnv)
}

func (s *ServiceImpl) GetCurrentEnvironment() (string, error) {
	currentEnv := filepath.Join(s.configDir, "environments", "current")
	if _, err := os.Stat(currentEnv); err != nil {
		return "", fmt.Errorf("current environment not found: %w", err)
	}
	return currentEnv, nil
}

func (s *ServiceImpl) ValidateRestoredEnvironment(envPath string) error {
	// Basic validation - check if the directory exists and is readable
	if _, err := os.Stat(envPath); err != nil {
		return fmt.Errorf("restored environment not found: %w", err)
	}

	// Add more validation as needed (e.g., check for required files)
	requiredFiles := []string{"config.yaml", "environment.yaml"}
	for _, file := range requiredFiles {
		path := filepath.Join(envPath, file)
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("missing required file %s: %w", file, err)
		}
	}

	return nil
}
