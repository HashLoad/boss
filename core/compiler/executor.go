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

	readFileStr += " \n@SET DCC_UnitSearchPath=%DCC_UnitSearchPath%;" + getNewPathsDep(dep, abs) + " "
	for _, value := range []string{"Win32"} {
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

func getNewPathsDep(dep *models.Dependency, basePath string) string {
	if graphDep, err := loadOrderGraphDep(dep); err == nil {
		var result = ""
		for {
			if graphDep.IsEmpty() {
				break
			}
			dequeue := graphDep.Dequeue()
			var modulePath = filepath.Join(env.GetModulesDir(), dequeue.Dep.GetName())
			if depPkg, err := models.LoadPackageOther(filepath.Join(modulePath, consts.FilePackage)); err == nil {
				result += getPaths(filepath.Join(modulePath, depPkg.MainSrc), basePath)
			} else {
				result += getPaths(modulePath, basePath)
			}
		}
		return result
	} else {
		return getNewPathsAll(basePath)
	}
}

func getNewPathsAll(basePath string) string {
	return getPaths(env.GetModulesDir(), basePath)
}

func getPaths(path string, basePath string) string {
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

		matched, _ := regexp.MatchString(consts.RegexArtifacts, info.Name())
		dir := filepath.Dir(path)
		dir, err = filepath.Rel(basePath, dir)
		utils.HandleError(err)
		if matched && !utils.Contains(paths, dir) {
			paths = append(paths, dir)
		}
		return nil
	})
	return strings.Join(paths, ";") + ";"
}
