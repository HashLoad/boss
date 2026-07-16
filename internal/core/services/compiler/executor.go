package compiler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/compilerselector"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/pkg/pkgmanager"
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
	var searchPath strings.Builder

	if dep != nil {
		searchPath.WriteString(filepath.Join(env.GetModulesDir(), dep.Name()))

		packageData, err := pkgmanager.LoadPackageOther(filepath.Join(env.GetModulesDir(), dep.Name(), consts.FilePackage))
		if err == nil {
			searchPath.WriteString(";")
			searchPath.WriteString(filepath.Join(env.GetModulesDir(), dep.Name(), packageData.MainSrc))
			for _, lib := range packageData.GetParsedDependencies() {
				searchPath.WriteString(";")
				searchPath.WriteString(buildSearchPath(&lib))
			}
		}
	}
	return searchPath.String()
}

//nolint:funlen,gocognit,lll // Complex compilation orchestration with long function signature
func compile(dprojPath string, dep *domain.Dependency, rootLock domain.PackageLock, tracker *BuildTracker, selectedCompiler *compilerselector.SelectedCompiler) bool {
	if tracker == nil || !tracker.IsEnabled() {
		msg.Info("  🔨 Building " + filepath.Base(dprojPath))
	}

	bossPackagePath := filepath.Join(env.GetModulesDir(), dep.Name(), consts.FilePackage)

	if dependencyPackage, err := pkgmanager.LoadPackageOther(bossPackagePath); err == nil {
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
		msg.Debug("  🛠️ Using: %s (Platform: %s)", filepath.Join(dccDir, compilerBinary), platform)
	}

	rsvars := filepath.Join(dccDir, "rsvars.bat")
	fileRes := "build_boss_" + strings.TrimSuffix(filepath.Base(dprojPath), filepath.Ext(dprojPath))
	abs, _ := filepath.Abs(filepath.Dir(dprojPath))
	buildLog := filepath.Join(abs, fileRes+".log")
	buildBat := filepath.Join(abs, fileRes+".bat")
	cfgPath := filepath.Join(abs, "boss.cfg")

	// Create boss.cfg to hold search paths and avoid command-line too long errors (Issue #205)
	var cfgContent strings.Builder
	dcuPath := filepath.Join(env.GetModulesDir(), consts.DcuFolder)
	dcpPath := filepath.Join(env.GetModulesDir(), consts.DcpFolder)
	fmt.Fprintf(&cfgContent, "-I\"%s\"\n-U\"%s\"\n", dcuPath, dcuPath)
	fmt.Fprintf(&cfgContent, "-I\"%s\"\n-U\"%s\"\n", dcpPath, dcpPath)

	if searchPathsStr := buildSearchPath(dep); searchPathsStr != "" {
		paths := strings.Split(searchPathsStr, ";")
		for _, p := range paths {
			p = strings.TrimSpace(p)
			if p != "" {
				fmt.Fprintf(&cfgContent, "-I\"%s\"\n-U\"%s\"\n", p, p)
			}
		}
	}

	err := os.WriteFile(cfgPath, []byte(cfgContent.String()), 0600)
	if err != nil {
		if tracker == nil || !tracker.IsEnabled() {
			msg.Warn("  ⚠️ Error on create compiler configuration file")
		}
		return false
	}

	readFile, err := os.ReadFile(rsvars) // #nosec G304 -- Reading Delphi environment variables file from known location
	if err != nil {
		msg.Err("    ❌ Error on read rsvars.bat")
	}
	readFileStr := string(readFile)
	project, _ := filepath.Abs(dprojPath)

	var scriptBuilder strings.Builder
	scriptBuilder.WriteString(readFileStr)
	scriptBuilder.WriteString("\n@SET PATH=%PATH%;")
	scriptBuilder.WriteString(filepath.Join(env.GetModulesDir(), consts.BplFolder))
	scriptBuilder.WriteString(";")
	for _, value := range []string{platform} {
		scriptBuilder.WriteString(" \n msbuild \"")
		scriptBuilder.WriteString(project)
		scriptBuilder.WriteString("\" /p:Configuration=Debug ")
		scriptBuilder.WriteString(getCompilerParameters(env.GetModulesDir(), dep, value))
		scriptBuilder.WriteString(" /p:DCC_AdditionalParameters=\"@")
		scriptBuilder.WriteString(cfgPath)
		scriptBuilder.WriteString("\"")
	}
	scriptBuilder.WriteString(" > \"")
	scriptBuilder.WriteString(buildLog)
	scriptBuilder.WriteString("\"")

	err = os.WriteFile(buildBat, []byte(scriptBuilder.String()), 0600)
	if err != nil {
		if tracker == nil || !tracker.IsEnabled() {
			msg.Warn("  ⚠️ Error on create build file")
		}
		_ = os.Remove(cfgPath)
		return false
	}

	command := exec.CommandContext(context.Background(), buildBat) // #nosec G204 -- Executing controlled build script generated by Boss
	command.Dir = abs
	if _, err = command.Output(); err != nil {
		if tracker == nil || !tracker.IsEnabled() {
			msg.Err("  ❌ Failed to compile, see " + buildLog + " for more information")
		}
		_ = os.Remove(cfgPath)
		return false
	}
	if tracker == nil || !tracker.IsEnabled() {
		msg.Info("  ✅️ Success!")
	}

	if err := os.Remove(buildLog); err != nil {
		msg.Debug("Could not remove build log %s: %v", buildLog, err)
	}
	if err := os.Remove(buildBat); err != nil {
		msg.Debug("Could not remove build script %s: %v", buildBat, err)
	}
	if err := os.Remove(cfgPath); err != nil {
		msg.Debug("Could not remove boss.cfg %s: %v", cfgPath, err)
	}

	return true
}
