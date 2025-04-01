package installer

import (
	"os"
	"path/filepath"

	goGit "github.com/go-git/go-git/v5"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/git"
	"github.com/hashload/boss/pkg/models"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/utils"
)

//nolint:gochecknoglobals //TODO: Refactor this
var updatedDependencies []string

func GetDependency(dep models.Dependency) {
	if utils.Contains(updatedDependencies, dep.HashName()) {
		msg.Debug("Using cached of %s", dep.Name())
		return
	}
	msg.Info("Updating cache of dependency %s", dep.Name())

	updatedDependencies = append(updatedDependencies, dep.HashName())
	var repository *goGit.Repository
	if hasCache(dep) {
		repository = git.UpdateCache(dep)
	} else {
		_ = os.RemoveAll(filepath.Join(env.GetCacheDir(), dep.HashName()))
		repository = git.CloneCache(dep)
	}
	tagsShortNames := git.GetTagsShortName(repository)
	models.CacheRepositoryDetails(dep, tagsShortNames)
}

func hasCache(dep models.Dependency) bool {
	dir := filepath.Join(env.GetCacheDir(), dep.HashName())
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
