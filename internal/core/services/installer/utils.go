package installer

import (
	"regexp"
	"strings"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/consts"
)

//nolint:lll // This regex is too long and it's better to keep it like this
const urlVersionMatcher = `(?m)^(?:http[s]?:\/\/|git@)?(?P<url>[\w\.\-\/:]+?)(?:[@:](?P<version>[\^~]?(?:\d+\.)?(?:\d+\.)?(?:\*|\d+|[\w\-]+)))?$`

var (
	reURLVersion    = regexp.MustCompile(urlVersionMatcher)
	reHasSlash      = regexp.MustCompile(`(?m)(([?^/]).*)`)
	reHasMultiSlash = regexp.MustCompile(`(?m)([?^/].*)(([?^/]).*)`)
)

// EnsureDependency ensures that the dependencies are added to the package.
func EnsureDependency(pkg *domain.Package, args []string) {
	for _, dependency := range args {
		dependency = ParseDependency(dependency)

		match := make(map[string]string)
		split := reURLVersion.FindStringSubmatch(dependency)

		for i, name := range reURLVersion.SubexpNames() {
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

// ParseDependency parses the dependency name and returns the full URL if needed.
func ParseDependency(dependencyName string) string {
	if !reHasSlash.MatchString(dependencyName) {
		return "github.com/hashload/" + dependencyName
	}
	if !reHasMultiSlash.MatchString(dependencyName) {
		return "github.com/" + dependencyName
	}
	return dependencyName
}
