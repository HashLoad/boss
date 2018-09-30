package core

import (
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
)

func EnsureDependencies(pkg *models.Package) {
	if pkg.Dependencies == nil {
		return
	}
	rawDeps := pkg.Dependencies.(map[string]interface{})

	deps := models.GetDependencies(rawDeps)

	makeCache(deps)

	ensureModules
}

func makeCache(deps []models.Dependency)  {
	msg.Info("Building cache files..")

	for _, dep := range deps {
		GetDependency(dep)
	}
}

func ensureModules(deps []models.Dependency)  {
	msg.Info("Installing modules in project paht");
	for _, dep := range deps {
		GetDependency(dep)
	}


}
