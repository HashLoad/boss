package compiler

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils/librarypath"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func isCommandAvailable(name string) bool {
	cmd := exec.Command(name, "-version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func getCompilerParameters(rootPath string) string {
	var binPath string
	if !env.Global {
		binPath = filepath.Join(rootPath, consts.BinFolder)
	} else {
		binPath = env.GetGlobalBinPath()
	}

	return " /p:DCC_BplOutput=\"" + filepath.Join(rootPath, consts.BplFolder) + "\" " +
		"/p:DCC_DcpOutput=\"" + filepath.Join(rootPath, consts.DcpFolder) + "\" " +
		"/p:DCC_DcuOutput=\"" + filepath.Join(rootPath, consts.DcuFolder) + "\" " +
		"/p:DCC_ExeOutput=\"" + binPath + "\" " +
		"/target:Rebuild " +
		"/p:config=Debug " +
		"/P:platform=Win32 "
}

func compile(dprojPath string, rootPath string) {
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

	readFileStr += " \n@SET DCC_UnitSearchPath=%DCC_UnitSearchPath%;" + getNewPaths(env.GetModulesDir()) + " "
	readFileStr += " \n msbuild " + project + " /t:Build /p:Configuration=Debug " + getCompilerParameters(rootPath)
	readFileStr += " > " + buildLog

	err = ioutil.WriteFile(buildBat, []byte(readFileStr), os.ModePerm)
	if err != nil {
		msg.Warn("  - error on create build file")
		return
	}

	command := exec.Command(buildBat)
	command.Dir = abs
	if _, err := command.Output(); err != nil {
		msg.Err("  - Falied to compile, see " + buildLog + " for more information")
	} else {
		msg.Info("  - Success!")
	}

}

func compilePas(path string, additionalPaths string) {
	command := exec.Command("dcc32.exe", filepath.Base(path), additionalPaths)
	command.Dir = filepath.Dir(path)
	_ = command.Wait()
}

func BuildDucs() {
	rootPath := env.GetModulesDir()
	if !isCommandAvailable("dcc32.exe") {
		msg.Warn("dcc32 not found in path")
		return
	}

	buildAllPas()

	if pkg, err := models.LoadPackage(false); err != nil || pkg.Dependencies == nil {
		buildAllDproj(rootPath)
	} else {
		rawDeps := pkg.Dependencies.(map[string]interface{})
		deps := models.GetDependencies(rawDeps)
		for _, dep := range deps {
			modulePkg, err := models.LoadPackageOther(filepath.Join(rootPath, dep.GetName(), consts.FilePackage))
			if err != nil {
				continue
			}

			dprojs := modulePkg.Projects
			for _, dproj := range dprojs {
				s, _ := filepath.Abs(filepath.Join(rootPath, dep.GetName(), dproj))
				compile(s, rootPath)
			}
		}
	}

}

func buildAllPas() {
	paths := getNewPaths(env.GetModulesDir())
	additionalPaths := "-U" + strconv.Quote(paths)

	_ = filepath.Walk(env.GetModulesDir(),
		func(path string, info os.FileInfo, err error) error {
			if info != nil && info.IsDir() {
				return nil
			}

			if filepath.Ext(path) != ".pas" {
				return nil
			}

			compilePas(path, additionalPaths)
			return nil
		})
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
			compile(path, rootPath)
			return nil
		})
}

func getNewPaths(path string) string {
	var paths []string
	_ = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}
		matched, _ := regexp.MatchString(".*.pas$", info.Name())
		dir := filepath.Dir(path)
		dir, _ = filepath.Abs(dir)
		if matched && !librarypath.Contains(paths, dir) {
			paths = append(paths, dir)
		}
		return nil
	})
	return strings.Join(paths, ";")
}
