package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"archive/tar"
	"compress/gzip"

	"github.com/shawnkhoffman/nix-foundry/internal/pkg/errors"
	"github.com/shawnkhoffman/nix-foundry/internal/pkg/logging"
	config "github.com/shawnkhoffman/nix-foundry/internal/services/config"
	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/shawnkhoffman/nix-foundry/internal/services/project"
	"github.com/shawnkhoffman/nix-foundry/pkg/metrics"
)

// Service defines the interface for backup operations
type Service interface {
	// Core backup operations
	Create(name string, force bool) error
	Restore(name string, force bool) error
	List() ([]BackupEntry, error)
	Delete(name string) error
	Rotate(maxAge time.Duration) error

	// Compression and encryption
	Compress(name string) error
	EncryptBackup(name string, keyPath string) error
	DecryptBackup(name string, keyPath string) error

	// Configuration
	GetConfig() Config
	UpdateConfig(config Config) error

	// Additional operations
	CreateSafetyBackup() (string, error)
	ValidateRestoredEnvironment(envPath string) error
}

// Entry represents the structure of a backup entry
type Entry struct {
	ID        string
	Timestamp time.Time
	Size      int64
}

// Config represents backup service configuration
type Config struct {
	RetentionDays    int      `yaml:"retentionDays"`
	MaxBackups       int      `yaml:"maxBackups"`
	CompressionLevel int      `yaml:"compressionLevel"`
	ExcludePatterns  []string `yaml:"excludePatterns"`
}

// BackupEntry struct to replace both Info and Entry
type BackupEntry struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Size      int64     `json:"size"`
	Path      string    `json:"path"`
}

// ServiceImpl implements the backup Service interface
type ServiceImpl struct {
	configManager config.Service
	envService    environment.Service
	projectSvc    project.Service
	logger        *logging.Logger
	backupDir     string
	configDir     string
}

// NewService creates a new backup service
func NewService(
	cfgManager config.Service,
	envSvc environment.Service,
	projectSvc project.Service,
) *ServiceImpl {
	return &ServiceImpl{
		configManager: cfgManager,
		envService:    envSvc,
		projectSvc:    projectSvc,
		logger:        logging.GetLogger(),
		backupDir:     cfgManager.GetBackupDir(),
		configDir:     cfgManager.GetConfigDir(),
	}
}

// NewServiceWithDeps creates a new backup service with dependencies
func NewServiceWithDeps(cfg config.Service, projectSvc project.Service) *ServiceImpl {
	return &ServiceImpl{
		configManager: cfg,
		projectSvc:    projectSvc,
		logger:        logging.GetLogger(),
		backupDir:     cfg.GetBackupDir(),
		configDir:     cfg.GetConfigDir(),
	}
}

// RestoreBackup restores a backup with proper error handling
func (s *ServiceImpl) RestoreBackup(backupID string) error {
	backupEnv := filepath.Join(s.backupDir, backupID)
	currentEnv := filepath.Join(s.configManager.GetConfigDir(), "environments", "current")

	// Validate backup before restoring
	if err := s.validateBackup(backupEnv); err != nil {
		return fmt.Errorf("invalid backup: %w", err)
	}

	var restoreErr error
	// Track operation
	defer func() {
		s.trackBackupOperation("restore", restoreErr == nil, 0, restoreErr)
	}()

	// Create safety backup
	if err := os.Rename(currentEnv, currentEnv+".old"); err != nil {
		restoreErr = fmt.Errorf("failed to create safety backup: %w", err)
		return restoreErr
	}

	// Attempt restore
	if err := os.Rename(backupEnv, currentEnv); err != nil {
		// Attempt rollback if restore fails
		if rollbackErr := os.Rename(currentEnv+".old", currentEnv); rollbackErr != nil {
			restoreErr = fmt.Errorf("restore failed: %v, rollback failed: %v", err, rollbackErr)
			return restoreErr
		}
		restoreErr = fmt.Errorf("failed to restore backup: %w", err)
		return restoreErr
	}

	if err := syscall.Sync(); err != nil {
		s.logger.Error("Filesystem sync failed", "error", err)
	}
	return os.RemoveAll(currentEnv + ".old")
}

