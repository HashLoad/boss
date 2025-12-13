package installer

import (
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/utils/dcp"
)

func LocalInstall(options InstallOptions, pkg *domain.Package) {
	// TODO noSave
	EnsureDependency(pkg, options.Args)
	DoInstall(options, pkg)
	dcp.InjectDpcs(pkg, pkg.Lock)
}
