package models

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/utils/parser"
)

type Package struct {
	fileName     string
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Version      string            `json:"version"`
	Homepage     string            `json:"homepage"`
	MainSrc      string            `json:"mainsrc"`
	BrowsingPath string            `json:"browsingpath"`
	Projects     []string          `json:"projects"`
	Scripts      map[string]string `json:"scripts,omitempty"`
	Dependencies map[string]string `json:"dependencies"`
	Lock         PackageLock       `json:"-"`
}

func (p *Package) Save() []byte {
	marshal, _ := parser.JSONMarshal(p, true)
	_ = os.WriteFile(p.fileName, marshal, 0600)
	p.Lock.Save()
	return marshal
}

func (p *Package) AddDependency(dep string, ver string) {
	for key := range p.Dependencies {
		if strings.EqualFold(key, dep) {
			p.Dependencies[key] = ver
			return
		}
	}

	p.Dependencies[dep] = ver
}

func (p *Package) AddProject(project string) {
	p.Projects = append(p.Projects, project)
}

func (p *Package) GetParsedDependencies() []Dependency {
	if len(p.Dependencies) == 0 {
		return []Dependency{}
	}
	return GetDependencies(p.Dependencies)
}

func (p *Package) UninstallDependency(dep string) {
	if p.Dependencies != nil {
		for key := range p.Dependencies {
			if strings.EqualFold(key, dep) {
				delete(p.Dependencies, key)
				return
			}
		}
	}
}

func getNew(file string) *Package {
	res := new(Package)
	res.fileName = file

	res.Dependencies = make(map[string]string)
	res.Projects = []string{}
	res.Lock = LoadPackageLock(res)
	return res
}

func LoadPackage(createNew bool) (*Package, error) {
	fileBytes, err := os.ReadFile(env.GetBossFile())
	if err != nil {
		if createNew {
			err = nil
		}
		return getNew(env.GetBossFile()), err
	}
	result := getNew(env.GetBossFile())
	if err := json.Unmarshal(fileBytes, result); err != nil {
		return nil, err
	}
	result.Lock = LoadPackageLock(result)
	return result, nil
}

func LoadPackageOther(path string) (*Package, error) {
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return getNew(path), err
	}

	result := getNew(path)

	err = json.Unmarshal(fileBytes, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
