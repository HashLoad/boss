package installer

import (
	"regexp"
	"strings"

	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/models"
)

//nolint:lll // This regex is too long and it's better to keep it like this
const urlVersionMatcher = `(?m)^(?:http[s]?:\/\/|git@)?(?P<url>[\w\.\-\/:]+?)(?:[@:](?P<version>[\^~]?(?:\d+\.)?(?:\d+\.)?(?:\*|\d+|[\w\-]+)))?$`

func EnsureDependencyOfArgs(pkg *models.Package, args []string) {
	for _, dependency := range args {
		dependency = ParseDependency(dependency)

		re := regexp.MustCompile(urlVersionMatcher)
		match := make(map[string]string)
		split := re.FindStringSubmatch(dependency)

		for i, name := range re.SubexpNames() {
			if i != 0 && name != "" {
				match[name] = split[i]
			}
		}
		var ver string
		var dep string
		dep = match["url"]
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
	if !re.MatchString(dependencyName) {
		return "github.com/hashload/" + dependencyName
	}
	re = regexp.MustCompile(`(?m)([?^/].*)(([?^/]).*)`)
	if !re.MatchString(dependencyName) {
		return "github.com/" + dependencyName
	}
	return dependencyName
}
