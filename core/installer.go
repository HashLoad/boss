package core

import (
	"os"

	"github.com/hashload/boss/core/installer"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
)

func InstallModules(args []string, lockedVersion bool) {
	_ = lockedVersion
	pkg, e := models.LoadPackage(env.Global)
	if e != nil {
		if os.IsNotExist(e) {
			msg.Die("boss.json not exists in " + env.GetCurrentDir())
		} else {
			msg.Die("Fail on open dependencies file: %s", e)
		}
	}

	if env.Global {
		installer.GlobalInstall(args, pkg, lockedVersion)
	} else {
		installer.LocalInstall(args, pkg, lockedVersion)
	}
}

func UninstallModules(args []string) {
	pkg, e := models.LoadPackage(false)
	if e != nil {
		msg.Err(e.Error())
	}

	if pkg == nil {
		return
	}

	for e := range args {
		pkg.UninstallDependency(installer.ParseDependency(installer.ParseDependency(args[e])))
	}
	pkg.Save()
	//TODO implement remove without reinstall process
	InstallModules([]string{}, false)
}
