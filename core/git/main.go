package git

import (
	"github.com/hashload/boss/core/paths"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/git/crazy"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage"
	"os"
	"path/filepath"
)

func CloneCache(dep models.Dependency) *git.Repository {
	msg.Info("Downloading dependency: " + dep.Repository)
	storer := makeStorageCache(dep)
	url := dep.GetURL()
	repository, e := git.Clone(storer, memfs.New(), &git.CloneOptions{
		URL:  url,
		Tags: git.AllTags,
		Auth: models.GlobalConfiguration.GetAuth(dep.GetURLPrefix()),
	})
	initSubmodules(dep, repository)
	if e != nil {
		msg.Die("Error to get repository of %s: %s", dep.Repository, e)
	}
	return repository
}

func UpdateCache(dep models.Dependency) *git.Repository {
	msg.Info("Updating dependency: " + dep.Repository)
	storer := makeStorageCache(dep)
	repository, e := git.Open(storer, memfs.New())
	if e != nil {
		msg.Warn("Error to open cache of %s: %s", dep.Repository, e)
		repository = refreshCopy(dep)
	} else {
		worktree, _ := repository.Worktree()
		worktree.Reset(&git.ResetOptions{
			Mode: git.HardReset,
		})
	}
	repository.Fetch(&git.FetchOptions{
		Force: true,
		Auth:  models.GlobalConfiguration.GetAuth(dep.GetURLPrefix())})
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
	if newStorage, e := crazy.NewStorage(fs); e != nil {
		msg.Die("Error to make filesystem for cache", e)
		return nil
	} else {
		return newStorage
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

func initSubmodules(dep models.Dependency, repository *git.Repository) {
	worktree, e := repository.Worktree()
	if e != nil {
		msg.Err("... %s", e)
	}
	submodules, e := worktree.Submodules()
	if e != nil {
		msg.Err("On get submodules... %s", e)
	}
	submodules.Update(&git.SubmoduleUpdateOptions{
		Init:              true,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Auth:              models.GlobalConfiguration.GetAuth(dep.GetURLPrefix()),
	})

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
