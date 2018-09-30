package models

import (
	"encoding/json"
	"github.com/hashload/boss/consts"
	. "io/ioutil"
	"time"
)

type Package struct {
	fileName        string
	IsNew           bool
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	Version         string      `json:"version"`
	Private         bool        `json:"private"`
	Homepage        string      `json:"homepage"`
	MainSrc         string      `json:"mainsrc"`
	Supported       string      `json:"supported"`
	DprojName       string      `json:"dprojFile"`
	ModifyedAt      time.Time   `json:"modifyedAt"`
	Scripts         interface{} `json:"scripts"`
	Dependencies    interface{} `json:"dependencies"`
	DevDependencies interface{} `json:"devDependencies"`
}

func (p *Package) updateTime() {
	p.ModifyedAt = time.Now()
}

func (p *Package) Save() {
	p.updateTime()
	marshal, _ := json.MarshalIndent(p, "", "\t")
	WriteFile(p.fileName, marshal, 664)
}

func (p *Package) AddDependency(dep string, ver string) {
	if p.Dependencies == nil {
		p.Dependencies = make(map[string]interface{})
	}
	deps := p.Dependencies.(map[string]interface{})
	deps[dep] = ver
}

func (p *Package) AddDevDependency(dep string, ver string) {
	if p.DevDependencies == nil {
		p.DevDependencies = make(map[string]interface{})
	}
	deps := p.DevDependencies.(map[string]interface{})
	deps[dep] = ver
}

func (p *Package) RemoveDependency(dep string) {
	if p.Dependencies != nil {
		deps := p.Dependencies.(map[string]interface{})
		if i := deps[dep]; i != nil {
			delete(deps, dep)
		}
	}

	if p.DevDependencies != nil {
		depsDev := p.DevDependencies.(map[string]interface{})
		if i := depsDev[dep]; i != nil {
			delete(depsDev, dep)
		}
	}
}

func getNew(file string) *Package {
	res := new(Package)
	/*
	    ALL PROPS CREATION
		res.DevDependencies = make(map[string]interface{})
		res.Dependencies = make(map[string]interface{})
		res.Scripts = make(map[string]interface{})
	*/
	res.fileName = file
	res.IsNew = true

	return res
}

func LoadPackage(createNew bool) (*Package, error) {
	if bytes, e := ReadFile(consts.FILE_PACKAGE); e != nil {
		if createNew {
			e = nil
		}
		return getNew(consts.FILE_PACKAGE), e
	} else {
		result := getNew(consts.FILE_PACKAGE)
		if err := json.Unmarshal(bytes, result); err != nil {
			return nil, e
		}
		result.IsNew = false
		return result, nil
	}
}
