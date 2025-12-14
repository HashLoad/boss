// Package domain contains the core business entities for Boss dependency manager.
// It defines Package, Dependency, Lock file structures and their associated operations.
package domain

import (
	"strings"
)

// Package represents the boss.json file structure.
// This is a pure domain entity containing only business data and logic.
// Use PackageRepository (ports.PackageRepository) for persistence operations.
type Package struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Version      string            `json:"version"`
	Homepage     string            `json:"homepage"`
	MainSrc      string            `json:"mainsrc"`
	BrowsingPath string            `json:"browsingpath"`
	Projects     []string          `json:"projects"`
	Scripts      map[string]string `json:"scripts,omitempty"`
	Dependencies map[string]string `json:"dependencies"`
	Engines      *PackageEngines   `json:"engines,omitempty"`
	Toolchain    *PackageToolchain `json:"toolchain,omitempty"`
	Lock         PackageLock       `json:"-"`
}

// PackageEngines represents the engines configuration in boss.json.
type PackageEngines struct {
	Delphi    string   `json:"delphi,omitempty"`
	Compiler  string   `json:"compiler,omitempty"`
	Platforms []string `json:"platforms,omitempty"`
}

// PackageToolchain represents the toolchain configuration in boss.json.
type PackageToolchain struct {
	Delphi   string `json:"delphi,omitempty"`
	Compiler string `json:"compiler,omitempty"`
	Platform string `json:"platform,omitempty"`
	Path     string `json:"path,omitempty"`
	Strict   bool   `json:"strict,omitempty"`
}

// NewPackage creates a new Package with initialized collections.
func NewPackage() *Package {
	return &Package{
		Dependencies: make(map[string]string),
		Projects:     []string{},
	}
}

// AddDependency adds or updates a dependency in the package.
func (p *Package) AddDependency(dep string, ver string) {
	for key := range p.Dependencies {
		if strings.EqualFold(key, dep) {
			p.Dependencies[key] = ver
			return
		}
	}

	p.Dependencies[dep] = ver
}

// AddProject adds a project to the package.
func (p *Package) AddProject(project string) {
	p.Projects = append(p.Projects, project)
}

// GetParsedDependencies returns the dependencies parsed as Dependency objects.
func (p *Package) GetParsedDependencies() []Dependency {
	if p == nil || len(p.Dependencies) == 0 {
		return []Dependency{}
	}
	return GetDependencies(p.Dependencies)
}

// UninstallDependency removes a dependency from the package.
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
