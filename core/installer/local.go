package installer

import (
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/utils/dcp"
)

func LocalInstall(args []string, pkg *models.Package, lockedVersion bool) {
	EnsureDependencyOfArgs(pkg, args)
	DoInstall(pkg, lockedVersion)
	dcp.InjectDpcs(pkg, pkg.Lock)
}
