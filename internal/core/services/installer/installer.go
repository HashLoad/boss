package installer

import (
	"os"

	"github.com/hashload/boss/internal/adapters/secondary/filesystem"
	"github.com/hashload/boss/internal/adapters/secondary/repository"
	"github.com/hashload/boss/internal/core/domain"
	lockService "github.com/hashload/boss/internal/core/services/lock"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
)

// createLockService creates a new lock service instance.
func createLockService() *lockService.Service {
	fs := filesystem.NewOSFileSystem()
	lockRepo := repository.NewFileLockRepository(fs)
	return lockService.NewService(lockRepo, fs)
}

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
	lockSvc := createLockService()
	_ = lockSvc.Save(&pkg.Lock, env.GetCurrentDir())

	InstallModules([]string{}, false, noSave)
}
