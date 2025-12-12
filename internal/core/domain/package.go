package domain

import (
	"encoding/json"
	"fmt"
	"strings"

	fs "github.com/hashload/boss/internal/adapters/secondary/filesystem"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/utils/parser"
)

type Package struct {
	fileName     string
	fs           fs.FileSystem
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

// Save persists the package to disk and returns the marshaled bytes.
func (p *Package) Save() []byte {
	marshal, _ := parser.JSONMarshal(p, true)
	_ = p.getFS().WriteFile(p.fileName, marshal, 0600)
	p.Lock.Save()
	return marshal
}

// getFS returns the filesystem to use, defaulting to fs.Default.
func (p *Package) getFS() fs.FileSystem {
	if p.fs == nil {
		return fs.Default
	}
	return p.fs
}

// SetFS sets the filesystem implementation for testing.
func (p *Package) SetFS(filesystem fs.FileSystem) {
	p.fs = filesystem
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
	if p == nil || len(p.Dependencies) == 0 {
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

func getNewWithFS(file string, filesystem fs.FileSystem) *Package {
	res := new(Package)
	res.fileName = file
	res.fs = filesystem

	res.Dependencies = make(map[string]string)
	res.Projects = []string{}
	res.Lock = LoadPackageLockWithFS(res, filesystem)
	return res
}

// LoadPackage loads the package from the default boss file location.
func LoadPackage(createNew bool) (*Package, error) {
	return LoadPackageWithFS(createNew, fs.Default)
}

// LoadPackageWithFS loads the package using the specified filesystem.
func LoadPackageWithFS(createNew bool, filesystem fs.FileSystem) (*Package, error) {
	fileBytes, err := filesystem.ReadFile(env.GetBossFile())
	if err != nil {
		if createNew {
			err = nil
		}
		return getNewWithFS(env.GetBossFile(), filesystem), err
	}
	result := getNewWithFS(env.GetBossFile(), filesystem)

	if err := json.Unmarshal(fileBytes, result); err != nil {
		if !filesystem.Exists(env.GetBossFile()) {
			return nil, err
		}

		return nil, fmt.Errorf("error on unmarshal file %s: %w", env.GetBossFile(), err)
	}
	result.Lock = LoadPackageLockWithFS(result, filesystem)
	return result, nil
}

// LoadPackageOther loads a package from a specified path.
func LoadPackageOther(path string) (*Package, error) {
	return LoadPackageOtherWithFS(path, fs.Default)
}

// LoadPackageOtherWithFS loads a package from a specified path using the given filesystem.
func LoadPackageOtherWithFS(path string, filesystem fs.FileSystem) (*Package, error) {
	fileBytes, err := filesystem.ReadFile(path)
	if err != nil {
		return getNewWithFS(path, filesystem), err
	}

	result := getNewWithFS(path, filesystem)

	err = json.Unmarshal(fileBytes, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
