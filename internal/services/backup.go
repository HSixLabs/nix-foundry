package services

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/shawnkhoffman/nix-foundry/pkg/config"
)

// BackupService handles backup and restore operations
type BackupService struct {
	configManager *config.Manager
	backupConfig  config.BackupConfig
}

// NewBackupService creates a new backup service
func NewBackupService() (*BackupService, error) {
	configManager := config.NewManager()

	// Load and validate backup configuration
	var backupConfig config.BackupConfig
	if err := configManager.LoadSection("backup", &backupConfig); err != nil {
		// Use defaults if config doesn't exist
		backupConfig = config.DefaultBackupConfig()
		// Log warning but continue with defaults
		fmt.Fprintf(os.Stderr, "Warning: using default backup settings: %v\n", err)
	} else {
		// Validate existing configuration
		if err := backupConfig.Validate(); err != nil {
			return nil, fmt.Errorf("invalid backup configuration: %w", err)
		}
	}

	return &BackupService{
		configManager: configManager,
		backupConfig:  backupConfig,
	}, nil
}

// CreateBackup creates a backup of the current configuration
func (s *BackupService) CreateBackup(name string) error {
	configDir := s.configManager.GetConfigDir()
	backupDir := filepath.Join(configDir, "backups")
	backupPath := filepath.Join(backupDir, name+".tar.gz")

	// Create backups directory if it doesn't exist
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Create the tar.gz file
	file, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer file.Close()

	// Use configured compression level
	gw, err := gzip.NewWriterLevel(file, s.backupConfig.CompressionLevel)
	if err != nil {
		return fmt.Errorf("failed to create gzip writer: %w", err)
	}
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Walk through the config directory and add files to the archive
	err = filepath.Walk(configDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the backups directory itself
		if path == backupDir {
			return filepath.SkipDir
		}

		// Get the relative path
		relPath, err := filepath.Rel(configDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Skip if it's in the backups directory
		cleanPath := filepath.Clean(relPath)
		pathParts := strings.Split(cleanPath, string(filepath.Separator))
		if len(pathParts) > 0 && pathParts[0] == "backups" {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return fmt.Errorf("failed to create tar header: %w", err)
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header: %w", err)
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", path, err)
			}
			defer file.Close()

			if _, err := io.Copy(tw, file); err != nil {
				return fmt.Errorf("failed to write file to tar: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// After successful backup creation, rotate old backups
	if err := s.RotateBackups(); err != nil {
		// Log the error but don't fail the backup creation
		fmt.Fprintf(os.Stderr, "Warning: failed to rotate old backups: %v\n", err)
	}

	return nil
}

// RestoreBackup restores a configuration from a backup
func (s *BackupService) RestoreBackup(name string) error {
	configDir := s.configManager.GetConfigDir()
	backupPath := filepath.Join(configDir, "backups", name+".tar.gz")

	// Open the tar.gz file
	file, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("backup not found: %w", err)
	}
	defer file.Close()

	gr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	// Clear existing configuration (except backups)
	if err := s.clearConfiguration(); err != nil {
		return fmt.Errorf("failed to clear existing configuration: %w", err)
	}

	// Extract the backup
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		target := filepath.Join(configDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			dir := filepath.Dir(target)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}

			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			defer f.Close()

			if _, err := io.Copy(f, tr); err != nil {
				return fmt.Errorf("failed to write file: %w", err)
			}
		}
	}

	return nil
}

// clearConfiguration removes all configuration files except backups
func (s *BackupService) clearConfiguration() error {
	configDir := s.configManager.GetConfigDir()
	entries, err := os.ReadDir(configDir)
	if err != nil {
		return fmt.Errorf("failed to read config directory: %w", err)
	}

	for _, entry := range entries {
		if entry.Name() != "backups" {
			path := filepath.Join(configDir, entry.Name())
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf("failed to remove %s: %w", path, err)
			}
		}
	}

	return nil
}

// ListBackups returns a list of available backup names
func (s *BackupService) ListBackups() ([]string, error) {
	configDir := s.configManager.GetConfigDir()
	backupDir := filepath.Join(configDir, "backups")

	// Create backups directory if it doesn't exist
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Get list of backup files
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	// Filter and format backup names
	var backups []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".gz" {
			// Remove the .tar.gz extension
			name := entry.Name()
			name = name[:len(name)-7] // Remove .tar.gz
			backups = append(backups, name)
		}
	}

	return backups, nil
}

// DeleteBackup removes a backup file
func (s *BackupService) DeleteBackup(name string) error {
	configDir := s.configManager.GetConfigDir()
	backupPath := filepath.Join(configDir, "backups", name+".tar.gz")

	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup '%s' not found", name)
	}

	// Remove the backup file
	if err := os.Remove(backupPath); err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	return nil
}

// RotateBackups removes old backups
func (s *BackupService) RotateBackups() error {
	// Use configured values instead of constants
	maxBackups := s.backupConfig.MaxBackups
	maxBackupAge := time.Duration(s.backupConfig.MaxAgeInDays) * 24 * time.Hour

	configDir := s.configManager.GetConfigDir()
	backupDir := filepath.Join(configDir, "backups")

	// List all backups
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	type backupInfo struct {
		path    string
		modTime time.Time
	}

	var backups []backupInfo
	now := time.Now()

	// Collect backup information
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".gz" {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			backups = append(backups, backupInfo{
				path:    filepath.Join(backupDir, entry.Name()),
				modTime: info.ModTime(),
			})
		}
	}

	// Sort backups by modification time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].modTime.After(backups[j].modTime)
	})

	// Remove old backups
	for i, backup := range backups {
		shouldDelete := false

		// Delete if we have too many backups
		if i >= maxBackups {
			shouldDelete = true
		}

		// Delete if backup is too old
		if now.Sub(backup.modTime) > maxBackupAge {
			shouldDelete = true
		}

		// Skip safety backups (pre-restore-*)
		if strings.HasPrefix(filepath.Base(backup.path), "pre-restore-") {
			continue
		}

		if shouldDelete {
			if err := os.Remove(backup.path); err != nil {
				return fmt.Errorf("failed to delete old backup %s: %w", backup.path, err)
			}
		}
	}

	return nil
}

// GetConfig returns the current backup configuration
func (s *BackupService) GetConfig() config.BackupConfig {
	return s.backupConfig
}

// UpdateConfig updates the backup configuration
func (s *BackupService) UpdateConfig(cfg config.BackupConfig) error {
	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create a copy of the configuration
	s.backupConfig = cfg.Clone()

	// Save to config file
	return s.configManager.SaveSection("backup", cfg)
}
