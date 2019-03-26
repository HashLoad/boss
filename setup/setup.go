package setup

import (
	"github.com/hashload/boss/consts"
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
	environment, e := penv.Load()
	if e != nil {
		msg.Die("Failed to load env \n %s", e.Error())
	}

	currentPath := getPath(environment.Setters)
	if !strings.Contains(currentPath, consts.ENV_BOSS_BIN) {
		msg.Info("Initializing boss in your system...")
		pathEnv := consts.ENV_BOSS_BIN + ";"
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
