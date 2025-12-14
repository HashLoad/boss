// Package domain contains core business entities and their contracts.
package domain

import (
	"io"
	"os"
)

// FileSystem abstracts file system operations for testability.
// This port is implemented by adapters in the infrastructure layer.
type FileSystem interface {
	// ReadFile reads the entire file and returns its contents.
	ReadFile(name string) ([]byte, error)

	// WriteFile writes data to a file with the given permissions.
	WriteFile(name string, data []byte, perm os.FileMode) error

	// MkdirAll creates a directory along with any necessary parents.
	MkdirAll(path string, perm os.FileMode) error

	// Stat returns file info for the given path.
	Stat(name string) (os.FileInfo, error)

	// Remove removes the named file or empty directory.
	Remove(name string) error

	// RemoveAll removes path and any children it contains.
	RemoveAll(path string) error

	// Rename renames (moves) a file.
	Rename(oldpath, newpath string) error

	// Open opens a file for reading.
	Open(name string) (io.ReadCloser, error)

	// Create creates or truncates the named file.
	Create(name string) (io.WriteCloser, error)

	// Exists returns true if the file exists.
	Exists(name string) bool

	// IsDir returns true if path is a directory.
	IsDir(name string) bool
}
