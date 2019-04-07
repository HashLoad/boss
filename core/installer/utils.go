package installer

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/models"
	"strings"
)

func EnsureDependencyOfArgs(pkg *models.Package, args []string) {
	for e := range args {
		dependency := args[e]
		split := strings.Split(dependency, ":")
		var ver string
		if len(split) == 1 {
			ver = consts.MinimalDependencyVersion
		} else {
			ver = split[1]
		}
		pkg.AddDependency(split[0], ver)
	}
}
