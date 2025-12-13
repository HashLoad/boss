// Package filesystem provides filesystem abstractions to enable testing and reduce coupling.
// This package follows the Dependency Inversion Principle (DIP) by implementing
// the FileSystem interface defined in the infra package.
package filesystem

import (
	"io"
	"os"

	"github.com/hashload/boss/internal/infra"
)

// Compile-time check that OSFileSystem implements infra.FileSystem.
var _ infra.FileSystem = (*OSFileSystem)(nil)

// FileSystem is an alias for infra.FileSystem for backward compatibility.
// New code should use infra.FileSystem directly.
type FileSystem = infra.FileSystem

// OSFileSystem is the default implementation using the os package.
type OSFileSystem struct{}

// NewOSFileSystem creates a new OSFileSystem instance.
func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

// ReadFile reads the entire file and returns its contents.
func (fs *OSFileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

// WriteFile writes data to a file with the given permissions.
func (fs *OSFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

// MkdirAll creates a directory along with any necessary parents.
func (fs *OSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Stat returns file info for the given path.
func (fs *OSFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// Remove removes the named file or empty directory.
func (fs *OSFileSystem) Remove(name string) error {
	return os.Remove(name)
}

// RemoveAll removes path and any children it contains.
func (fs *OSFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Rename renames (moves) a file.
func (fs *OSFileSystem) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

// Open opens a file for reading.
func (fs *OSFileSystem) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

// Create creates or truncates the named file.
func (fs *OSFileSystem) Create(name string) (io.WriteCloser, error) {
	return os.Create(name)
}

// Exists returns true if the file exists.
func (fs *OSFileSystem) Exists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

// IsDir returns true if path is a directory.
func (fs *OSFileSystem) IsDir(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// Default is the default filesystem implementation.
//
//nolint:gochecknoglobals // This is intentional for ease of use
var Default FileSystem = NewOSFileSystem()
