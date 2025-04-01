package installer

import (
	"github.com/hashload/boss/pkg/models"
	"github.com/hashload/boss/utils/dcp"
)

func LocalInstall(args []string, pkg *models.Package, lockedVersion bool, _ /* noSave */ bool) {
	// TODO noSave
	EnsureDependency(pkg, args)
	DoInstall(pkg, lockedVersion)
	dcp.InjectDpcs(pkg, pkg.Lock)
}
