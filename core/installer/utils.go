package installer

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/models"
	"regexp"
	"strings"
)

func EnsureDependencyOfArgs(pkg *models.Package, args []string) {
	for e := range args {
		dependency := ParseDependency(args[e])
		dependency = strings.ToLower(dependency)

		re := regexp.MustCompile(`(?m)((.*)(:\W[\d.]+)|.*)$`)
		split := re.FindAllString(dependency, -1)
		var ver string
		if len(split) == 1 {
			ver = consts.MinimalDependencyVersion
		} else {
			ver = split[1]
		}
		pkg.AddDependency(split[0], ver)
	}
}

func ParseDependency(dependencyName string) string {
	re := regexp.MustCompile(`(?m)(([?^/]).*)`)
	if !re.Match([]byte(dependencyName)) {
		return "github.com/HashLoad/" + dependencyName
	}
	re = regexp.MustCompile(`(?m)([?^/].*)(([?^/]).*)`)
	if !re.Match([]byte(dependencyName)) {
		return "github.com/" + dependencyName
	}
	return dependencyName
}
