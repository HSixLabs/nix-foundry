package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDir creates a temporary directory for testing and returns a cleanup function
func TestDir(t *testing.T) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "nix-foundry-test-*")
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	return dir, func() { os.RemoveAll(dir) }
}

// SetupTestHome creates a temporary home directory and sets HOME environment variable
func SetupTestHome(t *testing.T) (string, func()) {
	t.Helper()
	homeDir, cleanup := TestDir(t)
	oldHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", homeDir); err != nil {
		cleanup()
		t.Fatalf("Failed to set HOME environment: %v", err)
	}
	return homeDir, func() {
		os.Setenv("HOME", oldHome)
		cleanup()
	}
}

// MockCommand creates a mock command that returns specified output
func MockCommand(t *testing.T, name string, output string) func() {
	t.Helper()
	dir, err := os.MkdirTemp("", "mock-*")
	if err != nil {
		t.Fatalf("Failed to create mock directory: %v", err)
	}

	script := `#!/bin/sh
echo "` + output + `"`

	mockPath := filepath.Join(dir, name)
	if err := os.WriteFile(mockPath, []byte(script), 0755); err != nil {
		os.RemoveAll(dir)
		t.Fatalf("Failed to write mock script: %v", err)
	}

	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", dir+string(os.PathListSeparator)+oldPath)

	return func() {
		os.Setenv("PATH", oldPath)
		os.RemoveAll(dir)
	}
}
