package installer

import (
	"context"

	goGit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	git "github.com/hashload/boss/internal/adapters/secondary/git"
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/ports"
	"github.com/hashload/boss/pkg/env"
)

var _ ports.GitClientV2 = (*DefaultGitClient)(nil)

// DefaultGitClient is the production implementation of GitClient.
type DefaultGitClient struct {
	config env.ConfigProvider
}

// NewDefaultGitClient creates a new DefaultGitClient.
func NewDefaultGitClient(config env.ConfigProvider) *DefaultGitClient {
	return &DefaultGitClient{config: config}
}

// CloneCache clones a dependency repository to cache.
func (c *DefaultGitClient) CloneCache(dep domain.Dependency) (*goGit.Repository, error) {
	return git.CloneCache(c.config, dep)
}

// UpdateCache updates an existing cached repository.
func (c *DefaultGitClient) UpdateCache(dep domain.Dependency) (*goGit.Repository, error) {
	return git.UpdateCache(c.config, dep)
}

// GetRepository returns the repository for a dependency.
func (c *DefaultGitClient) GetRepository(dep domain.Dependency) *goGit.Repository {
	return git.GetRepository(dep)
}

// GetVersions returns all version tags for a repository.
func (c *DefaultGitClient) GetVersions(repository *goGit.Repository, dep domain.Dependency) []*plumbing.Reference {
	return git.GetVersions(c.config, repository, dep)
}

// GetByTag returns a reference by tag name.
func (c *DefaultGitClient) GetByTag(repository *goGit.Repository, tag string) *plumbing.Reference {
	return git.GetByTag(repository, tag)
}

// GetMain returns the main branch reference.
func (c *DefaultGitClient) GetMain(repository *goGit.Repository) (ports.Branch, error) {
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

// configBranch wraps config.Branch to implement ports.Branch interface.
type configBranch struct {
	*config.Branch
}

// Name returns the branch name.
func (b *configBranch) Name() string {
	return b.Branch.Name
}

// Remote returns the remote name.
func (b *configBranch) Remote() string {
	return b.Branch.Remote
}

// CloneCacheWithContext clones with context support for cancellation.
// Note: go-git's Clone operation doesn't support context natively.
// We check for cancellation before starting, but the clone operation itself
// may not be interruptible once started.
//
//nolint:lll // Function signature cannot be easily shortened
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
//
//nolint:lll // Function signature cannot be easily shortened
func (c *DefaultGitClient) UpdateCacheWithContext(ctx context.Context, dep domain.Dependency) (*goGit.Repository, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return c.UpdateCache(dep)
}
