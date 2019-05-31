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
	msg.Info("Initializing boss system...")

	msg.Debug("\tInitializing delphi version")
	initializeDelphiVersion()
	paths := []string{consts.EnvBossBin, env.GetGlobalBinPath()}
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

	env.Global = OldGlobal
	env.Internal = false
	msg.Debug("finish boss system initialization")

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
		for installed, _ := range dependencies {
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

	installer.GlobalInstall(modules, pkg)
	moveBptIdentifier()

}

func moveBptIdentifier() {

	var exePath = filepath.Join(env.GetModulesDir(), consts.BinFolder, consts.BplIdentifierName)
	err := os.MkdirAll(filepath.Dir(exePath), os.ModePerm)
	if err != nil {
		msg.Err(err.Error())
	}

	var OutExeCompilation = filepath.Join(env.GetGlobalBinPath(), consts.BplIdentifierName)

	err = os.Rename(OutExeCompilation, exePath)
	if err != nil {
		msg.Err(err.Error())
	}
}

func migration() {
	if env.GlobalConfiguration.ConfigVersion < 1 {
		env.GlobalConfiguration.InternalRefreshRate = 5
		env.GlobalConfiguration.ConfigVersion++
		env.GlobalConfiguration.SaveConfiguration()
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

	byRegisty := dcc32.GetDelphiPathsByRegisty()
	if len(byRegisty) != 0 {
		env.GlobalConfiguration.DelphiPath = byRegisty[len(byRegisty)-1]
		env.GlobalConfiguration.SaveConfiguration()
		return
	}

}
