// Package ports defines port interfaces for hexagonal architecture.
package ports

// Registry defines the contract for system registry operations.
// On Windows, this interacts with the Windows Registry.
// On Unix systems, this may use environment variables or config files.
type Registry interface {
	// GetDelphiPath returns the path to the Delphi installation.
	GetDelphiPath() string

	// SetEnvPath sets an environment variable path.
	SetEnvPath(path string) error

	// GetEnvPath gets an environment variable path.
	GetEnvPath() string

	// AddToPath adds a path to the system PATH.
	AddToPath(path string) error
}
