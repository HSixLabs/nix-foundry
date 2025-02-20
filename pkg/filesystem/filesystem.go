package filesystem

import (
	"os"
)

type FileSystem interface {
	Exists(path string) bool
	CreateDir(path string) error
	WriteFile(name string, data []byte, perm os.FileMode) error
	ReadFile(name string) ([]byte, error)
	Remove(path string) error
	RemoveAll(path string) error
}

type OSFileSystem struct{}

func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

func (fs *OSFileSystem) Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func (fs *OSFileSystem) CreateDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func (fs *OSFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (fs *OSFileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (fs *OSFileSystem) Remove(path string) error {
	return os.Remove(path)
}

func (fs *OSFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}
