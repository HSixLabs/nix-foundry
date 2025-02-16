package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const backupDir = ".config/nix-foundry/backups"

func Create(configDir string) (string, error) {
	// Ensure backup directory exists
	home, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return "", fmt.Errorf("failed to get home directory: %w", homeErr)
	}
	backupPath := filepath.Join(home, backupDir)
	if mkdirErr := os.MkdirAll(backupPath, 0755); mkdirErr != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", mkdirErr)
	}

	// Create backup file
	timestamp := time.Now().Format("20060102-150405")
	backupFile := filepath.Join(backupPath, fmt.Sprintf("nix-foundry-backup-%s.tar.gz", timestamp))
	file, err := os.Create(backupFile)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}
	defer file.Close()

	// Create gzip writer
	gw := gzip.NewWriter(file)
	defer gw.Close()

	// Create tar writer
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Walk through config directory and add files to tar
	walkErr := filepath.Walk(configDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		// Get the relative path for the tar header
		relPath, relErr := filepath.Rel(configDir, path)
		if relErr != nil {
			return fmt.Errorf("failed to get relative path: %w", relErr)
		}

		// Handle symlinks specially
		if info.Mode()&os.ModeSymlink != 0 {
			// Read the symlink target
			target, err := os.Readlink(path)
			if err != nil {
				return fmt.Errorf("failed to read symlink: %w", err)
			}

			// Create symlink header
			header := &tar.Header{
				Name:     relPath,
				Linkname: target,
				Mode:     int64(info.Mode()),
				ModTime:  info.ModTime(),
				Typeflag: tar.TypeSymlink,
			}

			if err := tw.WriteHeader(header); err != nil {
				return fmt.Errorf("failed to write symlink header: %w", err)
			}
			return nil
		}

		// Regular file handling (existing code)
		header, err := tar.FileInfoHeader(info, info.Name())
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
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer file.Close()

			if _, err := io.Copy(tw, file); err != nil {
				return fmt.Errorf("failed to copy file content: %w", err)
			}
		}

		return nil
	})

	if walkErr != nil {
		return "", fmt.Errorf("failed to create backup: %w", walkErr)
	}

	return backupFile, nil
}

func Restore(backupPath string, configDir string) error {
	// Ensure backup file exists
	if _, err := os.Stat(backupPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("backup file not found: %s", backupPath)
		}
		return fmt.Errorf("failed to access backup file: %w", err)
	}

	// Create a temporary directory for extraction
	tempDir, err := os.MkdirTemp("", "nix-foundry-restore-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Extract backup to temporary directory first
	cmd := exec.Command("tar", "-xzf", backupPath, "-C", tempDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract backup: %w", err)
	}

	// Validate backup contents
	if err := validateBackup(tempDir); err != nil {
		return fmt.Errorf("invalid backup file: %w", err)
	}

	// Create backup of current config before restoring
	if _, err := Create(configDir); err != nil {
		return fmt.Errorf("failed to backup current configuration: %w", err)
	}

	// Move contents to config directory
	cmd = exec.Command("rsync", "-a", "--delete", tempDir+"/", configDir+"/")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restore files: %w", err)
	}

	return nil
}

// validateBackup checks if the extracted backup has the expected structure
func validateBackup(dir string) error {
	required := []string{
		"environments",
		"config.yaml",
	}

	for _, path := range required {
		if _, err := os.Stat(filepath.Join(dir, path)); err != nil {
			return fmt.Errorf("missing required file/directory: %s", path)
		}
	}

	return nil
}

func ListBackups() ([]string, error) {
	homeDir, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", homeErr)
	}

	backupPath := filepath.Join(homeDir, backupDir)
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".gz" {
			backups = append(backups, filepath.Join(backupPath, entry.Name()))
		}
	}

	return backups, nil
}
