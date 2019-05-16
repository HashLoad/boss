package core

import (
	"github.com/hashload/boss/core/cache"
	"github.com/hashload/boss/core/git"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	git2 "gopkg.in/src-d/go-git.v4"
)

var updatedDependencies []string

func GetDependency(dep models.Dependency) {
	if utils.Contains(updatedDependencies, dep.GetHashName()) {
		msg.Debug("Using cached of %s", dep.Repository)
		return
	}

	updatedDependencies = append(updatedDependencies, dep.GetHashName())

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
