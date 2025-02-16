package testutil

import (
	"os"
	"path/filepath"
)

func CreateTempConfig(content string) (string, error) {
	dir := filepath.Join(os.TempDir(), "nix-foundry-test")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	tmpFile := filepath.Join(dir, "test-config.yaml")
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		return "", err
	}
	return tmpFile, nil
}

// Add test utility functions here...
