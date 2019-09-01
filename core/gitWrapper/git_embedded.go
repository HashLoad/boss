package gitWrapper

import (
	"github.com/hashload/boss/core/paths"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4"
	cache2 "gopkg.in/src-d/go-git.v4/plumbing/cache"
	"gopkg.in/src-d/go-git.v4/storage"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
	"os"
	"path/filepath"
)

func CloneCacheEmbedded(dep models.Dependency) *git.Repository {
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

func UpdateCacheEmbedded(dep models.Dependency) *git.Repository {
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

	err = repository.Fetch(&git.FetchOptions{
		Force: true,
		Auth:  env.GlobalConfiguration.GetAuth(dep.GetURLPrefix())})
	if err != nil && err.Error() != "already up-to-date" {
		msg.Debug("Error to fetch repository of %s: %s", dep.Repository, err)
	}
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
