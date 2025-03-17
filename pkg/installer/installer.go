package installer

import (
	"os"

	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/models"
	"github.com/hashload/boss/pkg/msg"
)

func InstallModules(args []string, lockedVersion bool, noSave bool) {
	_ = lockedVersion
	pkg, e := models.LoadPackage(env.GetGlobal())
	if e != nil {
		if os.IsNotExist(e) {
			msg.Die("boss.json not exists in " + env.GetCurrentDir())
		} else {
			msg.Die("Fail on open dependencies file: %s", e)
		}
	}

	if env.GetGlobal() {
		GlobalInstall(args, pkg, lockedVersion, noSave)
	} else {
		LocalInstall(args, pkg, lockedVersion, noSave)
	}
}

func UninstallModules(args []string, _ /* noSave */ bool) {
	pkg, e := models.LoadPackage(false)

	if e != nil {
		msg.Err(e.Error())
	}

	if pkg == nil {
		return
	}

	for e := range args {
		dependencyRepository := ParseDependency(args[e])
		pkg.UninstallDependency(dependencyRepository)
	}

	pkg.Save()

	// TODO noSave
	// TODO implement remove without reinstall process

	InstallModules([]string{}, false, false)
}
