package git

import (
	"github.com/hashload/boss/core/paths"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	cache2 "gopkg.in/src-d/go-git.v4/plumbing/cache"
	"gopkg.in/src-d/go-git.v4/storage"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
	"os"
	"path/filepath"
)

func CloneCache(dep models.Dependency) *git.Repository {
	msg.Info("Downloading dependency %s", dep.Repository)
	storageCache := makeStorageCache(dep)
	url := dep.GetURL()
	auth := env.GlobalConfiguration.GetAuth(dep.GetURLPrefix())
	repository, e := git.Clone(storageCache, memfs.New(), &git.CloneOptions{
		URL:  url,
		Tags: git.AllTags,
		Auth: auth,
	})
	if e != nil {
		_ = os.RemoveAll(filepath.Join(env.GetCacheDir(), dep.GetHashName()))
		msg.Die("Error to get repository of %s: %s", dep.Repository, e)
	}
	initSubmodules(dep, repository)
	return repository
}

func UpdateCache(dep models.Dependency) *git.Repository {
	msg.Info("Updating dependency %s", dep.Repository)
	storageCache := makeStorageCache(dep)
	repository, err := git.Open(storageCache, memfs.New())
	if err != nil {
		msg.Warn("Error to open cache of %s: %s", dep.Repository, err)
		repository = refreshCopy(dep)
	} else {
		worktree, _ := repository.Worktree()
		_ = worktree.Reset(&git.ResetOptions{
			Mode: git.HardReset,
		})
	}
	_ = repository.Fetch(&git.FetchOptions{
		Force: true,
		Auth:  env.GlobalConfiguration.GetAuth(dep.GetURLPrefix())})
	initSubmodules(dep, repository)
	return repository
}

func refreshCopy(dep models.Dependency) *git.Repository {
	dir := filepath.Join(env.GetCacheDir(), dep.GetHashName())
	e := os.RemoveAll(dir)
	if e == nil {
		return CloneCache(dep)
	} else {
		msg.Err("Error on retry get refresh copy: %s", e)
	}
	return nil
}

func makeStorageCache(dep models.Dependency) storage.Storer {
	paths.EnsureCacheDir(dep)
	dir := filepath.Join(env.GetCacheDir(), dep.GetHashName())
	fs := osfs.New(dir)

	newStorage := filesystem.NewStorage(fs, cache2.NewObjectLRUDefault())
	return newStorage

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

func GetMaster(repository *git.Repository) *config.Branch {

	if branch, err := repository.Branch("master"); err != nil {
		return nil
	} else {
		return branch
	}
}

func GetVersions(repository *git.Repository) []*plumbing.Reference {
	tags, err := repository.Tags()
	if err != nil {
		msg.Err("Fail to retrieve versions: %", err)
	}
	var result = make([]*plumbing.Reference, 0)

loop:
	reference, err := tags.Next()
	if err != nil {
		goto end
	}
	result = append(result, reference)
	goto loop
end:

	return result

}