// RotateAndRestore handles backup rotation and restoration
func (s *ServiceImpl) RotateAndRestore(backupID string, force bool) error {
	currentEnv := filepath.Join(s.configManager.GetConfigDir(), "environments", "current")

	// Create safety backup
	if err := os.Rename(currentEnv, currentEnv+".old"); err != nil {
		return fmt.Errorf("failed to create safety backup: %w", err)
	}

	// Track operation
	defer s.trackBackupOperation("rotate_restore", true, 0, nil)

	// Use RestoreBackup instead of Restore
	if err := s.RestoreBackup(backupID); err != nil {
		// Attempt rollback if restore fails
		if rollbackErr := os.Rename(currentEnv+".old", currentEnv); rollbackErr != nil {
			return fmt.Errorf("restore failed: %v, rollback failed: %v", err, rollbackErr)
		}
		return err
	}

	return nil
}

// validateBackup checks if a backup is valid
func (s *ServiceImpl) validateBackup(backupPath string) error {
	requiredFiles := []string{"manifest.json", "environment.nix"}
	for _, file := range requiredFiles {
		if _, err := os.Stat(filepath.Join(backupPath, file)); err != nil {
			return fmt.Errorf("missing required file %s: %w", file, err)
		}
	}
	return nil
}

// trackBackupOperation records metrics for backup operations
func (s *ServiceImpl) trackBackupOperation(opType string, success bool, size int64, err error) {
	metrics.Record(map[string]interface{}{
		"type":       opType,
		"success":    success,
		"size":       size,
		"backup_dir": s.backupDir,
		"timestamp":  time.Now().UTC(),
		"error":      err,
	})
}

func (s *ServiceImpl) CreateProjectBackup(projectID string) error {
	return s.projectSvc.Backup(projectID)
}

// Add this method to fulfill the Service interface
func (s *ServiceImpl) EncryptBackup(name string, keyPath string) error {
	return fmt.Errorf("encryption not implemented yet")
}

// Add this method to fulfill the Service interface
func (s *ServiceImpl) DecryptBackup(name string, keyPath string) error {
	// Implementation logic here
	return fmt.Errorf("not implemented yet")
}

// Add missing DeleteBackup method
func (s *ServiceImpl) DeleteBackup(id string) error {
	path := filepath.Join(s.backupDir, id)
	return os.RemoveAll(path)
}

// Compress compresses a backup using gzip
func (s *ServiceImpl) Compress(name string) error {
	backupPath := filepath.Join(s.backupDir, name)
	compressedPath := backupPath + ".tar.gz"

	if err := s.createArchive(backupPath, compressedPath); err != nil {
		return errors.NewBackupError(backupPath, err, "compression failed")
	}

	// Remove uncompressed backup
	if err := os.RemoveAll(backupPath); err != nil {
		s.logger.Warn("Failed to remove uncompressed backup", "path", backupPath, "error", err)
	}

	return nil
}

