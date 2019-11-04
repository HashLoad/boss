package installer

import (
	"github.com/hashload/boss/core/gitWrapper"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"gopkg.in/src-d/go-git.v4"
	"os"
	"path/filepath"
)

var updatedDependencies []string

func GetDependency(dep models.Dependency) {
	if utils.Contains(updatedDependencies, dep.GetHashName()) {
		msg.Debug("Using cached of %s", dep.GetName())
		return
	} else {
		msg.Info("Updating cache of dependency %s", dep.GetName())
	}

	updatedDependencies = append(updatedDependencies, dep.GetHashName())
	var repository *git.Repository
	if hasCache(dep) {
		repository = gitWrapper.UpdateCache(dep)
	} else {
		_ = os.RemoveAll(filepath.Join(env.GetCacheDir(), dep.GetHashName()))
		repository = gitWrapper.CloneCache(dep)
	}
	tagsShortNames := gitWrapper.GetTagsShortName(repository)
	models.SaveRepoData(dep.GetHashName(), tagsShortNames)
}

func hasCache(dep models.Dependency) bool {
	dir := filepath.Join(env.GetCacheDir(), dep.GetHashName())
	info, err := os.Stat(dir)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	if !info.IsDir() {
		_ = os.RemoveAll(dir)
		return false
	}
	_, err = os.Stat(dir)
	return !os.IsNotExist(err)

}
