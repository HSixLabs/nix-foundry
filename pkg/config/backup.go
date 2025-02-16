package config

import (
	"fmt"
)

// BackupConfig holds configuration for backup management
type BackupConfig struct {
	MaxBackups       int `yaml:"maxBackups"`       // Maximum number of backups to keep
	MaxAgeInDays     int `yaml:"maxAgeInDays"`     // Maximum age of backups in days
	CompressionLevel int `yaml:"compressionLevel"` // Gzip compression level (1-9)
}

func DefaultBackupConfig() BackupConfig {
	return BackupConfig{
		MaxBackups:       10,
		MaxAgeInDays:     30,
		CompressionLevel: 6,
	}
}

// Add these validation methods to the existing file
func (c BackupConfig) Validate() error {
	if c.MaxBackups < 1 {
		return fmt.Errorf("maxBackups must be at least 1")
	}
	if c.MaxAgeInDays < 1 {
		return fmt.Errorf("maxAgeInDays must be at least 1")
	}
	if c.CompressionLevel < 1 || c.CompressionLevel > 9 {
		return fmt.Errorf("compressionLevel must be between 1 and 9")
	}
	return nil
}

// Clone creates a copy of the backup configuration
func (c BackupConfig) Clone() BackupConfig {
	return BackupConfig{
		MaxBackups:       c.MaxBackups,
		MaxAgeInDays:     c.MaxAgeInDays,
		CompressionLevel: c.CompressionLevel,
	}
}

// BackupSettings contains backup-related configuration
type BackupSettings struct {
	MaxBackups     int      `yaml:"maxBackups"`
	RetentionDays  int      `yaml:"retentionDays"`
	BackupDir      string   `yaml:"backupDir"`
	ExcludePattern []string `yaml:"excludePattern"`
}
