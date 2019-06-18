package installer

import (
	"regexp"
	"strings"

	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/models"
)

func EnsureDependencyOfArgs(pkg *models.Package, args []string) {
	for e := range args {
		dependency := ParseDependency(args[e])
		dependency = strings.ToLower(dependency)

		re := regexp.MustCompile(`(?m)((.*)(\:[*\W\d\.]{0,})|.*)$`)
		split := re.FindAllString(dependency, -1)
		var ver string
		if len(split) == 1 {
			ver = consts.MinimalDependencyVersion
		} else {
			ver = split[1][1:]
		}

		var dep = split[0]
		if strings.HasSuffix(strings.ToLower(dep), ".git") {
			dep = dep[:len(dep)-4]
		}

		pkg.AddDependency(dep, ver)
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
