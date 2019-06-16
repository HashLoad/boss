package compiler

import (
	"fmt"
	"github.com/hashload/boss/models"
)

func dependencyOrder(pkg *models.Package) {
	rawDeps := pkg.Dependencies.(map[string]interface{})
	deps := models.GetDependencies(rawDeps)

	fmt.Printf("%v", deps)
}
