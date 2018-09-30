package core

import (
	"github.com/hashload/boss/core/cache"
	"github.com/hashload/boss/core/git"
	"github.com/hashload/boss/models"
	git2 "gopkg.in/src-d/go-git.v4"
)

func GetDependency(dep models.Dependency) {
	if cache.HasCache(dep) {
		git.UpdateCache(dep)
	} else {
		git.CloneCache(dep)
	}
}

func OpenRepository() *git2.Repository {
	git.CloneCache()
}
