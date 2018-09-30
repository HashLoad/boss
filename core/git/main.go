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
	"gopkg.in/src-d/go-git.v4/storage"
	"os"
	"path/filepath"
)

func CloneCache(dep models.Dependency) *git.Repository {
	msg.Default.Info("Downloading dependency: " + dep.Repository)
	storer := makeStorageCache(dep)
	repository, e := git.Clone(storer, memfs.New(), &git.CloneOptions{
		URL:  dep.GetURL(),
		Tags: git.AllTags,
	})
	if e != nil {
		msg.Default.Die("Error to get repository of %s: %s", dep.Repository, e)
	}
	return repository
}

func UpdateCache(dep models.Dependency) *git.Repository {
	msg.Default.Info("Updating dependency: " + dep.Repository)
	storer := makeStorageCache(dep)
	repository, e := git.Open(storer, memfs.New())
	if e != nil {
		msg.Default.Warn("Error to open cache of %s: %s", dep.Repository, e)
		repository = refreshCopy(dep)
	}
	repository.Fetch(&git.FetchOptions{Force: true,})
	return repository
}

func refreshCopy(dep models.Dependency) *git.Repository {
	dir := filepath.Join(env.GetCacheDir(), dep.GetHashName());
	e := os.RemoveAll(dir)
	if (e == nil) {
		return CloneCache(dep)
	} else {
		msg.Default.Err("Error on retry get refresh copy: %s", e)
	}
	return nil
}

func makeStorageCache(dep models.Dependency) (storage.Storer) {
	paths.EnsureCacheDir(dep)
	dir := filepath.Join(env.GetCacheDir(), dep.GetHashName());
	fs := osfs.New(dir)
	if newStorage, e := crazy.NewStorage(fs); e != nil {
		msg.Default.Die("Error to make filesystem for cache", e)
		return nil
	} else {
		return newStorage
	}
}

func GetRepository(dep models.Dependency) *git.Repository {
	cache := makeStorageCache(dep)
	mfs := memfs.New()
	repository, e := git.Open(cache, mfs)
	if e != nil {
		msg.Err("Error on open repository %s: %s", dep.Repository, e)
	}

	worktree, _ := repository.Worktree()

	return repository

}
