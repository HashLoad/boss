package installer

import (
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/hashload/boss/pkg/msg"
)

// npmRangePattern detects npm-style hyphen ranges (1.0.0 - 2.0.0)
var npmRangePattern = regexp.MustCompile(`^\s*([v\d][^\s]*)\s*-\s*([v\d][^\s]*)\s*$`)

// ParseConstraint parses a version constraint, converting npm-style ranges to Go-compatible format.
func ParseConstraint(constraintStr string) (*semver.Constraints, error) {
	constraint, err := semver.NewConstraint(constraintStr)
	if err == nil {
		return constraint, nil
	}

	if matches := npmRangePattern.FindStringSubmatch(constraintStr); matches != nil {
		start := strings.TrimPrefix(matches[1], "v")
		end := strings.TrimPrefix(matches[2], "v")
		converted := ">=" + start + " <=" + end
		msg.Info("Converting npm-style range '%s' to '%s'", constraintStr, converted)
		return semver.NewConstraint(converted)
	}

	converted := convertNpmConstraint(constraintStr)
	if converted != constraintStr {
		msg.Info("Converting constraint '%s' to '%s'", constraintStr, converted)
		return semver.NewConstraint(converted)
	}

	return nil, err
}

// convertNpmConstraint converts common npm constraint patterns to Go-compatible format.
func convertNpmConstraint(constraint string) string {
	constraint = strings.ReplaceAll(constraint, ".x", ".*")
	constraint = strings.ReplaceAll(constraint, ".X", ".*")
	constraint = strings.ReplaceAll(constraint, " && ", " ")
	return constraint
}

// NormalizeVersion removes common prefixes and ensures valid semver format.
func NormalizeVersion(version string) string {
	version = strings.TrimPrefix(version, "v")
	version = strings.TrimPrefix(version, "V")
	version = strings.TrimPrefix(version, "release-")
	version = strings.TrimPrefix(version, "version-")
	return version
}
