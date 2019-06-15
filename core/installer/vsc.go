package installer

import (
	"github.com/hashload/boss/core/cache"
	"github.com/hashload/boss/core/gitWrapper"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
)

var updatedDependencies []string

func GetDependency(dep models.Dependency) {
	if utils.Contains(updatedDependencies, dep.GetHashName()) {
		msg.Debug("Using cached of %s", dep.Repository)
		return
	}

	updatedDependencies = append(updatedDependencies, dep.GetHashName())

	if cache.HasCache(dep) {
		gitWrapper.UpdateCache(dep)
	} else {
		gitWrapper.CloneCache(dep)
	}
	models.SaveRepoData(dep.GetHashName())
}
