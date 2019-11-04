package setup

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/core/installer"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils/dcc32"
	"github.com/snakeice/penv"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const PATH string = "PATH"

func Initialize() {
	var OldGlobal = env.Global
	env.Internal = true
	env.Global = true

	msg.Debug("DEBUG MODE")
	msg.Debug("\tInitializing delphi version")
	initializeDelphiVersion()

	paths := []string{consts.EnvBossBin, env.GetGlobalBinPath(), env.GetGlobalEnvBpl(), env.GetGlobalEnvDcu(), env.GetGlobalEnvDcp()}

	modules := []string{"bpl-identifier"}

	msg.Debug("\tExecuting migrations")
	migration()
	msg.Debug("\tAdjusting paths")
	tester := make(chan string, 1)
	go func() {
		addPath(paths)
		tester <- "ok"
	}()
	select {
	case _ = <-tester:
	case <-time.After(time.Second * 10):
		msg.Warn("Failed to update paths, please run with administrator privileges")
	}
	msg.Debug("\tInstalling internal modules")
	installModules(modules)
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

func getPath(arr []penv.NameValue) string {
	for _, nv := range arr {
		if nv.Name == PATH {
			return nv.Value
		}
	}
	return ""
}

func addPath(paths []string) {
	var needAdd = false
	environment, e := penv.Load()
	if e != nil {
		msg.Die("Failed to load env \n %s", e.Error())
		return
	}

	currentPath := getPath(environment.Setters)
	pathEnv := ""
	for _, path := range paths {
		if !strings.Contains(currentPath, path) {
			pathEnv = path + ";"
			needAdd = true
		}
	}

	if needAdd {
		if !strings.HasSuffix(currentPath, ";") {
			pathEnv = ";" + pathEnv
		}
		if err := penv.AppendEnv(PATH, pathEnv); err != nil {
			msg.Err("Failed to set env " + PATH)
			msg.Die(err.Error())
		}
		msg.Warn("Please restart your console after complete.")
	}

}

func installModules(modules []string) {
	pkg, _ := models.LoadPackage(true)
	dependencies := pkg.Dependencies.(map[string]interface{})
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

	installer.GlobalInstall(modules, pkg, false)
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

	byRegistry := dcc32.GetDelphiPathsByRegistry()
	if len(byRegistry) != 0 {
		env.GlobalConfiguration.DelphiPath = byRegistry[len(byRegistry)-1]
		env.GlobalConfiguration.SaveConfiguration()
		return
	}

}
