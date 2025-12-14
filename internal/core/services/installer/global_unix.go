//go:build !windows

package installer

import (
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
)

// GlobalInstall installs dependencies globally (Unix implementation).
func GlobalInstall(config env.ConfigProvider, args []string, pkg *domain.Package, lockedVersion bool, noSave bool) {
	EnsureDependency(pkg, args)
	if err := DoInstall(config, InstallOptions{
		Args:          args,
		LockedVersion: lockedVersion,
		NoSave:        noSave,
	}, pkg); err != nil {
		msg.Die("❌ %s", err)
	}
	msg.Err("❌ Cannot install global packages on this platform, only build and install local")
}
