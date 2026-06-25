// Package installer provides utility functions for dependency management.
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

// parseUrlAndVersion parses the dependency URL and version, handling SSH git@ URLs safely.
func parseUrlAndVersion(input string) (url string, version string) {
	if strings.HasPrefix(input, "git@") {
		// SSH URL format: git@host:owner/repo[.git][:version] or git@host:owner/repo[.git]@version
		firstColon := strings.Index(input, ":")
		if firstColon == -1 {
			return input, ""
		}

		remainder := input[firstColon+1:]

		// Look for last colon or @ in remainder to see if a version is specified
		lastSep := strings.LastIndexAny(remainder, ":@")
		if lastSep != -1 {
			possibleVersion := remainder[lastSep+1:]
			// Version shouldn't contain slashes (which paths do) and shouldn't be empty
			if !strings.Contains(possibleVersion, "/") && possibleVersion != "" {
				url = input[:firstColon+1+lastSep]
				version = possibleVersion
				return url, version
			}
		}
		return input, ""
	}

	// For standard HTTP/HTTPS/standard paths, use the original regex
	match := make(map[string]string)
	split := reURLVersion.FindStringSubmatch(input)
	if len(split) > 0 {
		for i, name := range reURLVersion.SubexpNames() {
			if i != 0 && name != "" && i < len(split) {
				match[name] = split[i]
			}
		}
	}

	url = match["url"]
	version = match["version"]
	return url, version
}

// EnsureDependency ensures that the dependencies are added to the package.
func EnsureDependency(pkg *domain.Package, args []string) {
	for _, dependency := range args {
		dependency = ParseDependency(dependency)

		dep, ver := parseUrlAndVersion(dependency)
		if ver == "" {
			ver = consts.MinimalDependencyVersion
		}

		if strings.HasSuffix(strings.ToLower(dep), ".git") {
			dep = dep[:len(dep)-4]
		}

		pkg.AddDependency(dep, ver)
	}
}

// ParseDependency parses the dependency name and returns the full URL if needed.
func ParseDependency(dependencyName string) string {
	if strings.HasPrefix(dependencyName, "git@") || strings.HasPrefix(dependencyName, "http://") || strings.HasPrefix(dependencyName, "https://") {
		return dependencyName
	}
	if !reHasSlash.MatchString(dependencyName) {
		return "github.com/hashload/" + dependencyName
	}
	if !reHasMultiSlash.MatchString(dependencyName) {
		return "github.com/" + dependencyName
	}
	return dependencyName
}
