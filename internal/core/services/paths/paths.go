// Package paths provides utilities for managing file system paths used by Boss.
// It handles cache directory creation, module directory cleaning, and artifact management.
package paths

import (
	"os"
	"path/filepath"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/utils"
)

// EnsureCleanModulesDir ensures that the modules directory is clean and contains only the required dependencies.
//
//nolint:gocognit // Refactoring would reduce readability
func EnsureCleanModulesDir(dependencies []domain.Dependency, lock domain.PackageLock) {
	cacheDir := env.GetModulesDir()
	cacheDirInfo, err := os.Stat(cacheDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(cacheDir, 0755) // #nosec G301 -- Standard permissions for cache directory
		if err != nil {
			msg.Die("❌ Failed to create modules directory: %v", err)
		}
	}

	if cacheDirInfo != nil && !cacheDirInfo.IsDir() {
		msg.Die("❌ 'modules' is not a directory")
	}

	fileInfos, err := os.ReadDir(cacheDir)
	if err != nil {
		msg.Die("❌ Failed to read modules directory: %v", err)
	}
	dependenciesNames := domain.GetDependenciesNames(dependencies)
	for _, info := range fileInfos {
		if !info.IsDir() {
			err = os.Remove(info.Name())
			if err != nil {
				msg.Debug("Failed to remove file %s: %v", info.Name(), err)
			}
		}
		if utils.Contains(consts.DefaultPaths(), info.Name()) {
			cleanArtifacts(filepath.Join(cacheDir, info.Name()), lock)
			continue
		}

		if !utils.Contains(dependenciesNames, info.Name()) {
		remove:
			if err = os.RemoveAll(filepath.Join(cacheDir, info.Name())); err != nil {
				msg.Warn("⚠️ Failed to remove old cache: %s", err.Error())
				goto remove
			}
		}
	}

	for _, path := range consts.DefaultPaths() {
		createPath(filepath.Join(cacheDir, path))
	}
}

// EnsureCacheDir ensures that the cache directory exists for the dependency.
func EnsureCacheDir(config env.ConfigProvider, dep domain.Dependency) {
	if !config.GetGitEmbedded() {
		return
	}
	cacheDir := filepath.Join(env.GetCacheDir(), dep.HashName())

	fi, err := os.Stat(cacheDir)
	if err != nil {
		msg.Debug("Creating %s", cacheDir)
		err = os.MkdirAll(cacheDir, 0755) // #nosec G301 -- Standard permissions for cache directory
		if err != nil {
			msg.Die("❌ Could not create %s: %s", cacheDir, err)
		}
	} else if !fi.IsDir() {
		msg.Die("❌ 'cache' is not a directory")
	}
}

func createPath(path string) {
	if err := os.MkdirAll(path, os.ModeDir|0755); err != nil {
		msg.Die("❌ Failed to create path %s: %v", path, err)
	}
}

func cleanArtifacts(dir string, lock domain.PackageLock) {
	fileInfos, err := os.ReadDir(dir)
	if err != nil {
		msg.Warn("⚠️ Failed to read artifacts directory: %v", err)
		return
	}
	artifactList := lock.GetArtifactList()
	for _, infoArtifact := range fileInfos {
		if infoArtifact.IsDir() {
			continue
		}
		if !utils.Contains(artifactList, infoArtifact.Name()) {
			err = os.Remove(filepath.Join(dir, infoArtifact.Name()))
			if err != nil {
				msg.Debug("Failed to remove artifact %s: %v", infoArtifact.Name(), err)
			}
		}
	}
}
