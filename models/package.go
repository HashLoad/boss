package models

import (
	"encoding/json"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/utils/parser"
	. "io/ioutil"
	"strings"
)

type Package struct {
	fileName     string
	IsNew        bool        `json:"-"`
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	Version      string      `json:"version"`
	Homepage     string      `json:"homepage"`
	MainSrc      string      `json:"mainsrc"`
	Projects     []string    `json:"projects"`
	Scripts      interface{} `json:"scripts,omitempty"`
	Dependencies interface{} `json:"dependencies"`
	Lock         PackageLock `json:"-"`
}

func (p *Package) Save() {
	marshal, _ := parser.JSONMarshal(p, true)

	_ = WriteFile(p.fileName, marshal, 664)

	p.Lock.Save()
}

func (p *Package) AddDependency(dep string, ver string) {
	if p.Dependencies == nil {
		p.Dependencies = make(map[string]interface{})
	}
	deps := p.Dependencies.(map[string]interface{})

	for key := range deps {
		if strings.ToLower(key) == strings.ToLower(dep) {
			deps[key] = ver
			return
		}
	}

	deps[dep] = ver
}

func (p *Package) AddProject(project string) {
	p.Projects = append(p.Projects, project)
}

func (p *Package) GetParsedDependencies() []Dependency {
	dependencies := p.Dependencies.(map[string]interface{})
	if len(dependencies) == 0 {
		return []Dependency{}
	}
	return GetDependencies(dependencies)
}

func (p *Package) UninstallDependency(dep string) {
	if p.Dependencies != nil {
		deps := p.Dependencies.(map[string]interface{})
		for key := range deps {
			if strings.ToLower(key) == strings.ToLower(dep) {
				delete(deps, key)
			}
		}
		p.Dependencies = deps
	}
}

func getNew(file string) *Package {
	res := new(Package)
	res.fileName = file
	res.IsNew = true

	res.Dependencies = make(map[string]interface{})
	res.Projects = []string{}
	res.Lock = LoadPackageLock(res)
	return res
}

func LoadPackage(createNew bool) (*Package, error) {
	if fileBytes, e := ReadFile(env.GetBossFile()); e != nil {
		if createNew {
			e = nil
		}
		return getNew(env.GetBossFile()), e
	} else {
		result := getNew(env.GetBossFile())
		if err := json.Unmarshal(fileBytes, result); err != nil {
			return nil, err
		}
		result.Lock = LoadPackageLock(result)
		result.IsNew = false
		return result, nil
	}
}

func LoadPackageOther(path string) (*Package, error) {
	if fileBytes, e := ReadFile(path); e != nil {
		return getNew(path), e
	} else {
		result := getNew(path)
		if err := json.Unmarshal(fileBytes, result); err != nil {
			return nil, e
		}
		return result, nil
	}
}
