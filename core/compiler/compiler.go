package compiler

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"github.com/hashload/boss/utils/librarypath"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func isCommandAvailable(name string) bool {
	cmd := exec.Command(name, "-version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func getCompilerParameters(rootPath string, dep *models.Dependency) string {
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
		"/P:platform=Win32 "
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
	readFileStr += " \n msbuild \"" + project + "\" /p:Configuration=Debug " + getCompilerParameters(rootPath, dep)
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
		err = os.Remove(old)
		if !os.IsNotExist(err) {
			utils.HandleError(err)
		}
	}

}

func MoveArtifacts(dep *models.Dependency, rootPath string) {
	var moduleName = dep.GetName()
	movePath(filepath.Join(rootPath, moduleName, consts.BplFolder), filepath.Join(rootPath, consts.BplFolder))
	movePath(filepath.Join(rootPath, moduleName, consts.DcpFolder), filepath.Join(rootPath, consts.DcpFolder))
	movePath(filepath.Join(rootPath, moduleName, consts.BinFolder), filepath.Join(rootPath, consts.BinFolder))
	movePath(filepath.Join(rootPath, moduleName, consts.DcuFolder), filepath.Join(rootPath, consts.DcuFolder))

}

func EnsureArtifacts(lock *models.PackageLock, dep models.Dependency, rootPath string) {

	var moduleName = dep.GetName()

	lockedDependency := lock.GetInstalled(dep)

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

	lock.SetInstalled(dep, lockedDependency)
}

func compilePas(path string, additionalPaths string) {
	command := exec.Command("dcc32.exe", filepath.Base(path), additionalPaths)
	command.Dir = filepath.Dir(path)
	_ = command.Wait()
}

func Build(pkg *models.Package) {
	rootPath := env.GetCurrentDir()
	buildAllDprojByPackage(rootPath, pkg.Lock)
}

func buildAllDprojByPackage(rootPath string, lock models.PackageLock) {
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

			if !dependency.Changed {
				continue
			} else {
				dependency.Changed = false
				dprojs := modulePkg.Projects
				for _, dproj := range dprojs {
					s, _ := filepath.Abs(filepath.Join(env.GetModulesDir(), dep.GetName(), dproj))
					if !compile(s, env.GetModulesDir(), &dep) {
						dependency.Failed = true
					}
					EnsureArtifacts(&lock, dep, env.GetModulesDir())
					MoveArtifacts(&dep, env.GetModulesDir())
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
			MoveArtifacts(nil, rootPath)
			return nil
		})
}

func getNewPaths(path string, basePath string) string {
	var paths []string
	_ = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		matched, _ := regexp.MatchString(".*.pas$", info.Name())
		if !matched {
			matched, _ = regexp.MatchString(".*.inc$", info.Name())
		}
		matched = true
		dir := filepath.Dir(path)
		dir, err = filepath.Rel(basePath, dir)
		utils.HandleError(err)
		if matched && !librarypath.Contains(paths, dir) {
			paths = append(paths, dir)
		}
		return nil
	})
	return strings.Join(paths, ";")
}
