//go:build !windows

package installer

import (
	"github.com/hashload/boss/pkg/models"
	"github.com/hashload/boss/pkg/msg"
)

func GlobalInstall(args []string, pkg *models.Package, lockedVersion bool, noSave bool) {
	EnsureDependencyOfArgs(pkg, args)
	DoInstall(pkg, lockedVersion)
	msg.Err("Cannot install global packages on this platform, only build and install local")
}
