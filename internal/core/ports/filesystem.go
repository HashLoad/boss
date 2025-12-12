package ports

import "io/fs"

// FileSystem defines the contract for file system operations.
// This abstraction allows for testing and alternative implementations.
type FileSystem interface {
	// ReadFile reads the content of a file.
	ReadFile(name string) ([]byte, error)

	// WriteFile writes data to a file with the specified permissions.
	WriteFile(name string, data []byte, perm fs.FileMode) error

	// Remove removes a file or empty directory.
	Remove(name string) error

	// RemoveAll removes a path and any children it contains.
	RemoveAll(path string) error

	// MkdirAll creates a directory along with any necessary parents.
	MkdirAll(path string, perm fs.FileMode) error

	// Stat returns file info for the named file.
	Stat(name string) (fs.FileInfo, error)

	// ReadDir reads the directory and returns directory entries.
	ReadDir(name string) ([]fs.DirEntry, error)

	// Rename renames (moves) a file or directory.
	Rename(oldpath, newpath string) error

	// Exists checks if a path exists.
	Exists(path string) bool

	// IsDir checks if a path is a directory.
	IsDir(path string) bool
}
