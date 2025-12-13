// Package infra provides infrastructure interfaces that domain entities can depend on.
// These are low-level abstractions that don't depend on domain types, avoiding import cycles.
// This follows the Dependency Inversion Principle (DIP).
package infra

import (
	"io"
	"os"
)

// FileSystem defines the contract for file system operations.
// This abstraction allows for testing and alternative implementations.
// Domain entities should depend on this interface, not on concrete implementations.
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
