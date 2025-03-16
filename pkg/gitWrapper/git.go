package gitWrapper

import (
	"path/filepath"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/models"
	"github.com/hashload/boss/pkg/msg"
)

func CloneCache(dep models.Dependency) *git.Repository {
	if env.GlobalConfiguration.GitEmbedded {
		return CloneCacheEmbedded(dep)
	} else {
		return CloneCacheNative(dep)
	}
}

func UpdateCache(dep models.Dependency) *git.Repository {
	if env.GlobalConfiguration.GitEmbedded {
		return UpdateCacheEmbedded(dep)
	} else {
		return UpdateCacheNative(dep)
	}
}

func initSubmodules(dep models.Dependency, repository *git.Repository) {
	worktree, err := repository.Worktree()
	if err != nil {
		msg.Err("... %s", err)
	}
	submodules, err := worktree.Submodules()
	if err != nil {
		msg.Err("On get submodules... %s", err)
	}

	err = submodules.Update(&git.SubmoduleUpdateOptions{
		Init:              true,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Auth:              env.GlobalConfiguration.GetAuth(dep.GetURLPrefix()),
	})
	if err != nil {
		msg.Err("Failed on update submodules from dependency %s: %s", dep.Repository, err.Error())
	}

}

func GetMain(repository *git.Repository) (*config.Branch, error) {
	branch, err := repository.Branch("main")
	if err != nil {
		branch, err = repository.Branch("master")
	}
	return branch, err
}

func GetVersions(repository *git.Repository) []*plumbing.Reference {
	tags, err := repository.Tags()
	if err != nil {
		msg.Err("Fail to retrieve versions: %", err)
	}
	var result = make([]*plumbing.Reference, 0)
	for {
		reference, err := tags.Next()
		if err != nil {
			return result
		}
		result = append(result, reference)
	}
}

func GetTagsShortName(repository *git.Repository) []string {
	tags, _ := repository.Tags()
	var result = []string{}
	_ = tags.ForEach(func(reference *plumbing.Reference) error {
		result = append(result, reference.Name().Short())
		return nil
	})
	return result
}

func GetByTag(repository *git.Repository, shortName string) *plumbing.Reference {
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

func GetRepository(dep models.Dependency) *git.Repository {
	cache := makeStorageCache(dep)
	dir := osfs.New(filepath.Join(env.GetModulesDir(), dep.GetName()))
	repository, e := git.Open(cache, dir)
	if e != nil {
		msg.Err("Error on open repository %s: %s", dep.Repository, e)
	}

	return repository
}
