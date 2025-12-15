// Package ports defines port interfaces for dependency management.
package ports

import "github.com/hashload/boss/internal/core/domain"

// DependencyInstaller defines the contract for installing dependencies.
type DependencyInstaller interface {
	// Install installs dependencies from the package file.
	Install(args []string, buildAfter bool, noSave bool)

	// GetDependency retrieves a dependency, using cache if available.
	GetDependency(dep domain.Dependency) error

	// Uninstall removes a dependency.
	Uninstall(args []string)

	// Update updates dependencies to their latest versions.
	Update()
}

// DependencyCache defines the contract for caching dependency state.
type DependencyCache interface {
	// IsUpdated checks if a dependency has been updated in this session.
	IsUpdated(name string) bool

	// MarkUpdated marks a dependency as updated.
	MarkUpdated(name string)

	// Reset clears the cache.
	Reset()

	// Count returns the number of cached entries.
	Count() int
}
