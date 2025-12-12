package installer

import (
	"sync"

	goGit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashload/boss/internal/core/domain"
)

// GitClient abstracts Git operations for testability.
type GitClient interface {
	// CloneCache clones a dependency repository to cache.
	CloneCache(dep domain.Dependency) *goGit.Repository

	// UpdateCache updates an existing cached repository.
	UpdateCache(dep domain.Dependency) *goGit.Repository

	// GetRepository returns the repository for a dependency.
	GetRepository(dep domain.Dependency) *goGit.Repository

	// GetVersions returns all version tags for a repository.
	GetVersions(repository *goGit.Repository, dep domain.Dependency) []*plumbing.Reference

	// GetByTag returns a reference by tag name.
	GetByTag(repository *goGit.Repository, tag string) *plumbing.Reference

	// GetMain returns the main branch reference.
	GetMain(repository *goGit.Repository) (Branch, error)

	// GetTagsShortName returns short names of all tags.
	GetTagsShortName(repository *goGit.Repository) []string
}

// Branch represents a git branch.
type Branch interface {
	Name() string
}

// Compiler abstracts compilation operations for testability.
type Compiler interface {
	// Build compiles all packages in dependency order.
	Build(pkg *domain.Package)
}

// DependencyCache tracks which dependencies have been updated in current session.
// Thread-safe implementation to replace global variable.
type DependencyCache struct {
	updated map[string]bool
	mu      sync.RWMutex
}

// NewDependencyCache creates a new DependencyCache instance.
func NewDependencyCache() *DependencyCache {
	return &DependencyCache{
		updated: make(map[string]bool),
	}
}

// IsUpdated checks if a dependency has been updated in current session.
func (c *DependencyCache) IsUpdated(hashName string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.updated[hashName]
}

// MarkUpdated marks a dependency as updated in current session.
func (c *DependencyCache) MarkUpdated(hashName string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.updated[hashName] = true
}

// Reset clears all cached updates.
func (c *DependencyCache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.updated = make(map[string]bool)
}

// Count returns the number of updated dependencies.
func (c *DependencyCache) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.updated)
}

// FileSystem abstracts file system operations for testability.
type FileSystem interface {
	// Stat returns file info.
	Stat(name string) (FileInfo, error)

	// RemoveAll removes a path and all children.
	RemoveAll(path string) error

	// ReadDir reads directory contents.
	ReadDir(name string) ([]DirEntry, error)

	// IsNotExist checks if error is "not exist".
	IsNotExist(err error) bool
}

// FileInfo minimal interface for file info.
type FileInfo interface {
	IsDir() bool
}

// DirEntry minimal interface for directory entry.
type DirEntry interface {
	Name() string
	IsDir() bool
}
