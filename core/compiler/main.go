package compiler

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
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
	return " /p:DCC_BplOutput=\"" + rootPath + consts.SEPARATOR + ".bpl\" " +
		"/p:DCC_DcpOutput=\"" + rootPath + consts.SEPARATOR + ".dcp\" " +
		"/p:DCC_DcuOutput=\"" + rootPath + consts.SEPARATOR + ".dcu\" " +
		"/target:Build " +
		"/p:config=Debug " +
		"/P:platform=Win32 "
}

//noinspection GoUnhandledErrorResult
func compile(path string, rootPath string) {
	msg.Info("  Building " + filepath.Base(path))
	dccDir := getDcc32Dir()
	rsvars := dccDir + consts.SEPARATOR + "rsvars.bat"
	fileRes := "build_boss_" + strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	abs, _ := filepath.Abs(filepath.Dir(path))
	buildLog := abs + consts.SEPARATOR + fileRes + ".log"
	buildBat := abs + consts.SEPARATOR + fileRes + ".bat"
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
	curr, _ := os.Getwd()
	rootPath := curr + consts.SEPARATOR + "modules"
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
			modulePkg, err := models.LoadPackageOther(rootPath + consts.SEPARATOR + dep.GetName() + consts.SEPARATOR + consts.FILE_PACKAGE)
			if err != nil {
				return
			}

			if len(modulePkg.Projects) == 0 {
				msg.Warn("Dependency " + dep.GetName() + " has no seted dproj for compile, assuming compilation for all dproj")
				buildAllDproj(rootPath + consts.SEPARATOR + dep.GetName())
			} else {
				dprojs := modulePkg.Projects
				for _, dproj := range dprojs {
					s, _ := filepath.Abs(rootPath + consts.SEPARATOR + dep.GetName() + "/" + dproj)
					compile(s, rootPath)
				}
			}
		}
	}

}

func buildAllPas() {
	paths := getNewPaths(filepath.Dir("./modules"))
	additionalPaths := "-U" + strconv.Quote(paths)

	_ = filepath.Walk("./modules",
		func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
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
		matched, _ := regexp.MatchString(".*.pas$", info.Name())
		dir := filepath.Dir(path)
		dir, _ = filepath.Abs(dir)
		if matched && !utils.Contains(paths, dir) {
			paths = append(paths, dir)
		}
		return nil
	})
	return strings.Join(paths, ";")
}
