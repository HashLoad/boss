package installer

import (
	"regexp"
	"strings"

	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/models"
)

func EnsureDependencyOfArgs(pkg *models.Package, args []string) {
	for e := range args {
		dependency := ParseDependency(args[e])

		re := regexp.MustCompile(`(?U)(?m)^(?:http[s]{0,1}://)?(?P<host>.*)(?::(?P<version>[\^~]?(?:\d+\.)?(?:\d+\.)?(?:\*|\d+)))?$`)
		match := make(map[string]string)
		split := re.FindStringSubmatch(dependency)

		for i, name := range re.SubexpNames() {
			if i != 0 && name != "" {
				match[name] = split[i]
			}
		}
		var ver string
		var dep string
		dep = match["host"]
		if len(match["version"]) == 0 {
			ver = consts.MinimalDependencyVersion
		} else {
			ver = match["version"]
		}

		if strings.HasSuffix(strings.ToLower(dep), ".git") {
			dep = dep[:len(dep)-4]
		}

		pkg.AddDependency(dep, ver)
	}
}

func ParseDependency(dependencyName string) string {
	re := regexp.MustCompile(`(?m)(([?^/]).*)`)
	if !re.Match([]byte(dependencyName)) {
		return "github.com/hashload/" + dependencyName
	}
	re = regexp.MustCompile(`(?m)([?^/].*)(([?^/]).*)`)
	if !re.Match([]byte(dependencyName)) {
		return "github.com/" + dependencyName
	}
	return dependencyName
}
