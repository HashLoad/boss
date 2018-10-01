package models

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"regexp"
	"strings"
)

type Dependency struct {
	Repository string
	Version    string
	UseSSH     bool
}

func (d *Dependency) GetHashName() string {
	hash := md5.New()
	io.WriteString(hash, d.Repository)
	return hex.EncodeToString(hash.Sum(nil))
}

func (d *Dependency) GetURL() string {
	return "https://" + d.Repository + ".git"
}

func ParseDependency(repo string, info string) Dependency {
	parsed := strings.Split(info, ":")
	dependency := Dependency{}
	dependency.Repository = repo
	dependency.Version = parsed[0]
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
	var re = regexp.MustCompile(`(?m)\w+$`)
	return re.FindString(d.Repository)
}
