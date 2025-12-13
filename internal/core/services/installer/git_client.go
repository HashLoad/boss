package installer

import (
	"context"

	goGit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	git "github.com/hashload/boss/internal/adapters/secondary/git"
	"github.com/hashload/boss/internal/core/domain"
)

// Ensure DefaultGitClient implements GitClientV2.
var _ GitClientV2 = (*DefaultGitClient)(nil)

// DefaultGitClient is the production implementation of GitClient.
type DefaultGitClient struct{}

// NewDefaultGitClient creates a new DefaultGitClient.
func NewDefaultGitClient() *DefaultGitClient {
	return &DefaultGitClient{}
}

// CloneCache clones a dependency repository to cache.
func (c *DefaultGitClient) CloneCache(dep domain.Dependency) (*goGit.Repository, error) {
	return git.CloneCache(dep)
}

// UpdateCache updates an existing cached repository.
func (c *DefaultGitClient) UpdateCache(dep domain.Dependency) (*goGit.Repository, error) {
	return git.UpdateCache(dep)
}

// GetRepository returns the repository for a dependency.
func (c *DefaultGitClient) GetRepository(dep domain.Dependency) *goGit.Repository {
	return git.GetRepository(dep)
}

// GetVersions returns all version tags for a repository.
func (c *DefaultGitClient) GetVersions(repository *goGit.Repository, dep domain.Dependency) []*plumbing.Reference {
	return git.GetVersions(repository, dep)
}

// GetByTag returns a reference by tag name.
func (c *DefaultGitClient) GetByTag(repository *goGit.Repository, tag string) *plumbing.Reference {
	return git.GetByTag(repository, tag)
}

// GetMain returns the main branch reference.
func (c *DefaultGitClient) GetMain(repository *goGit.Repository) (Branch, error) {
	branch, err := git.GetMain(repository)
	if err != nil {
		return nil, err
	}
	return &configBranch{branch}, nil
}

// GetTagsShortName returns short names of all tags.
func (c *DefaultGitClient) GetTagsShortName(repository *goGit.Repository) []string {
	return git.GetTagsShortName(repository)
}

// configBranch wraps config.Branch to implement Branch interface.
type configBranch struct {
	*config.Branch
}

// Name returns the branch name.
func (b *configBranch) Name() string {
	return b.Branch.Name
}

// CloneCacheWithContext clones with context support for cancellation.
// Note: go-git's Clone operation doesn't support context natively.
// We check for cancellation before starting, but the clone operation itself
// may not be interruptible once started.
func (c *DefaultGitClient) CloneCacheWithContext(ctx context.Context, dep domain.Dependency) (*goGit.Repository, error) {
	// Check for cancellation before starting
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return c.CloneCache(dep)
}

// UpdateCacheWithContext updates with context support for cancellation.
// Note: go-git's Fetch operation doesn't support context natively.
// We check for cancellation before starting, but the update operation itself
// may not be interruptible once started.
func (c *DefaultGitClient) UpdateCacheWithContext(ctx context.Context, dep domain.Dependency) (*goGit.Repository, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return c.UpdateCache(dep)
}
