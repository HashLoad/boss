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
	paths.EnsureCleanModulesDir()

	msg.Info("Installing modules in project patch")

	core.EnsureDependencies(pkg)

	librarypath.UpdateLibraryPath()

	msg.Info("Compiling units")

	compiler.BuildDucs()

	pkg.Save()
}
