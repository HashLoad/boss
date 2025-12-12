package installer

import (
	"os"
	"path/filepath"

	goGit "github.com/go-git/go-git/v5"
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
)

// DependencyManager manages dependency fetching with proper dependency injection.
type DependencyManager struct {
	gitClient GitClient
	cache     *DependencyCache
	cacheDir  string
}

// NewDependencyManager creates a new DependencyManager with the given dependencies.
func NewDependencyManager(gitClient GitClient, cache *DependencyCache) *DependencyManager {
	return &DependencyManager{
		gitClient: gitClient,
		cache:     cache,
		cacheDir:  env.GetCacheDir(),
	}
}

// NewDefaultDependencyManager creates a DependencyManager with default implementations.
func NewDefaultDependencyManager() *DependencyManager {
	return NewDependencyManager(
		NewDefaultGitClient(),
		NewDependencyCache(),
	)
}

// GetDependency fetches or updates a dependency in cache.
func (dm *DependencyManager) GetDependency(dep domain.Dependency) {
	if dm.cache.IsUpdated(dep.HashName()) {
		msg.Debug("Using cached of %s", dep.Name())
		return
	}

	msg.Info("Updating cache of dependency %s", dep.Name())
	dm.cache.MarkUpdated(dep.HashName())

	var repository *goGit.Repository
	if dm.hasCache(dep) {
		repository = dm.gitClient.UpdateCache(dep)
	} else {
		_ = os.RemoveAll(filepath.Join(dm.cacheDir, dep.HashName()))
		repository = dm.gitClient.CloneCache(dep)
	}

	tagsShortNames := dm.gitClient.GetTagsShortName(repository)
	domain.CacheRepositoryDetails(dep, tagsShortNames)
}

// hasCache checks if a dependency is already cached.
func (dm *DependencyManager) hasCache(dep domain.Dependency) bool {
	dir := filepath.Join(dm.cacheDir, dep.HashName())
	info, err := os.Stat(dir)
	if err == nil {
		// Path exists, check if it's a directory
		if !info.IsDir() {
			// It's a file, remove it and return false
			_ = os.RemoveAll(dir)
			return false
		}
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	// Other error, try to clean up and return false
	_ = os.RemoveAll(dir)
	return false
}

// Reset clears the dependency cache for a new session.
func (dm *DependencyManager) Reset() {
	dm.cache.Reset()
}

// Cache returns the underlying cache for inspection.
func (dm *DependencyManager) Cache() *DependencyCache {
	return dm.cache
}