// Add helper method for archive creation
func (s *ServiceImpl) createArchive(sourcePath, destPath string) error {
	// Create the destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}
	defer destFile.Close()

	// Create gzip writer with configured compression level
	gw, err := gzip.NewWriterLevel(destFile, s.configManager.GetCompressionLevel())
	if err != nil {
		return fmt.Errorf("failed to create gzip writer: %w", err)
	}
	defer gw.Close()

	// Create tar writer
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Walk through the source directory
	err = filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path for tar header
		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("failed to create tar header: %w", err)
		}
		header.Name = relPath

		// Write header
		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header: %w", err)
		}

		// If it's a regular file, copy its contents
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer file.Close()

			if _, err := io.Copy(tw, file); err != nil {
				return fmt.Errorf("failed to write file contents: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}

	return nil
}

// Add these methods to ServiceImpl
func (s *ServiceImpl) GetConfig() Config {
	return Config{
		RetentionDays:    s.configManager.GetRetentionDays(),
		MaxBackups:       s.configManager.GetMaxBackups(),
		CompressionLevel: s.configManager.GetCompressionLevel(),
	}
}

func (s *ServiceImpl) UpdateConfig(config Config) error {
	// Implementation that updates the configuration
	s.configManager.SetRetentionDays(config.RetentionDays)
	s.configManager.SetMaxBackups(config.MaxBackups)
	s.configManager.SetCompressionLevel(config.CompressionLevel)
	return s.configManager.Save()
}

// Rename the key-based update method
func (s *ServiceImpl) UpdateConfigKey(key string, value interface{}) error {
	switch key {
	case "retention":
		if v, ok := value.(int); ok {
			s.configManager.SetRetentionDays(v)
		}
	case "max-backups":
		if v, ok := value.(int); ok {
			s.configManager.SetMaxBackups(v)
		}
	case "compression":
		if v, ok := value.(int); ok {
			s.configManager.SetCompressionLevel(v)
		}
	default:
		return fmt.Errorf("invalid config key: %s", key)
	}
	return s.configManager.Save()
}

// Update ServiceImpl to implement new restore methods
func (s *ServiceImpl) CreateSafetyBackup() (string, error) {
	backupID := fmt.Sprintf("safety-%s", time.Now().Format("20060102-150405"))
	if err := s.Create(backupID, true); err != nil {
		return "", fmt.Errorf("failed to create safety backup: %w", err)
	}
	return backupID, nil
}

// Update List method to return BackupEntry
func (s *ServiceImpl) List() ([]BackupEntry, error) {
	entries, err := os.ReadDir(s.backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []BackupEntry
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Parse timestamp from backup name
		timestamp, err := time.Parse("20060102-150405", strings.TrimSuffix(entry.Name(), ".tar.gz"))
		if err != nil {
			continue
		}

		backups = append(backups, BackupEntry{
			ID:        strings.TrimSuffix(entry.Name(), ".tar.gz"),
			Timestamp: timestamp,
			Size:      info.Size(),
			Path:      filepath.Join(s.backupDir, entry.Name()),
		})
	}

	return backups, nil
}

// Complete the Restore implementation
func (s *ServiceImpl) Restore(name string, force bool) error {
	backupPath := filepath.Join(s.backupDir, fmt.Sprintf("%s.tar.gz", name))

	if _, err := os.Stat(backupPath); err != nil {
		return fmt.Errorf("backup %s not found", name)
	}

	// Track operation
	var restoreErr error
	defer func() {
		s.trackBackupOperation("restore", restoreErr == nil, 0, restoreErr)
	}()

	currentEnv := filepath.Join(s.configManager.GetConfigDir(), "environments", "current")

	// Handle force flag
	if !force {
		if err := os.Rename(currentEnv, currentEnv+".old"); err != nil {
			restoreErr = fmt.Errorf("failed to create safety backup: %w", err)
			return restoreErr
		}
	} else {
		if err := os.RemoveAll(currentEnv); err != nil {
			restoreErr = fmt.Errorf("failed to remove current environment: %w", err)
			return restoreErr
		}
	}

	if err := s.configManager.RestoreBackup(backupPath); err != nil {
		// Attempt rollback if not forcing
		if !force {
			if rollbackErr := os.Rename(currentEnv+".old", currentEnv); rollbackErr != nil {
				restoreErr = fmt.Errorf("restore failed: %v, rollback failed: %v", err, rollbackErr)
				return restoreErr
			}
		}
		restoreErr = fmt.Errorf("failed to restore backup: %w", err)
		return restoreErr
	}

	// Clean up old backup if it exists and restore succeeded
	if !force {
		return os.RemoveAll(currentEnv + ".old")
	}
	return nil
}
