package installer

import (
	"sync"

	"github.com/hashload/boss/internal/core/domain"
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
func GetDependency(dep domain.Dependency) error {
	return getDefaultDependencyManager().GetDependency(dep)
}

// GetDependencyWithProgress fetches or updates a dependency with optional progress tracking.
func GetDependencyWithProgress(dep domain.Dependency, progress *ProgressTracker) error {
	return getDefaultDependencyManager().GetDependencyWithProgress(dep, progress)
}

// ResetDependencyCache clears the dependency cache for a new session.
// This should be called at the start of a new install operation.
func ResetDependencyCache() {
	getDefaultDependencyManager().Reset()
}
