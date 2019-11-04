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

		re := regexp.MustCompile(`(?m)(?:(?P<host>.*)(?::(?P<version>[\^~]?(?:(?:(?:[0-9]+)(?:\.[0-9]+)(?:\.[0-9]+)?))))$|(?P<host_only>.*))`)
		match := make(map[string]string)
		split := re.FindStringSubmatch(dependency)

		for i, name := range re.SubexpNames() {
			if i != 0 && name != "" {
				match[name] = split[i]
			}
		}
		var ver string
		var dep string
		if len(match["version"]) == 0 {
			ver = consts.MinimalDependencyVersion
			dep = match["host_only"]
		} else {
			ver = match["version"]
			dep = match["host"]
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
		return "github.com/HashLoad/" + dependencyName
	}
	re = regexp.MustCompile(`(?m)([?^/].*)(([?^/]).*)`)
	if !re.Match([]byte(dependencyName)) {
		return "github.com/" + dependencyName
	}
	return dependencyName
}
