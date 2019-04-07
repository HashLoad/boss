package installer

import (
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"path/filepath"
)

func GlobalInstall(args []string) {

	pkg, _ := models.LoadPackageOther(filepath.Join(env.GetCurrentDir(), "GLOBAL"))
	EnsureDependencyOfArgs(pkg, args)

	DoInstall(pkg)
}
