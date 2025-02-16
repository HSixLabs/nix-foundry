package backup

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func GenerateChecksums(rootDir, checksumFile string) error {
	outFile, err := os.Create(checksumFile)
	if err != nil {
		return fmt.Errorf("failed to create checksum file: %w", err)
	}
	defer outFile.Close()

	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		hasher := sha256.New()
		if _, err := io.Copy(hasher, file); err != nil {
			return err
		}

		relPath, _ := filepath.Rel(rootDir, path)
		_, err = fmt.Fprintf(outFile, "%x  %s\n", hasher.Sum(nil), relPath)
		return err
	})
}
