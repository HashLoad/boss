package paths

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"io/ioutil"
	"os"
	"path/filepath"
)

func EnsureCleanModulesDir(dependencies []models.Dependency, lock models.PackageLock) {
	cacheDir := env.GetModulesDir()
	cacheDirInfo, err := os.Stat(cacheDir)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(cacheDir, os.ModeDir|0755); err != nil {
			msg.Die("Could not create %s: %s", cacheDir, err)
		}
	} else if cacheDirInfo != nil && !cacheDirInfo.IsDir() {
		msg.Die("modules is not a directory")
	} else {
		fileInfos, err := ioutil.ReadDir(cacheDir)
		utils.HandleError(err)
		dependenciesNames := models.GetDependenciesNames(dependencies)
		for _, info := range fileInfos {
			if !info.IsDir() {
				err := os.Remove(info.Name())
				utils.HandleError(err)
			}
			if utils.Contains(consts.DefaultPaths, info.Name()) {
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
	}
	createPath(filepath.Join(cacheDir, consts.BplFolder))
	createPath(filepath.Join(cacheDir, consts.DcuFolder))
	createPath(filepath.Join(cacheDir, consts.DcpFolder))
	createPath(filepath.Join(cacheDir, consts.BinFolder))
}

func EnsureCacheDir(dep models.Dependency) {
	if !env.GlobalConfiguration.GitEmbedded {
		return
	}
	cacheDir := filepath.Join(env.GetCacheDir(), dep.GetHashName())

	fi, err := os.Stat(cacheDir)
	if err != nil {
		msg.Debug("Creating %s", cacheDir)
		if err := os.MkdirAll(cacheDir, os.ModeDir|0755); err != nil {
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
	fileInfos, err := ioutil.ReadDir(dir)
	utils.HandleError(err)
	artifactList := lock.GetArtifactList()
	for _, infoArtifact := range fileInfos {
		if infoArtifact.IsDir() {
			continue
		}
		if !utils.Contains(artifactList, infoArtifact.Name()) {
			for {
				err := os.Remove(filepath.Join(dir, infoArtifact.Name()))
				utils.HandleError(err)
				if err == nil {
					break
				}
			}
		}

	}
}
