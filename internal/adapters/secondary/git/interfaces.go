package gitadapter

import (
	goGit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashload/boss/internal/core/domain"
)

// Repository abstracts git repository operations.
type Repository interface {
	CloneCache(dep domain.Dependency) (*goGit.Repository, error)
	UpdateCache(dep domain.Dependency) (*goGit.Repository, error)
	GetVersions(repository *goGit.Repository, dep domain.Dependency) []*plumbing.Reference
	GetMain(repository *goGit.Repository) (*config.Branch, error)
	GetByTag(repository *goGit.Repository, shortName string) *plumbing.Reference
	GetTagsShortName(repository *goGit.Repository) []string
	GetRepository(dep domain.Dependency) *goGit.Repository
}

// DefaultRepository implements Repository using the package-level functions.
type DefaultRepository struct{}

// CloneCache clones a dependency to cache.
func (d *DefaultRepository) CloneCache(dep domain.Dependency) (*goGit.Repository, error) {
	return CloneCache(dep)
}

// UpdateCache updates a cached dependency.
func (d *DefaultRepository) UpdateCache(dep domain.Dependency) (*goGit.Repository, error) {
	return UpdateCache(dep)
}

// GetVersions retrieves all versions (tags and branches) for a repository.
func (d *DefaultRepository) GetVersions(repository *goGit.Repository, dep domain.Dependency) []*plumbing.Reference {
	return GetVersions(repository, dep)
}

// GetMain returns the main or master branch.
func (d *DefaultRepository) GetMain(repository *goGit.Repository) (*config.Branch, error) {
	return GetMain(repository)
}

// GetByTag returns a reference by tag short name.
func (d *DefaultRepository) GetByTag(repository *goGit.Repository, shortName string) *plumbing.Reference {
	return GetByTag(repository, shortName)
}

// GetTagsShortName returns all tag short names.
func (d *DefaultRepository) GetTagsShortName(repository *goGit.Repository) []string {
	return GetTagsShortName(repository)
}

// GetRepository opens a repository for a dependency.
func (d *DefaultRepository) GetRepository(dep domain.Dependency) *goGit.Repository {
	return GetRepository(dep)
}
