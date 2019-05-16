package installer

import (
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/utils/dcp"
)

func LocalInstall(args []string, pkg *models.Package) {

	EnsureDependencyOfArgs(pkg, args)
	DoInstall(pkg)

	dcp.InjectDpcs(pkg)
}
