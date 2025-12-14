// Package gitadapter provides Git storage abstraction for caching repositories.
// This file creates filesystem-based storage for go-git operations.
package gitadapter

import (
	"path/filepath"

	"github.com/go-git/go-billy/v5/osfs"
	cache2 "github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/env"
)

// makeStorageCacheWithoutEnsure creates storage without ensuring cache dir exists.
// Used by GetRepository which is called after repo already exists.
func makeStorageCacheWithoutEnsure(dep domain.Dependency) storage.Storer {
	dir := filepath.Join(env.GetCacheDir(), dep.HashName())
	fs := osfs.New(dir)
	return filesystem.NewStorage(fs, cache2.NewObjectLRUDefault())
}
