package installer

import (
	"errors"
	"os"
	"path/filepath"

	goGit "github.com/go-git/go-git/v5"
	"github.com/hashload/boss/internal/adapters/secondary/filesystem"
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/cache"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
)

// ErrRepositoryNil is returned when the repository is nil after cloning or updating.
var ErrRepositoryNil = errors.New("failed to clone or update repository")

// DependencyManager manages dependency fetching with proper dependency injection.
type DependencyManager struct {
	gitClient    GitClient
	cache        *DependencyCache
	cacheDir     string
	cacheService *cache.Service
}

// NewDependencyManager creates a new DependencyManager with the given dependencies.
func NewDependencyManager(gitClient GitClient, depCache *DependencyCache, cacheService *cache.Service) *DependencyManager {
	return &DependencyManager{
		gitClient:    gitClient,
		cache:        depCache,
		cacheDir:     env.GetCacheDir(),
		cacheService: cacheService,
	}
}

// NewDefaultDependencyManager creates a DependencyManager with default implementations.
func NewDefaultDependencyManager() *DependencyManager {
	return NewDependencyManager(
		NewDefaultGitClient(),
		NewDependencyCache(),
		cache.NewService(filesystem.NewOSFileSystem()),
	)
}

// GetDependency fetches or updates a dependency in cache.
func (dm *DependencyManager) GetDependency(dep domain.Dependency) error {
	return dm.GetDependencyWithProgress(dep, nil)
}

// GetDependencyWithProgress fetches or updates a dependency with optional progress tracking.
func (dm *DependencyManager) GetDependencyWithProgress(dep domain.Dependency, progress *ProgressTracker) error {
	if dm.cache.IsUpdated(dep.HashName()) {
		msg.Debug("Using cached of %s", dep.Name())
		return nil
	}

	if progress == nil || !progress.IsEnabled() {
		msg.Info("Updating cache of dependency %s", dep.Name())
	}
	dm.cache.MarkUpdated(dep.HashName())

	var repository *goGit.Repository
	var err error
	if dm.hasCache(dep) {
		repository, err = dm.gitClient.UpdateCache(dep)
	} else {
		_ = os.RemoveAll(filepath.Join(dm.cacheDir, dep.HashName()))
		repository, err = dm.gitClient.CloneCache(dep)
	}

	if err != nil {
		return err
	}

	if repository == nil {
		return ErrRepositoryNil
	}

	tagsShortNames := dm.gitClient.GetTagsShortName(repository)
	if err := dm.cacheService.SaveRepositoryDetails(dep, tagsShortNames); err != nil {
		msg.Warn("Failed to cache repository details: %v", err)
	}
	return nil
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
