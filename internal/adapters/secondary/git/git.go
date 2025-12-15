// Package gitadapter provides Git operations for cloning and updating dependency repositories.
// It supports both embedded (go-git) and native Git implementations.
package gitadapter

import (
	"path/filepath"

	"github.com/go-git/go-billy/v5/osfs"
	goGit "github.com/go-git/go-git/v5"
	gitConfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
)

// CloneCache clones the dependency repository to the cache.
func CloneCache(config env.ConfigProvider, dep domain.Dependency) (*goGit.Repository, error) {
	if config.GetGitEmbedded() {
		return CloneCacheEmbedded(config, dep)
	}

	return CloneCacheNative(dep)
}

// UpdateCache updates the dependency repository in the cache.
func UpdateCache(config env.ConfigProvider, dep domain.Dependency) (*goGit.Repository, error) {
	if config.GetGitEmbedded() {
		return UpdateCacheEmbedded(config, dep)
	}

	return UpdateCacheNative(dep)
}

func initSubmodules(config env.ConfigProvider, dep domain.Dependency, repository *goGit.Repository) error {
	worktree, err := repository.Worktree()
	if err != nil {
		return err
	}
	submodules, err := worktree.Submodules()
	if err != nil {
		return err
	}

	err = submodules.Update(&goGit.SubmoduleUpdateOptions{
		Init:              true,
		RecurseSubmodules: goGit.DefaultSubmoduleRecursionDepth,
		Auth:              config.GetAuth(dep.GetURLPrefix()),
	})
	if err != nil {
		return err
	}
	return nil
}

// GetMain returns the main branch of the repository.
func GetMain(repository *goGit.Repository) (*gitConfig.Branch, error) {
	branch, err := repository.Branch(consts.GitBranchMain)
	if err != nil {
		branch, err = repository.Branch(consts.GitBranchMaster)
	}
	return branch, err
}

// GetVersions returns all versions (tags and branches) of the repository.
func GetVersions(config env.ConfigProvider, repository *goGit.Repository, dep domain.Dependency) []*plumbing.Reference {
	var result = make([]*plumbing.Reference, 0)

	err := repository.Fetch(&goGit.FetchOptions{
		Force: true,
		Prune: true,
		Auth:  config.GetAuth(dep.GetURLPrefix()),
		RefSpecs: []gitConfig.RefSpec{
			"refs/*:refs/*",
			"HEAD:refs/heads/HEAD",
		},
	})

	if err != nil {
		msg.Warn("⚠️ Fail to fetch repository %s: %s", dep.Repository, err)
	}

	tags, err := repository.Tags()
	if err != nil {
		msg.Err("❌ Fail to retrieve versions: %v", err)
	} else {
		err = tags.ForEach(func(reference *plumbing.Reference) error {
			result = append(result, reference)
			return nil
		})
		if err != nil {
			msg.Err("❌ Fail to retrieve versions: %v", err)
		}
	}

	branches, err := repository.Branches()
	if err != nil {
		msg.Err("❌ Fail to retrieve branches: %v", err)
	} else {
		err = branches.ForEach(func(reference *plumbing.Reference) error {
			result = append(result, reference)
			return nil
		})
		if err != nil {
			msg.Err("❌ Fail to retrieve branches: %v", err)
		}
	}

	return result
}

func GetTagsShortName(repository *goGit.Repository) []string {
	tags, _ := repository.Tags()
	var result = []string{}
	_ = tags.ForEach(func(reference *plumbing.Reference) error {
		result = append(result, reference.Name().Short())
		return nil
	})
	return result
}

func GetByTag(repository *goGit.Repository, shortName string) *plumbing.Reference {
	tags, _ := repository.Tags()

	for {
		if reference, err := tags.Next(); err == nil {
			if reference.Name().Short() == shortName {
				return reference
			}
		} else {
			return nil
		}
	}
}

func GetRepository(dep domain.Dependency) *goGit.Repository {
	// GetRepository is used in places where we already have a cloned repo
	// So we don't need config for EnsureCacheDir check
	cache := makeStorageCacheWithoutEnsure(dep)
	dir := osfs.New(filepath.Join(env.GetModulesDir(), dep.Name()))
	repository, err := goGit.Open(cache, dir)
	if err != nil {
		msg.Err("❌ Error on open repository %s: %s", dep.Repository, err)
	}

	return repository
}

func Checkout(config env.ConfigProvider, dep domain.Dependency, referenceName plumbing.ReferenceName) error {
	if config.GetGitEmbedded() {
		return CheckoutEmbedded(config, dep, referenceName)
	}
	return CheckoutNative(dep, referenceName)
}

func Pull(config env.ConfigProvider, dep domain.Dependency) error {
	if config.GetGitEmbedded() {
		return PullEmbedded(config, dep)
	}
	return PullNative(dep)
}
