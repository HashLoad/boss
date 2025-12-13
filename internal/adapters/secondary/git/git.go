package gitadapter

import (
	"path/filepath"

	"github.com/go-git/go-billy/v5/osfs"
	goGit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
)

func CloneCache(dep domain.Dependency) (*goGit.Repository, error) {
	if env.GlobalConfiguration().GitEmbedded {
		return CloneCacheEmbedded(dep)
	}

	return CloneCacheNative(dep)
}

func UpdateCache(dep domain.Dependency) (*goGit.Repository, error) {
	if env.GlobalConfiguration().GitEmbedded {
		return UpdateCacheEmbedded(dep)
	}

	return UpdateCacheNative(dep)
}

func initSubmodules(dep domain.Dependency, repository *goGit.Repository) error {
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
		Auth:              env.GlobalConfiguration().GetAuth(dep.GetURLPrefix()),
	})
	if err != nil {
		return err
	}
	return nil
}

func GetMain(repository *goGit.Repository) (*config.Branch, error) {
	branch, err := repository.Branch("main")
	if err != nil {
		branch, err = repository.Branch("master")
	}
	return branch, err
}

func GetVersions(repository *goGit.Repository, dep domain.Dependency) []*plumbing.Reference {
	var result = make([]*plumbing.Reference, 0)

	err := repository.Fetch(&goGit.FetchOptions{
		Force: true,
		Prune: true,
		Auth:  env.GlobalConfiguration().GetAuth(dep.GetURLPrefix()),
		RefSpecs: []config.RefSpec{
			"refs/*:refs/*",
			"HEAD:refs/heads/HEAD",
		},
	})

	if err != nil {
		msg.Warn("Fail to fetch repository %s: %s", dep.Repository, err)
	}

	tags, err := repository.Tags()
	if err != nil {
		msg.Err("Fail to retrieve versions: %v", err)
	} else {
		err = tags.ForEach(func(reference *plumbing.Reference) error {
			result = append(result, reference)
			return nil
		})
		if err != nil {
			msg.Err("Fail to retrieve versions: %v", err)
		}
	}

	branches, err := repository.Branches()
	if err != nil {
		msg.Err("Fail to retrieve branches: %v", err)
	} else {
		err = branches.ForEach(func(reference *plumbing.Reference) error {
			result = append(result, reference)
			return nil
		})
		if err != nil {
			msg.Err("Fail to retrieve branches: %v", err)
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
	cache := makeStorageCache(dep)
	dir := osfs.New(filepath.Join(env.GetModulesDir(), dep.Name()))
	repository, err := goGit.Open(cache, dir)
	if err != nil {
		msg.Err("Error on open repository %s: %s", dep.Repository, err)
	}

	return repository
}

func Checkout(dep domain.Dependency, referenceName plumbing.ReferenceName) error {
	if env.GlobalConfiguration().GitEmbedded {
		return CheckoutEmbedded(dep, referenceName)
	}
	return CheckoutNative(dep, referenceName)
}

func Pull(dep domain.Dependency) error {
	if env.GlobalConfiguration().GitEmbedded {
		return PullEmbedded(dep)
	}
	return PullNative(dep)
}
