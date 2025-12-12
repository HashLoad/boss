//go:build !windows

package installer

import (
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/msg"
)

func GlobalInstall(args []string, pkg *domain.Package, lockedVersion bool, _ /* nosave */ bool) {
	EnsureDependency(pkg, args)
	DoInstall(pkg, lockedVersion)
	msg.Err("Cannot install global packages on this platform, only build and install local")
}
