package setup

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/msg"
	"github.com/snakeice/penv"
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
	addPath(consts.EnvBossBin)
	addPath(env.GetGlobalBinPath())
}

func addPath(path string) {
	environment, e := penv.Load()
	if e != nil {
		msg.Die("Failed to load env \n %s", e.Error())
	}

	currentPath := getPath(environment.Setters)
	if !strings.Contains(currentPath, path) {
		msg.Info("Initializing boss in your system...")
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
