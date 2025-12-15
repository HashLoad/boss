package domain

import (
	"strings"

	"github.com/hashload/boss/utils"
)

// DependencyArtifacts holds the compiled artifacts for a dependency.
type DependencyArtifacts struct {
	Bin []string `json:"bin,omitempty"`
	Dcp []string `json:"dcp,omitempty"`
	Dcu []string `json:"dcu,omitempty"`
	Bpl []string `json:"bpl,omitempty"`
}

// LockedDependency represents a locked dependency in the lock file.
type LockedDependency struct {
	Name      string              `json:"name"`
	Version   string              `json:"version"`
	Hash      string              `json:"hash"`
	Artifacts DependencyArtifacts `json:"artifacts"`
	Failed    bool                `json:"-"`
	Changed   bool                `json:"-"`
}

// PackageLock represents the lock file for a package.
// This is a pure domain entity. Use LockRepository for persistence.
type PackageLock struct {
	Hash      string                      `json:"hash"`
	Updated   string                      `json:"updated"` // ISO 8601 timestamp
	Installed map[string]LockedDependency `json:"installedModules"`
}

// AddDependency adds a dependency to the lock without performing I/O.
// The hash must be pre-calculated and passed as a parameter.
func (p *PackageLock) AddDependency(dep Dependency, version, hash string) {
	key := dep.GetKey()
	if locked, ok := p.Installed[key]; !ok {
		p.Installed[key] = LockedDependency{
			Name:    dep.Name(),
			Version: version,
			Changed: true,
			Hash:    hash,
			Artifacts: DependencyArtifacts{
				Bin: []string{},
				Bpl: []string{},
				Dcp: []string{},
				Dcu: []string{},
			},
		}
	} else {
		locked.Version = version
		locked.Hash = hash
		p.Installed[key] = locked
	}
}

// GetInstalled returns the locked dependency for the given dependency.
func (p *PackageLock) GetInstalled(dep Dependency) LockedDependency {
	return p.Installed[dep.GetKey()]
}

// SetInstalled sets a locked dependency without performing any I/O operations.
func (p *PackageLock) SetInstalled(dep Dependency, locked LockedDependency) {
	p.Installed[dep.GetKey()] = locked
}

// CleanRemoved removes dependencies that are no longer in the dependency list.
func (p *PackageLock) CleanRemoved(deps []Dependency) {
	var repositories []string
	for _, dep := range deps {
		repositories = append(repositories, dep.GetKey())
	}

	for key := range p.Installed {
		if !utils.Contains(repositories, strings.ToLower(key)) {
			delete(p.Installed, key)
		}
	}
}

// GetArtifactList returns all artifacts from all installed dependencies.
func (p *PackageLock) GetArtifactList() []string {
	var result []string
	for _, installed := range p.Installed {
		result = append(result, installed.GetArtifacts()...)
	}
	return result
}

// Clean clears all artifacts.
func (p *DependencyArtifacts) Clean() {
	p.Bin = []string{}
	p.Bpl = []string{}
	p.Dcp = []string{}
	p.Dcu = []string{}
}

// GetArtifacts returns all artifacts as a single slice.
func (p *LockedDependency) GetArtifacts() []string {
	var result []string
	result = append(result, p.Artifacts.Dcp...)
	result = append(result, p.Artifacts.Dcu...)
	result = append(result, p.Artifacts.Bin...)
	result = append(result, p.Artifacts.Bpl...)
	return result
}
