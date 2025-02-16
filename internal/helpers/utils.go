package helpers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Move utility functions here
// These would be internal-only helper functions

// GetCurrentDir returns the current working directory name
// Falls back to "project" if directory cannot be determined
func GetCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "project"
	}
	return filepath.Base(dir)
}

// EnsureDirectories creates all directories in the provided list
func EnsureDirectories(dirs []string, perm os.FileMode) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, perm); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// ConfirmAction prompts the user for confirmation
// Returns true if user confirms, false otherwise
func ConfirmAction(prompt string) bool {
	fmt.Print(prompt + " [y/N]: ")
	var confirm string
	if _, err := fmt.Scanln(&confirm); err != nil {
		return false
	}
	return strings.EqualFold(confirm, "y") || strings.EqualFold(confirm, "yes")
}

// FileExists checks if a file exists at the given path
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// CreateFile creates a new file with the given content
// If the file exists and force is false, returns an error
func CreateFile(path string, content []byte, force bool) error {
	if FileExists(path) && !force {
		return fmt.Errorf("file already exists at %s", path)
	}

	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}
	return nil
}

// StringSliceToMap converts a slice of strings to a map for O(1) lookups
func StringSliceToMap(slice []string) map[string]bool {
	m := make(map[string]bool, len(slice))
	for _, s := range slice {
		m[s] = true
	}
	return m
}

// MapToStringSlice converts a map[string]bool to a slice of strings
func MapToStringSlice(m map[string]bool) []string {
	slice := make([]string, 0, len(m))
	for s := range m {
		slice = append(slice, s)
	}
	return slice
}

// FilterStringSlice returns a new slice containing only elements that pass the filter function
func FilterStringSlice(slice []string, filter func(string) bool) []string {
	filtered := make([]string, 0, len(slice))
	for _, s := range slice {
		if filter(s) {
			filtered = append(filtered, s)
		}
	}
	return filtered
}
