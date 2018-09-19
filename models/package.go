package models

import (
	"encoding/json"
	. "io/ioutil"
	"time"
)

type Package struct {
	fileName string

	Name            *string     `json:"name"`
	Version         *string     `json:"version"`
	Private         bool        `json:"private"`
	Homepage        string      `json:"homepage"`
	MainSrc         *string     `json:"mainsrc"`
	Supported       string      `json:"supported"`
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
	deps := p.Dependencies.(map[string]interface{})
	deps[dep] = ver
}

func (p *Package) AddDevDependency(dep string, ver string) {
	deps := p.DevDependencies.(map[string]interface{})
	deps[dep] = ver
}

func (p *Package) RemoveDependency(dep string) {
	deps := p.Dependencies.(map[string]interface{})
	if i := deps[dep]; i != nil {
		delete(deps, dep)
	}

	depsDev := p.DevDependencies.(map[string]interface{})
	if i := depsDev[dep]; i != nil {
		delete(depsDev, dep)
	}
}

func LoadPackage(file string) (*Package, error) {
	if bytes, e := ReadFile(file); e != nil {
		return nil, e
	} else {
		result := &Package{}
		result.fileName = file
		if err := json.Unmarshal(bytes, result); err != nil {
			return nil, e
		}
		return result, nil
	}
}
