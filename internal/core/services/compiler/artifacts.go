package compiler

import (
	"os"
	"path/filepath"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/consts"
)

func moveArtifacts(dep domain.Dependency, rootPath string) {
	var moduleName = dep.Name()
	movePath(filepath.Join(rootPath, moduleName, consts.BplFolder), filepath.Join(rootPath, consts.BplFolder))
	movePath(filepath.Join(rootPath, moduleName, consts.DcpFolder), filepath.Join(rootPath, consts.DcpFolder))
	movePath(filepath.Join(rootPath, moduleName, consts.BinFolder), filepath.Join(rootPath, consts.BinFolder))
	movePath(filepath.Join(rootPath, moduleName, consts.DcuFolder), filepath.Join(rootPath, consts.DcuFolder))
}

func movePath(oldPath string, newPath string) {
	files, err := os.ReadDir(oldPath)
	var hasError = false
	if err == nil {
		for _, file := range files {
			if !file.IsDir() {
				err = os.Rename(filepath.Join(oldPath, file.Name()), filepath.Join(newPath, file.Name()))
				if err != nil {
					hasError = true
				}
			}
		}
	}
	if !hasError {
		err = os.RemoveAll(oldPath)
		if err != nil && !os.IsNotExist(err) {
			// Non-critical: artifact cleanup failed
		}
	}
}

func ensureArtifacts(lockedDependency *domain.LockedDependency, dep domain.Dependency, rootPath string) {
	var moduleName = dep.Name()
	lockedDependency.Artifacts.Clean()

	collectArtifacts(lockedDependency.Artifacts.Bpl, filepath.Join(rootPath, moduleName, consts.BplFolder))
	collectArtifacts(lockedDependency.Artifacts.Dcu, filepath.Join(rootPath, moduleName, consts.DcuFolder))
	collectArtifacts(lockedDependency.Artifacts.Bin, filepath.Join(rootPath, moduleName, consts.BinFolder))
	collectArtifacts(lockedDependency.Artifacts.Dcp, filepath.Join(rootPath, moduleName, consts.DcpFolder))
}

func collectArtifacts(artifactList []string, path string) {
	files, err := os.ReadDir(path)
	if err == nil {
		for _, file := range files {
			if !file.IsDir() {
				artifactList = append(artifactList, file.Name())
			}
		}
	}
}
