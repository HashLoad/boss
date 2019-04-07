package installer

import (
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
)

func LocalInstall(args []string) {

	pkg, e := models.LoadPackage(false)

	if e != nil {
		msg.Die("Fail on open dependencies file: %s", e)
	}

	EnsureDependencyOfArgs(pkg, args)

	DoInstall(pkg)
}
