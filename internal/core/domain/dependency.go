package domain

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"

	"github.com/hashload/boss/pkg/msg"
)

var (
	reSSHUrl            = regexp.MustCompile(`(?m)([\w\d.]*)(?:/)(.*)`)
	reURLPrefix         = regexp.MustCompile(`^[^/^:]+`)
	reHasHTTPS          = regexp.MustCompile(`(?m)^https?:\/\/`)
	reVersionMajorMinor = regexp.MustCompile(`(?m)^(.|)(\d+)\.(\d+)$`)
	reVersionMajor      = regexp.MustCompile(`(?m)^(.|)(\d+)$`)
	reDepName           = regexp.MustCompile(`[^/]+(:?/$|$)`)
)

type Dependency struct {
	Repository string
	version    string
	UseSSH     bool
}

// HashName returns the MD5 hash of the repository name
func (p *Dependency) HashName() string {
	//nolint:gosec // We are not using this for security purposes
	hash := md5.New()
	if _, err := io.WriteString(hash, strings.ToLower(p.Repository)); err != nil {
		msg.Warn("Failed on write dependency hash")
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// GetVersion returns the version of the dependency
func (p *Dependency) GetVersion() string {
	return p.version
}

// SSHUrl returns the SSH URL format for the repository.
func (p *Dependency) SSHUrl() string {
	if strings.Contains(p.Repository, "@") {
		return p.Repository
	}
	submatch := reSSHUrl.FindStringSubmatch(p.Repository)
	provider := submatch[1]
	repo := submatch[2]
	return "git@" + provider + ":" + repo
}

// GetURLPrefix returns the provider prefix of the repository URL
func (p *Dependency) GetURLPrefix() string {
	return reURLPrefix.FindString(p.Repository)
}

// GetURL returns the full URL for the repository, handling SSH and HTTPS
func (p *Dependency) GetURL() string {
	prefix := p.GetURLPrefix()
	auth := env.GlobalConfiguration().Auth[prefix]
	if auth != nil {
		if auth.UseSSH {
			return p.SSHUrl()
		}
	}
	if p.UseSSH {
		return p.SSHUrl()
	}
	if reHasHTTPS.MatchString(p.Repository) {
		return p.Repository
	}

	return "https://" + p.Repository
}

// ParseDependency creates a Dependency object from repository string and version info
func ParseDependency(repo string, info string) Dependency {
	parsed := strings.Split(info, ":")
	dependency := Dependency{}
	dependency.Repository = repo
	dependency.version = parsed[0]
	if reVersionMajorMinor.MatchString(dependency.version) {
		msg.Debug("Current version for %s is not semantic (x.y.z), for comparison using %s -> %s",
			dependency.Repository, dependency.version, dependency.version+".0")
		dependency.version += ".0"
	}
	if reVersionMajor.MatchString(dependency.version) {
		msg.Debug("Current version for %s is not semantic (x.y.z), for comparison using %s -> %s",
			dependency.Repository, dependency.version, dependency.version+".0.0")
		dependency.version += ".0.0"
	}
	if len(parsed) > 1 {
		dependency.UseSSH = parsed[1] == consts.GitProtocolSSH
	}
	return dependency
}

// GetDependencies converts a map of dependencies to a slice of Dependency objects
func GetDependencies(deps map[string]string) []Dependency {
	dependencies := make([]Dependency, 0)
	for repo, info := range deps {
		dependencies = append(dependencies, ParseDependency(repo, info))
	}
	return dependencies
}

// GetDependenciesNames returns a slice of dependency names
func GetDependenciesNames(deps []Dependency) []string {
	var dependencies []string
	for _, info := range deps {
		dependencies = append(dependencies, info.Name())
	}
	return dependencies
}

// Name returns the name of the dependency extracted from the repository URL
func (p *Dependency) Name() string {
	return reDepName.FindString(p.Repository)
}

// GetKey returns the normalized key for the dependency (lowercase repository).
func (p *Dependency) GetKey() string {
	return strings.ToLower(p.Repository)
}

// NeedsVersionUpdate checks if a version update is needed based on semver comparison.
func NeedsVersionUpdate(currentVersion, newVersion string) bool {
	parsedNew, err := semver.NewVersion(newVersion)
	if err != nil {
		return newVersion != currentVersion
	}

	parsedCurrent, err := semver.NewVersion(currentVersion)
	if err != nil {
		return newVersion != currentVersion
	}

	return parsedNew.GreaterThan(parsedCurrent)
}
