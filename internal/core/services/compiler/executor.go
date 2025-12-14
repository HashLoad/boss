package compiler

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/compiler_selector"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/utils"
	"github.com/hashload/boss/utils/dcp"
)

func getCompilerParameters(rootPath string, dep *domain.Dependency, platform string) string {
	var moduleName = ""

	if dep != nil {
		moduleName = dep.Name()
	}

	binPath := env.GetGlobalBinPath()

	if !env.GetGlobal() {
		binPath = filepath.Join(rootPath, moduleName, consts.BinFolder)
	}

	return " /p:DCC_BplOutput=\"" + filepath.Join(rootPath, moduleName, consts.BplFolder) + "\" " +
		"/p:DCC_DcpOutput=\"" + filepath.Join(rootPath, moduleName, consts.DcpFolder) + "\" " +
		"/p:DCC_DcuOutput=\"" + filepath.Join(rootPath, moduleName, consts.DcuFolder) + "\" " +
		"/p:DCC_ExeOutput=\"" + binPath + "\" " +
		"/target:Build " +
		"/p:config=Debug " +
		"/p:DCC_UseMSBuildExternally=true " +
		"/P:platform=" + platform + " "
}

func buildSearchPath(dep *domain.Dependency) string {
	var searchPath = ""

	if dep != nil {
		searchPath = filepath.Join(env.GetModulesDir(), dep.Name())

		packageData, err := domain.LoadPackageOther(filepath.Join(env.GetModulesDir(), dep.Name(), consts.FilePackage))
		if err == nil {
			searchPath += ";" + filepath.Join(env.GetModulesDir(), dep.Name(), packageData.MainSrc)
			for _, lib := range packageData.GetParsedDependencies() {
				searchPath += ";" + buildSearchPath(&lib)
			}
		}
	}
	return searchPath
}

func compile(dprojPath string, dep *domain.Dependency, rootLock domain.PackageLock, tracker *BuildTracker, selectedCompiler *compiler_selector.SelectedCompiler) bool {
	if tracker == nil || !tracker.IsEnabled() {
		msg.Info("  Building " + filepath.Base(dprojPath))
	}

	bossPackagePath := filepath.Join(env.GetModulesDir(), dep.Name(), consts.FilePackage)

	if dependencyPackage, err := domain.LoadPackageOther(bossPackagePath); err == nil {
		dcp.InjectDpcsFile(dprojPath, dependencyPackage, rootLock)
	}

	dccDir := env.GetDcc32Dir()
	platform := consts.PlatformWin32.String()
	compilerBinary := "dcc32.exe"

	if selectedCompiler != nil {
		dccDir = selectedCompiler.BinDir
		if selectedCompiler.Arch != "" {
			platform = selectedCompiler.Arch
		}
		switch selectedCompiler.Arch {
		case consts.PlatformWin64.String():
			compilerBinary = "dcc64.exe"
		case consts.PlatformOSX64.String():
			compilerBinary = "dccosx.exe"
		case consts.PlatformLinux64.String():
			compilerBinary = "dcclinux64.exe"
		}
	}

	if tracker == nil || !tracker.IsEnabled() {
		msg.Debug("  Using: %s (Platform: %s)", filepath.Join(dccDir, compilerBinary), platform)
	}

	rsvars := filepath.Join(dccDir, "rsvars.bat")
	fileRes := "build_boss_" + strings.TrimSuffix(filepath.Base(dprojPath), filepath.Ext(dprojPath))
	abs, _ := filepath.Abs(filepath.Dir(dprojPath))
	buildLog := filepath.Join(abs, fileRes+".log")
	buildBat := filepath.Join(abs, fileRes+".bat")
	readFile, err := os.ReadFile(rsvars)
	if err != nil {
		msg.Err("    error on read rsvars.bat")
	}
	readFileStr := string(readFile)
	project, _ := filepath.Abs(dprojPath)

	readFileStr += "\n@SET DCC_UnitSearchPath=%DCC_UnitSearchPath%;" +
		filepath.Join(env.GetModulesDir(), consts.DcuFolder) +
		";" + filepath.Join(env.GetModulesDir(), consts.DcpFolder) //+ ";" + getNewPathsDep(dep, abs) + " "

	readFileStr += ";" + buildSearchPath(dep)

	readFileStr += "\n@SET PATH=%PATH%;" + filepath.Join(env.GetModulesDir(), consts.BplFolder) + ";"
	for _, value := range []string{platform} {
		readFileStr += " \n msbuild \"" +
			project +
			"\" /p:Configuration=Debug " +
			getCompilerParameters(env.GetModulesDir(), dep, value)
	}
	readFileStr += " > \"" + buildLog + "\""

	err = os.WriteFile(buildBat, []byte(readFileStr), 0600)
	if err != nil {
		if tracker == nil || !tracker.IsEnabled() {
			msg.Warn("  - error on create build file")
		}
		return false
	}

	command := exec.Command(buildBat)
	command.Dir = abs
	if _, err = command.Output(); err != nil {
		if tracker == nil || !tracker.IsEnabled() {
			msg.Err("  - Failed to compile, see " + buildLog + " for more information")
		}
		return false
	}
	if tracker == nil || !tracker.IsEnabled() {
		msg.Info("  - Success!")
	}
	err = os.Remove(buildLog)
	utils.HandleError(err)
	err = os.Remove(buildBat)
	utils.HandleError(err)

	return true
}
