package installer

import (
	"github.com/hashload/boss/core"
	"github.com/hashload/boss/core/compiler"
	"github.com/hashload/boss/core/paths"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils/librarypath"
)

func DoInstall(pkg *models.Package) {
	msg.Info("Installing modules in project path")

	dependencies := core.EnsureDependencies(pkg.Lock, pkg)
	paths.EnsureCleanModulesDir(dependencies, pkg.Lock)

	pkg.Lock.CleanRemoved(dependencies)
	pkg.Save()

	librarypath.UpdateLibraryPath(pkg)
	msg.Info("Compiling units")
	compiler.Build(pkg)
	pkg.Save()
}
