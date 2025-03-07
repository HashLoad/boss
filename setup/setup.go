package setup

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/core/installer"
	"github.com/hashload/boss/core/registry"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"github.com/hashload/boss/utils/dcc32"
)

const PATH string = "PATH"

var (
	defaultModules = []string{
		"bpl-identifier",
	}
)

func Initialize() {
	var OldGlobal = env.Global
	env.Internal = true
	env.Global = true

	msg.Debug("DEBUG MODE")
	msg.Debug("\tInitializing delphi version")
	initializeDelphiVersion()

	paths := []string{
		consts.EnvBossBin,
		env.GetGlobalBinPath(),
		env.GetGlobalEnvBpl(),
		env.GetGlobalEnvDcu(),
		env.GetGlobalEnvDcp(),
	}

	msg.Debug("\tExecuting migrations")
	migration()
	msg.Debug("\tAdjusting paths")
	addPaths(paths)
	msg.Debug("\tInstalling internal modules")
	installModules(defaultModules)
	msg.Debug("\tCreating paths")
	createPaths()

	env.Global = OldGlobal
	env.Internal = false
	msg.Debug("finish boss system initialization")

}

func createPaths() {
	_, err := os.Stat(env.GetGlobalEnvBpl())
	if os.IsNotExist(err) {
		_ = os.MkdirAll(env.GetGlobalEnvBpl(), os.ModePerm)
	}
}

func addPaths(paths []string) {
	var needAdd = false
	currentPath, e := os.Getwd()
	if e != nil {
		msg.Die("Failed to load current working directory \n %s", e.Error())
		return
	}

	splitPath := strings.Split(currentPath, ";")

	for _, path := range paths {
		if !utils.Contains(splitPath, path) {
			splitPath = append(splitPath, path)
			needAdd = true
		}
	}

	if needAdd {
		newPath := strings.Join(splitPath, ";")
		currentPathEnv := os.Getenv(PATH)
		err := os.Setenv(PATH, currentPathEnv+";"+newPath)
		if err != nil {
			msg.Die("Failed to update PATH \n %s", err.Error())
			return

		}

		msg.Warn("Please restart your console after complete.")
	}

}

func installModules(modules []string) {
	pkg, _ := models.LoadPackage(true)
	dependencies := pkg.Dependencies.(map[string]any)
	encountered := 0
	for _, newPackage := range modules {
		for installed := range dependencies {
			if strings.Contains(installed, newPackage) {
				encountered++
			}
		}
	}

	nextUpdate := env.GlobalConfiguration.LastInternalUpdate.
		AddDate(0, 0, env.GlobalConfiguration.PurgeTime)

	if encountered == len(modules) && time.Now().Before(nextUpdate) {
		return
	}

	env.GlobalConfiguration.LastInternalUpdate = time.Now()
	env.GlobalConfiguration.SaveConfiguration()

	installer.GlobalInstall(modules, pkg, false, false)
	moveBptIdentifier()
}

func moveBptIdentifier() {
	var OutExeCompilation = filepath.Join(env.GetGlobalBinPath(), consts.BplIdentifierName)
	if _, err := os.Stat(OutExeCompilation); os.IsNotExist(err) {
		return
	}

	var exePath = filepath.Join(env.GetModulesDir(), consts.BinFolder, consts.BplIdentifierName)
	err := os.MkdirAll(filepath.Dir(exePath), os.ModePerm)
	if err != nil {
		msg.Err(err.Error())
	}

	err = os.Rename(OutExeCompilation, exePath)
	if err != nil {
		msg.Err(err.Error())
	}
}

func initializeDelphiVersion() {
	if len(env.GlobalConfiguration.DelphiPath) != 0 {
		return
	}
	dcc32DirByCmd := dcc32.GetDcc32DirByCmd()
	if len(dcc32DirByCmd) != 0 {
		env.GlobalConfiguration.DelphiPath = dcc32DirByCmd[0]
		env.GlobalConfiguration.SaveConfiguration()
		return
	}

	byRegistry := registry.GetDelphiPaths()
	if len(byRegistry) != 0 {
		env.GlobalConfiguration.DelphiPath = byRegistry[len(byRegistry)-1]
		env.GlobalConfiguration.SaveConfiguration()
		return
	}

}
