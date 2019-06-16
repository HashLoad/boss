package compiler

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func Build(pkg *models.Package) {
	//dependencyOrder(pkg)
	rootPath := env.GetCurrentDir()
	buildAllDprojByPackage(rootPath, &pkg.Lock)
}

func getCompilerParameters(rootPath string, dep *models.Dependency, platform string) string {
	var binPath string
	var moduleName = ""

	if dep != nil {
		moduleName = dep.GetName()
	}

	if !env.Global {
		binPath = filepath.Join(rootPath, moduleName, consts.BinFolder)
	} else {
		binPath = env.GetGlobalBinPath()
	}

	return " /p:DCC_BplOutput=\"" + filepath.Join(rootPath, moduleName, consts.BplFolder) + "\" " +
		"/p:DCC_DcpOutput=\"" + filepath.Join(rootPath, moduleName, consts.DcpFolder) + "\" " +
		"/p:DCC_DcuOutput=\"" + filepath.Join(rootPath, moduleName, consts.DcuFolder) + "\" " +
		"/p:DCC_ExeOutput=\"" + binPath + "\" " +
		"/target:Build " +
		"/p:config=Debug " +
		"/P:platform=" + platform + " "
}

func compile(dprojPath string, rootPath string, dep *models.Dependency) bool {
	msg.Info("  Building " + filepath.Base(dprojPath))
	dccDir := env.GetDcc32Dir()
	rsvars := filepath.Join(dccDir, "rsvars.bat")
	fileRes := "build_boss_" + strings.TrimSuffix(filepath.Base(dprojPath), filepath.Ext(dprojPath))
	abs, _ := filepath.Abs(filepath.Dir(dprojPath))
	buildLog := filepath.Join(abs, fileRes+".log")
	buildBat := filepath.Join(abs, fileRes+".bat")
	readFile, err := ioutil.ReadFile(rsvars)
	if err != nil {
		msg.Err("    error on read rsvars.bat")
	}
	readFileStr := string(readFile)
	project, _ := filepath.Abs(dprojPath)

	readFileStr += " \n@SET DCC_UnitSearchPath=%DCC_UnitSearchPath%;" + getNewPaths(env.GetModulesDir(), abs) + " "
	for _, value := range []string{"Win32"} {
		//librarypath.GetActivePlatforms(dprojPath) {
		readFileStr += " \n msbuild \"" + project + "\" /p:Configuration=Debug " + getCompilerParameters(rootPath, dep, value)
	}
	readFileStr += " > \"" + buildLog + "\""

	err = ioutil.WriteFile(buildBat, []byte(readFileStr), os.ModePerm)
	if err != nil {
		msg.Warn("  - error on create build file")
		return false
	}

	command := exec.Command(buildBat)
	command.Dir = abs
	if _, err := command.Output(); err != nil {
		msg.Err("  - Failed to compile, see " + buildLog + " for more information")
		return false
	} else {
		msg.Info("  - Success!")
		err := os.Remove(buildLog)
		utils.HandleError(err)
		err = os.Remove(buildBat)
		utils.HandleError(err)
		return true
	}
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

func moveArtifacts(dep *models.Dependency, rootPath string) {
	var moduleName = dep.GetName()
	movePath(filepath.Join(rootPath, moduleName, consts.BplFolder), filepath.Join(rootPath, consts.BplFolder))
	movePath(filepath.Join(rootPath, moduleName, consts.DcpFolder), filepath.Join(rootPath, consts.DcpFolder))
	movePath(filepath.Join(rootPath, moduleName, consts.BinFolder), filepath.Join(rootPath, consts.BinFolder))
	movePath(filepath.Join(rootPath, moduleName, consts.DcuFolder), filepath.Join(rootPath, consts.DcuFolder))

}

func ensureArtifacts(lockedDependency *models.LockedDependency, dep models.Dependency, rootPath string) {

	var moduleName = dep.GetName()

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

func buildAllDprojByPackage(rootPath string, lock *models.PackageLock) {
	if pkg, err := models.LoadPackageOther(filepath.Join(rootPath, consts.FilePackage)); err != nil || pkg.Dependencies == nil {
		buildAllDproj(rootPath)
	} else {
		rawDeps := pkg.Dependencies.(map[string]interface{})

		deps := models.GetDependencies(rawDeps)
		for _, dep := range deps {
			modulePkg, err := models.LoadPackageOther(filepath.Join(env.GetModulesDir(), dep.GetName(), consts.FilePackage))
			if err != nil {
				continue
			}

			buildAllDprojByPackage(filepath.Join(env.GetModulesDir(), dep.GetName()), lock)

			dependency := lock.GetInstalled(dep)

			if !dependency.Changed && !(len(dependency.GetArtifacts()) == 0 && len(modulePkg.Projects) > 0) {
				continue
			} else {
				dependency.Changed = false
				dprojs := modulePkg.Projects
				for _, dproj := range dprojs {
					s, _ := filepath.Abs(filepath.Join(env.GetModulesDir(), dep.GetName(), dproj))
					if !compile(s, env.GetModulesDir(), &dep) {
						dependency.Failed = true
					}
					ensureArtifacts(&dependency, dep, env.GetModulesDir())
					moveArtifacts(&dep, env.GetModulesDir())
				}
				lock.SetInstalled(dep, dependency)
			}
		}
	}
}

func buildAllDproj(rootPath string) {
	_ = filepath.Walk(rootPath,
		func(path string, info os.FileInfo, err error) error {
			if os.IsNotExist(err) {
				return nil
			} else if err != nil {
				msg.Err(err.Error())
			}

			if info.IsDir() {
				return nil
			}

			if filepath.Ext(path) != ".dproj" {
				return nil
			}
			compile(path, rootPath, nil)
			moveArtifacts(nil, rootPath)
			return nil
		})
}

func getNewPaths(path string, basePath string) string {
	var ignore = []string{consts.BplFolder, consts.BinFolder, consts.DcpFolder, consts.DcuFolder}
	var paths []string
	_ = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}
		if info.IsDir() {
			return nil
		} else {
			if utils.Contains(ignore, info.Name()) {
				return nil
			}
		}

		matched, _ := regexp.MatchString(consts.REGEX_ARTIFACTS, info.Name())
		dir := filepath.Dir(path)
		dir, err = filepath.Rel(basePath, dir)
		utils.HandleError(err)
		if matched && !utils.Contains(paths, dir) {
			paths = append(paths, dir)
		}
		return nil
	})

	return strings.Join(paths, ";")
}
