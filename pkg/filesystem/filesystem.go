/*
Package filesystem provides filesystem abstraction for Nix Foundry.
It defines interfaces and implementations for filesystem operations,
allowing for consistent file handling across different platforms.
*/
package filesystem

import (
	"io"
	"os"
)

/*
FileSystem provides an interface for filesystem operations.
This abstraction allows for easier testing and platform independence.
*/
type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm os.FileMode) error
	Remove(path string) error
	MkdirAll(path string, perm os.FileMode) error
	CreateDir(path string) error
	Exists(path string) bool
	Copy(src, dst string) error
}

/*
OSFileSystem implements FileSystem using the OS filesystem.
It provides concrete implementations of filesystem operations
using the standard library's os package.
*/
type OSFileSystem struct{}

/*
NewOSFileSystem creates a new OS filesystem instance.
*/
func NewOSFileSystem() FileSystem {
	return &OSFileSystem{}
}

/*
ReadFile reads a file from disk.
*/
func (fs *OSFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

/*
WriteFile writes data to a file.
*/
func (fs *OSFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

/*
Remove removes a file or directory.
*/
func (fs *OSFileSystem) Remove(path string) error {
	return os.Remove(path)
}

/*
MkdirAll creates a directory and any necessary parents.
*/
func (fs *OSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

/*
CreateDir creates a directory if it doesn't exist.
*/
func (fs *OSFileSystem) CreateDir(path string) error {
	if !fs.Exists(path) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

/*
Exists checks if a path exists.
*/
func (fs *OSFileSystem) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

/*
Copy copies a file from src to dst.
It creates any necessary parent directories and preserves file permissions.
*/
func (fs *OSFileSystem) Copy(src, dst string) error {
	srcFile, openErr := os.Open(src)
	if openErr != nil {
		return openErr
	}
	defer srcFile.Close()

	mkdirErr := fs.MkdirAll(dst, 0755)
	if mkdirErr != nil {
		return mkdirErr
	}

	dstFile, createErr := os.Create(dst)
	if createErr != nil {
		return createErr
	}
	defer dstFile.Close()

	_, copyErr := io.Copy(dstFile, srcFile)
	return copyErr
}
