package core

import (
	"github.com/hashload/boss/core/cache"
	"github.com/hashload/boss/core/git"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	git2 "gopkg.in/src-d/go-git.v4"
	"time"
)

func GetDependency(dep models.Dependency) {
	if repoInfo, err := models.RepoData(dep.GetHashName()); err == nil {
		if time.Now().Before(repoInfo.LastUpdate.Add(1.8e+10)) {
			msg.Debug("Using cached of %s", dep.Repository)
			return
		}
	}
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
