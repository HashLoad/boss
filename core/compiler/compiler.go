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

func getDcc32Dir() string {
	if env.GlobalConfiguration.DelphiPath != "" {
		return env.GlobalConfiguration.DelphiPath
	}

	command := exec.Command("where", "dcc32")
	output, err := command.Output()
	if err != nil {
		msg.Warn("dcc32 not found")
	}
	outputStr := strings.ReplaceAll(string(output), "\n", "")
	outputStr = strings.ReplaceAll(outputStr, "\r", "")
	outputStr = filepath.Dir(outputStr)

	return outputStr
}

func getCompilerParameters(rootPath string) string {
	var binPath string
	if !env.Global {
		binPath = rootPath + consts.Separator + consts.BinFolder
	} else {
		binPath = env.GetGlobalBinPath()
	}

	return " /p:DCC_BplOutput=\"" + rootPath + consts.Separator + ".bpl\" " +
		"/p:DCC_DcpOutput=\"" + rootPath + consts.Separator + ".dcp\" " +
		"/p:DCC_DcuOutput=\"" + rootPath + consts.Separator + ".dcu\" " +
		"/p:DCC_ExeOutput=\"" + binPath + "\" " +
		"/target:Build " +
		"/p:config=Debug " +
		"/P:platform=Win32 "
}

//noinspection GoUnhandledErrorResult
func compile(path string, rootPath string) {
	msg.Info("  Building " + filepath.Base(path))
	dccDir := getDcc32Dir()
	rsvars := dccDir + consts.Separator + "rsvars.bat"
	fileRes := "build_boss_" + strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	abs, _ := filepath.Abs(filepath.Dir(path))
	buildLog := abs + consts.Separator + fileRes + ".log"
	buildBat := abs + consts.Separator + fileRes + ".bat"
	readFile, err := ioutil.ReadFile(rsvars)
	if err != nil {
		msg.Err("    error on read rsvars.bat")
	}
	readFileStr := string(readFile)
	project, _ := filepath.Abs(path)

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

//noinspection GoUnhandledErrorResult
func compilePas(path string, additionalPaths string) {
	command := exec.Command("dcc32.exe", filepath.Base(path), additionalPaths)
	command.Dir = filepath.Dir(path)
	command.Output()

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
			modulePkg, err := models.LoadPackageOther(rootPath + consts.Separator + dep.GetName() + consts.Separator + consts.FilePackage)
			if err != nil {
				return
			}

			dprojs := modulePkg.Projects
			for _, dproj := range dprojs {
				s, _ := filepath.Abs(rootPath + consts.Separator + dep.GetName() + "/" + dproj)
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
