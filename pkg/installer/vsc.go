package installer

import (
	"sync"

	"github.com/hashload/boss/pkg/models"
)

//nolint:gochecknoglobals // Singleton for backward compatibility during refactor
var (
	defaultDependencyManager *DependencyManager
	dependencyManagerOnce    sync.Once
)

// getDefaultDependencyManager returns the singleton DependencyManager instance.
func getDefaultDependencyManager() *DependencyManager {
	dependencyManagerOnce.Do(func() {
		defaultDependencyManager = NewDefaultDependencyManager()
	})
	return defaultDependencyManager
}

// GetDependency fetches or updates a dependency in cache.
// Deprecated: Use DependencyManager.GetDependency instead for better testability.
func GetDependency(dep models.Dependency) {
	getDefaultDependencyManager().GetDependency(dep)
}

// ResetDependencyCache clears the dependency cache for a new session.
// This should be called at the start of a new install operation.
func ResetDependencyCache() {
	getDefaultDependencyManager().Reset()
}
