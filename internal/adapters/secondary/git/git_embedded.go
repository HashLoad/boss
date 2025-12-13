package gitadapter

import (
	"os"
	"path/filepath"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	cache2 "github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/paths"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
)

func CloneCacheEmbedded(dep domain.Dependency) (*git.Repository, error) {
	msg.Info("Downloading dependency %s", dep.Repository)
	storageCache := makeStorageCache(dep)
	worktreeFileSystem := createWorktreeFs(dep)
	url := dep.GetURL()
	auth := env.GlobalConfiguration().GetAuth(dep.GetURLPrefix())

	repository, err := git.Clone(storageCache, worktreeFileSystem, &git.CloneOptions{
		URL:  url,
		Tags: git.AllTags,
		Auth: auth,
	})
	if err != nil {
		_ = os.RemoveAll(filepath.Join(env.GetCacheDir(), dep.HashName()))
		return nil, err
	}
	if err := initSubmodules(dep, repository); err != nil {
		return nil, err
	}
	return repository, nil
}

func UpdateCacheEmbedded(dep domain.Dependency) (*git.Repository, error) {
	storageCache := makeStorageCache(dep)
	wtFs := createWorktreeFs(dep)

	repository, err := git.Open(storageCache, wtFs)
	if err != nil {
		msg.Warn("Error to open cache of %s: %s", dep.Repository, err)
		var errRefresh error
		repository, errRefresh = refreshCopy(dep)
		if errRefresh != nil {
			return nil, errRefresh
		}
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
	if err := initSubmodules(dep, repository); err != nil {
		return nil, err
	}
	return repository, nil
}

func refreshCopy(dep domain.Dependency) (*git.Repository, error) {
	dir := filepath.Join(env.GetCacheDir(), dep.HashName())
	err := os.RemoveAll(dir)
	if err == nil {
		return CloneCacheEmbedded(dep)
	}

	msg.Err("Error on retry get refresh copy: %s", err)

	return nil, err
}

func makeStorageCache(dep domain.Dependency) storage.Storer {
	paths.EnsureCacheDir(dep)
	dir := filepath.Join(env.GetCacheDir(), dep.HashName())
	fs := osfs.New(dir)

	newStorage := filesystem.NewStorage(fs, cache2.NewObjectLRUDefault())
	return newStorage
}

func createWorktreeFs(dep domain.Dependency) billy.Filesystem {
	paths.EnsureCacheDir(dep)
	fs := memfs.New()

	return fs
}

func CheckoutEmbedded(dep domain.Dependency, referenceName plumbing.ReferenceName) error {
	repository := GetRepository(dep)
	worktree, err := repository.Worktree()
	if err != nil {
		return err
	}
	return worktree.Checkout(&git.CheckoutOptions{
		Force:  true,
		Branch: referenceName,
	})
}

func PullEmbedded(dep domain.Dependency) error {
	repository := GetRepository(dep)
	worktree, err := repository.Worktree()
	if err != nil {
		return err
	}
	return worktree.Pull(&git.PullOptions{
		Force: true,
		Auth:  env.GlobalConfiguration().GetAuth(dep.GetURLPrefix()),
	})
}
