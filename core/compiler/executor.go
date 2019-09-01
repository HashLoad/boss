package compiler

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"github.com/hashload/boss/utils/dcp"
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

func compile(dprojPath string, dep *models.Dependency, rootLock models.PackageLock) bool {
	msg.Info("  Building " + filepath.Base(dprojPath))

	bossPackagePath := filepath.Join(env.GetModulesDir(), dep.GetName(), consts.FilePackage)

	if dependencyPackage, err := models.LoadPackageOther(bossPackagePath); err == nil {
		dcp.InjectDpcsFile(dprojPath, dependencyPackage, rootLock)
	}

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

	readFileStr += "\n@SET DCC_UnitSearchPath=%DCC_UnitSearchPath%;" + filepath.Join(env.GetModulesDir(), consts.DcuFolder) +
		";" + filepath.Join(env.GetModulesDir(), consts.DcpFolder) //+ ";" + getNewPathsDep(dep, abs) + " "

	readFileStr += "\n@SET PATH=%PATH%;" + filepath.Join(env.GetModulesDir(), consts.BplFolder) + ";"
	for _, value := range []string{"Win32"} {
		readFileStr += " \n msbuild \"" + project + "\" /p:Configuration=Debug " + getCompilerParameters(env.GetModulesDir(), dep, value)
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

func _(dep *models.Dependency, basePath string) string {
	if graphDep, err := loadOrderGraphDep(dep); err == nil {
		var result = filepath.Join(env.GetModulesDir(), consts.DcpFolder) + ";"
		for {
			if graphDep.IsEmpty() {
				break
			}
			dequeue := graphDep.Dequeue()
			var modulePath = filepath.Join(env.GetModulesDir(), dequeue.Dep.GetName())
			if modulePath == basePath {
				continue
			}
			if depPkg, err := models.LoadPackageOther(filepath.Join(modulePath, consts.FilePackage)); err == nil {
				result += getPaths(filepath.Join(modulePath, depPkg.MainSrc), basePath)
			} else {
				result += getPaths(modulePath, basePath)
			}
		}
		return result
	} else {
		return getNewPathsAll(basePath) + ";" + filepath.Join(env.GetModulesDir(), consts.DcpFolder)
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

func buildDCU(path string) {
	msg.Info("  Building %s", filepath.Base(path))
	var unitScopes = "-NSWinapi;System.Win;Data.Win;Datasnap.Win;Web.Win;Soap.Win;Xml.Win;Bde;System;Xml;Data;Datasnap;Web" +
		";Soap;Vcl;Vcl.Imaging;Vcl.Touch;Vcl.Samples;Vcl.Shell"
	var unitInputDir = "-U" + filepath.Join(env.GetModulesDir(), consts.DcuFolder)
	var unitOutputDir = "-NU" + filepath.Join(env.GetModulesDir(), consts.DcuFolder)
	command := exec.Command("cmd", "/c dcc32.exe "+unitScopes+" "+unitInputDir+" "+unitOutputDir+" "+path)
	command.Dir = filepath.Dir(path)
	if out, err := command.Output(); err != nil {
		msg.Err("  - Failed to compile")
		msg.Err(string(out))
	} else {
		msg.Info("  - Success!")
	}
}
