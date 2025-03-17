package git

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	cache2 "github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/models"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/pkg/paths"
)

func CloneCacheEmbedded(dep models.Dependency) *git.Repository {
	msg.Info("Downloading dependency %s", dep.Repository)
	storageCache := makeStorageCache(dep)
	wtFs := makeWtFileSystem(dep)
	url := dep.GetURL()
	auth := env.GlobalConfiguration().GetAuth(dep.GetURLPrefix())

	repository, e := git.Clone(storageCache, wtFs, &git.CloneOptions{
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
	wtFs := makeWtFileSystem(dep)

	repository, err := git.Open(storageCache, wtFs)
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
		Auth:  env.GlobalConfiguration().GetAuth(dep.GetURLPrefix())})
	if err != nil && err.Error() != "already up-to-date" {
		msg.Debug("Error to fetch repository of %s: %s", dep.Repository, err)
	}
	initSubmodules(dep, repository)
	return repository
}

func refreshCopy(dep models.Dependency) *git.Repository {
	dir := filepath.Join(env.GetCacheDir(), dep.GetHashName())
	err := os.RemoveAll(dir)
	if err == nil {
		return CloneCacheEmbedded(dep)
	}

	msg.Err("Error on retry get refresh copy: %s", err)

	return nil
}

func makeStorageCache(dep models.Dependency) storage.Storer {
	paths.EnsureCacheDir(dep)
	dir := filepath.Join(env.GetCacheDir(), dep.GetHashName())
	fs := osfs.New(dir)

	newStorage := filesystem.NewStorage(fs, cache2.NewObjectLRUDefault())
	return newStorage
}

func makeWtFileSystem(dep models.Dependency) billy.Filesystem {
	paths.EnsureCacheDir(dep)
	dir := filepath.Join(env.GetCacheDir(), fmt.Sprintf("%s_wt", dep.GetHashName()))
	fs := osfs.New(dir)

	return fs
}
