package setup

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/core/installer"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/snakeice/penv"
	"os"
	"path/filepath"
	"strings"
)

const PATH string = "PATH"

func getPath(arr []penv.NameValue) string {
	for _, nv := range arr {
		if nv.Name == PATH {
			return nv.Value
		}
	}
	return ""
}

func Initialize() {
	var OldGlobal = env.Global
	env.Internal = true
	env.Global = true

	msg.Info("Initializing boss system...")
	addPath(consts.EnvBossBin)
	addPath(env.GetGlobalBinPath())
	installBplIdentifier()

	env.Global = OldGlobal
	env.Internal = false
}

func addPath(path string) {
	environment, e := penv.Load()
	if e != nil {
		msg.Die("Failed to load env \n %s", e.Error())
	}

	currentPath := getPath(environment.Setters)
	if !strings.Contains(currentPath, path) {
		pathEnv := path + ";"
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

func installBplIdentifier() {
	var exePath = filepath.Join(env.GetModulesDir(), consts.BinFolder, consts.BplIdentifierName)
	filepath.Join(env.GetModulesDir(), consts.BinFolder)
	if _, err := os.Stat(exePath); os.IsNotExist(err) {

		pkg, _ := models.LoadPackage(true)
		installer.EnsureDependencyOfArgs(pkg, []string{"github.com/HashLoad/bpl-identifier"})
		installer.DoInstall(pkg)

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

}
