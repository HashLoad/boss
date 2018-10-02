package models

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"regexp"
	"strings"

	"github.com/hashload/boss/msg"
)

type Dependency struct {
	Repository string
	version    string
	UseSSH     bool
}

func (d *Dependency) GetHashName() string {
	hash := md5.New()
	io.WriteString(hash, d.Repository)
	return hex.EncodeToString(hash.Sum(nil))
}

func (d *Dependency) GetVersion() string {
	return d.version
}

func (d *Dependency) makeSshUrl() string {
	re = regexp.MustCompile(`(?m)([\w\d.]*)(?:\/)(.*)`)
	submatch := re.FindStringSubmatch(d.Repository)
	provider := submatch[1]
	repo := submatch[2]
	return "git@" + provider + ":" + repo + ".git"
}

func (d *Dependency) GetURLPrefix() string {
	var re = regexp.MustCompile(`(?m)(\w+\.\w+)`)
	return re.FindString(d.Repository)
}

func (d *Dependency) GetURL() string {
	prefix := d.GetURLPrefix()
	auth := GlobalConfiguration.Auth[prefix]
	if auth != nil {
		if auth.UseSsh {
			return d.makeSshUrl()
		}
	}
	return "https://" + d.Repository + ".git"
}

var re = regexp.MustCompile(`(?m)^(.|)(\d+)\.(\d+)$`)
var re2 = regexp.MustCompile(`(?m)^(.|)(\d+)$`)

func ParseDependency(repo string, info string) Dependency {
	parsed := strings.Split(info, ":")
	dependency := Dependency{}
	dependency.Repository = repo
	dependency.version = parsed[0]
	if dependency.version == "x" {
		dependency.version = "> 0.0.0"
	}
	if re.MatchString(dependency.version) {
		msg.Warn("Current version for %s is not semantic (x.y.z), for comparation using %s -> %s",
			dependency.Repository, dependency.version, dependency.version+".0")
		dependency.version = dependency.version + ".0"
	}
	if re2.MatchString(dependency.version) {
		msg.Warn("Current version for %s is not semantic (x.y.z), for comparation using %s -> %s",
			dependency.Repository, dependency.version, dependency.version+".0.0/")
		dependency.version = dependency.version + ".0.0"
	}
	if len(parsed) > 1 {
		dependency.UseSSH = parsed[1] == "ssh"
	}
	return dependency
}

func GetDependencies(deps map[string]interface{}) []Dependency {
	dependencies := make([]Dependency, 0)
	for repo, info := range deps {
		dependencies = append(dependencies, ParseDependency(repo, info.(string)))
	}
	return dependencies
}

func (d *Dependency) GetName() string {
	var re = regexp.MustCompile(`[^\/]+(:?\/$|$)`)
	return re.FindString(d.Repository)
}
