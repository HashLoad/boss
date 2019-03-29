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
	models.SaveRepoData(dep.GetHashName())
}

func OpenRepository(dep models.Dependency) *git2.Repository {
	return git.GetRepository(dep)
}

//func EnsureVersionModule(repository *git2.Repository, dependency models.Dependency) *git2.Repository {
//	gotoMaxVersion(repository, dependency)
//
//	return repository
//}
//
//func gotoMaxVersion(repository *git2.Repository, dependency models.Dependency) {
//}
