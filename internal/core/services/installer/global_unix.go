//go:build !windows

package installer

import (
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/msg"
)

func GlobalInstall(args []string, pkg *domain.Package, lockedVersion bool, noSave bool) {
	EnsureDependency(pkg, args)
	DoInstall(InstallOptions{
		Args:          args,
		LockedVersion: lockedVersion,
		NoSave:        noSave,
	}, pkg)
	msg.Err("Cannot install global packages on this platform, only build and install local")
}
