// Package filesystem provides filesystem abstraction.
package filesystem

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileSystem provides an interface for filesystem operations.
type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	Remove(path string) error
	MkdirAll(path string, perm os.FileMode) error
	CreateDir(path string) error
	Exists(path string) bool
	Copy(src, dst string) error
}

// OSFileSystem implements FileSystem using the OS filesystem.
type OSFileSystem struct{}

// NewOSFileSystem creates a new OS filesystem.
func NewOSFileSystem() FileSystem {
	return &OSFileSystem{}
}

// ReadFile reads a file from disk.
func (fs *OSFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile writes data to a file.
func (fs *OSFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

// Remove removes a file or directory.
func (fs *OSFileSystem) Remove(path string) error {
	return os.Remove(path)
}

// MkdirAll creates a directory and any necessary parents.
func (fs *OSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// CreateDir creates a directory if it doesn't exist.
func (fs *OSFileSystem) CreateDir(path string) error {
	if !fs.Exists(path) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// Exists checks if a path exists.
func (fs *OSFileSystem) Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// Copy copies a file from src to dst.
func (fs *OSFileSystem) Copy(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	if err := fs.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceFileStat.Mode())
}
