package installer

import (
	"os"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
)

func InstallModules(args []string, lockedVersion bool, noSave bool) {
	pkg, err := domain.LoadPackage(env.GetGlobal())
	if err != nil {
		if os.IsNotExist(err) {
			msg.Die("boss.json not exists in " + env.GetCurrentDir())
		} else {
			msg.Die("Fail on open dependencies file: %s", err)
		}
	}

	if env.GetGlobal() {
		GlobalInstall(args, pkg, lockedVersion, noSave)
	} else {
		LocalInstall(args, pkg, lockedVersion, noSave)
	}
}

func UninstallModules(args []string, noSave bool) {
	pkg, err := domain.LoadPackage(false)
	if err != nil && !os.IsNotExist(err) {
		msg.Die("Fail on open dependencies file: %s", err)
	}

	if pkg == nil {
		return
	}

	for _, arg := range args {
		dependencyRepository := ParseDependency(arg)
		pkg.UninstallDependency(dependencyRepository)
	}

	pkg.Save()

	// TODO implement remove without reinstall process
	InstallModules([]string{}, false, noSave)
}
