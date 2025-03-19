package paths

import (
	"os"
	"path/filepath"

	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/models"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/utils"
)

func EnsureCleanModulesDir(dependencies []models.Dependency, lock models.PackageLock) {
	cacheDir := env.GetModulesDir()
	cacheDirInfo, err := os.Stat(cacheDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(cacheDir, os.ModeDir|0755)
		utils.HandleError(err)
	}

	if cacheDirInfo != nil && !cacheDirInfo.IsDir() {
		msg.Die("modules is not a directory")
	}

	fileInfos, err := os.ReadDir(cacheDir)
	utils.HandleError(err)
	dependenciesNames := models.GetDependenciesNames(dependencies)
	for _, info := range fileInfos {
		if !info.IsDir() {
			err = os.Remove(info.Name())
			utils.HandleError(err)
		}
		if utils.Contains(consts.DefaultPaths(), info.Name()) {
			cleanArtifacts(filepath.Join(cacheDir, info.Name()), lock)
			continue
		}

		if !utils.Contains(dependenciesNames, info.Name()) {
		remove:
			if err = os.RemoveAll(filepath.Join(cacheDir, info.Name())); err != nil {
				msg.Warn("Failed to remove old cache: %s", err.Error())
				goto remove
			}
		}
	}

	for _, path := range consts.DefaultPaths() {
		createPath(filepath.Join(cacheDir, path))
	}
}

func EnsureCacheDir(dep models.Dependency) {
	if !env.GlobalConfiguration().GitEmbedded {
		return
	}
	cacheDir := filepath.Join(env.GetCacheDir(), dep.HashName())

	fi, err := os.Stat(cacheDir)
	if err != nil {
		msg.Debug("Creating %s", cacheDir)
		err = os.MkdirAll(cacheDir, os.ModeDir|0755)
		if err != nil {
			msg.Die("Could not create %s: %s", cacheDir, err)
		}
	} else if !fi.IsDir() {
		msg.Die("cache is not a directory")
	}
}

func createPath(path string) {
	utils.HandleError(os.MkdirAll(path, os.ModeDir|0755))
}

func cleanArtifacts(dir string, lock models.PackageLock) {
	fileInfos, err := os.ReadDir(dir)
	utils.HandleError(err)
	artifactList := lock.GetArtifactList()
	for _, infoArtifact := range fileInfos {
		if infoArtifact.IsDir() {
			continue
		}
		if !utils.Contains(artifactList, infoArtifact.Name()) {
			for {
				err = os.Remove(filepath.Join(dir, infoArtifact.Name()))
				utils.HandleError(err)
				if err == nil {
					break
				}
			}
		}
	}
}
