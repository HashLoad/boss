// Package installer provides dependency installation and uninstallation functionality.
// It manages both global and local dependency installations, handling version locking and updates.
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

// InstallOptions holds the options for the installation process.
type InstallOptions struct {
	Args          []string
	LockedVersion bool
	NoSave        bool
	Compiler      string
	Platform      string
	Strict        bool
	ForceUpdate   []string
}

// createLockService creates a new lock service instance.
func createLockService() *lockService.LockService {
	fs := filesystem.NewOSFileSystem()
	lockRepo := repository.NewFileLockRepository(fs)
	return lockService.NewLockService(lockRepo, fs)
}

// InstallModules installs the modules based on the provided options.
func InstallModules(options InstallOptions) {
	pkg, err := domain.LoadPackage(env.GetGlobal())
	if err != nil {
		if os.IsNotExist(err) {
			msg.Die("❌ 'boss.json' not exists in " + env.GetCurrentDir())
		} else {
			msg.Die("❌ Fail on open dependencies file: %s", err)
		}
	}

	if env.GetGlobal() {
		GlobalInstall(env.GlobalConfiguration(), options.Args, pkg, options.LockedVersion, options.NoSave)
	} else {
		LocalInstall(env.GlobalConfiguration(), options, pkg)
	}
}

// UninstallModules uninstalls the specified modules.
func UninstallModules(args []string, noSave bool) {
	pkg, err := domain.LoadPackage(false)
	if err != nil && !os.IsNotExist(err) {
		msg.Die("❌ Fail on open dependencies file: %s", err)
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

	InstallModules(InstallOptions{
		Args:          []string{},
		LockedVersion: false,
		NoSave:        noSave,
	})
}
