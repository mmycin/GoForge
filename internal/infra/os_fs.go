package infra

import (
	"io/fs"
	"os"
	"path/filepath"
)

// OsFileSystem is the production FileSystem implementation backed by the real OS.
type OsFileSystem struct{}

// NewOsFileSystem returns a FileSystem that delegates to the real os package.
func NewOsFileSystem() FileSystem {
	return &OsFileSystem{}
}

func (o *OsFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (o *OsFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (o *OsFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (o *OsFileSystem) Remove(path string) error {
	return os.Remove(path)
}

func (o *OsFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (o *OsFileSystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (o *OsFileSystem) ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}

func (o *OsFileSystem) WalkDir(root string, fn fs.WalkDirFunc) error {
	return filepath.WalkDir(root, fn)
}
