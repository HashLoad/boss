package installer

import (
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils/dcp"
	"os"
)

func LocalInstall(args []string) {

	pkg, e := models.LoadPackage(false)

	if e != nil {
		if os.IsNotExist(e) {
			msg.Die("boss.json not exists in " + env.GetCurrentDir())
		} else {
			msg.Die("Fail on open dependencies file: %s", e)
		}
	}

	EnsureDependencyOfArgs(pkg, args)
	DoInstall(pkg)

	dcp.InjectDpcs()
}
