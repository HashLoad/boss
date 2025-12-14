// Package installer provides version control system integration.
package installer

import (
	"sync"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/env"
)

//nolint:gochecknoglobals // Singleton for backward compatibility during refactor
var (
	defaultDependencyManager *DependencyManager
	dependencyManagerOnce    sync.Once
)

// getDefaultDependencyManager returns the singleton DependencyManager instance.
func getDefaultDependencyManager() *DependencyManager {
	dependencyManagerOnce.Do(func() {
		defaultDependencyManager = NewDefaultDependencyManager(env.GlobalConfiguration())
	})
	return defaultDependencyManager
}

// GetDependency fetches or updates a dependency in cache.
// This is a convenience function that uses the default DependencyManager.
// For better testability, inject DependencyManager directly in new code.
func GetDependency(dep domain.Dependency) error {
	return getDefaultDependencyManager().GetDependency(dep)
}

// GetDependencyWithProgress fetches or updates a dependency with optional progress tracking.
func GetDependencyWithProgress(dep domain.Dependency, progress *ProgressTracker) error {
	return getDefaultDependencyManager().GetDependencyWithProgress(dep, progress)
}
