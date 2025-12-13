// Package ports defines the interfaces (contracts) that the domain requires.
// These interfaces are implemented by adapters in the infrastructure layer.
package ports

import (
	"context"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashload/boss/internal/core/domain"
)

// GitRepository defines the contract for git operations.
// This interface is part of the domain and is implemented by adapters.
type GitRepository interface {
	// CloneCache clones a dependency repository to cache.
	// Returns the cloned repository or an error if cloning fails.
	CloneCache(ctx context.Context, dep domain.Dependency) (*git.Repository, error)

	// UpdateCache updates an existing cached repository.
	// Returns the updated repository or an error if update fails.
	UpdateCache(ctx context.Context, dep domain.Dependency) (*git.Repository, error)

	// GetVersions retrieves all versions (tags and branches) from a repository.
	GetVersions(repository *git.Repository, dep domain.Dependency) []*plumbing.Reference

	// GetMain returns the main or master branch configuration.
	GetMain(repository *git.Repository) (*config.Branch, error)

	// GetByTag returns a reference by its tag short name.
	GetByTag(repository *git.Repository, shortName string) *plumbing.Reference

	// GetTagsShortName returns all tag short names from a repository.
	GetTagsShortName(repository *git.Repository) []string

	// GetRepository opens and returns a repository for a dependency.
	GetRepository(dep domain.Dependency) *git.Repository
}

// Branch represents a git branch configuration.
type Branch interface {
	Name() string
	Remote() string
}
