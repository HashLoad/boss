package compiler

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/utils"
	"io/ioutil"
	"os"
	"path/filepath"
)

func moveArtifacts(dep models.Dependency, rootPath string) {
	var moduleName = dep.GetName()
	movePath(filepath.Join(rootPath, moduleName, consts.BplFolder), filepath.Join(rootPath, consts.BplFolder))
	movePath(filepath.Join(rootPath, moduleName, consts.DcpFolder), filepath.Join(rootPath, consts.DcpFolder))
	movePath(filepath.Join(rootPath, moduleName, consts.BinFolder), filepath.Join(rootPath, consts.BinFolder))
	movePath(filepath.Join(rootPath, moduleName, consts.DcuFolder), filepath.Join(rootPath, consts.DcuFolder))
}

func movePath(old string, new string) {
	files, err := ioutil.ReadDir(old)
	var hasError = false
	if err == nil {
		for _, file := range files {
			if !file.IsDir() {
				err = os.Rename(filepath.Join(old, file.Name()), filepath.Join(new, file.Name()))
				if err != nil {
					hasError = true
				}
				utils.HandleError(err)
			}
		}
	}
	if !hasError {
		err = os.RemoveAll(old)
		if !os.IsNotExist(err) {
			utils.HandleError(err)
		}
	}

}

func ensureArtifacts(lockedDependency *models.LockedDependency, dep models.Dependency, rootPath string) {

	var moduleName = dep.GetName()
	lockedDependency.Artifacts.Clean()

	files, err := ioutil.ReadDir(filepath.Join(rootPath, moduleName, consts.BplFolder))
	if err == nil {
		for _, file := range files {
			if !file.IsDir() {
				lockedDependency.Artifacts.Bpl = append(lockedDependency.Artifacts.Bpl, file.Name())
			}
		}
	}

	files, err = ioutil.ReadDir(filepath.Join(rootPath, moduleName, consts.DcuFolder))
	if err == nil {
		for _, file := range files {
			if !file.IsDir() {
				lockedDependency.Artifacts.Dcu = append(lockedDependency.Artifacts.Dcu, file.Name())
			}
		}
	}

	files, err = ioutil.ReadDir(filepath.Join(rootPath, moduleName, consts.BinFolder))
	if err == nil {
		for _, file := range files {
			if !file.IsDir() {
				lockedDependency.Artifacts.Bin = append(lockedDependency.Artifacts.Bin, file.Name())
			}
		}
	}

	files, err = ioutil.ReadDir(filepath.Join(rootPath, moduleName, consts.DcpFolder))
	if err == nil {
		for _, file := range files {
			if !file.IsDir() {
				lockedDependency.Artifacts.Dcp = append(lockedDependency.Artifacts.Dcp, file.Name())
			}
		}
	}
}
