package models

import (
	"bytes"
	"encoding/json"
	. "io/ioutil"

	"github.com/hashload/boss/consts"
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
}

func JSONMarshal(v interface{}, safeEncoding bool) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "\t")

	if safeEncoding {
		b = bytes.Replace(b, []byte("\\u003c"), []byte("<"), -1)
		b = bytes.Replace(b, []byte("\\u003e"), []byte(">"), -1)
		b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)
	}
	return b, err
}

func (p *Package) Save() {
	marshal, _ := JSONMarshal(p, true)

	_ = WriteFile(p.fileName, marshal, 664)
}

func (p *Package) AddDependency(dep string, ver string) {
	if p.Dependencies == nil {
		p.Dependencies = make(map[string]interface{})
	}
	deps := p.Dependencies.(map[string]interface{})
	deps[dep] = ver
}

func (p *Package) AddProject(project string) {
	p.Projects = append(p.Projects, project)
}

func (p *Package) RemoveDependency(dep string) {
	if p.Dependencies != nil {
		deps := p.Dependencies.(map[string]interface{})
		if i := deps[dep]; i != nil {
			delete(deps, dep)
		}
	}
}

func getNew(file string) *Package {
	res := new(Package)
	res.fileName = file
	res.IsNew = true

	res.Dependencies = make(map[string]interface{})
	res.Projects = []string{}

	return res
}

func LoadPackage(createNew bool) (*Package, error) {
	if fileBytes, e := ReadFile(consts.FilePackage); e != nil {
		if createNew {
			e = nil
		}
		return getNew(consts.FilePackage), e
	} else {
		result := getNew(consts.FilePackage)
		if err := json.Unmarshal(fileBytes, result); err != nil {
			return nil, e
		}
		result.IsNew = false
		return result, nil
	}
}

func LoadPackageOther(path string) (*Package, error) {
	if fileBytes, e := ReadFile(path); e != nil {
		return nil, e
	} else {
		result := getNew(path)
		if err := json.Unmarshal(fileBytes, result); err != nil {
			return nil, e
		}
		return result, nil
	}
}
