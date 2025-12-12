package installer

import (
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/utils/dcp"
)

func LocalInstall(args []string, pkg *domain.Package, lockedVersion bool, _ /* noSave */ bool) {
	// TODO noSave
	EnsureDependency(pkg, args)
	DoInstall(pkg, lockedVersion)
	dcp.InjectDpcs(pkg, pkg.Lock)
}
