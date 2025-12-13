package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/hashload/boss/internal/infra"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/utils/parser"
)

// defaultFS holds the default filesystem implementation.
// This is set by the infrastructure layer during application bootstrap.
//
//nolint:gochecknoglobals // Required for backward compatibility
var (
	defaultFS   infra.FileSystem
	defaultFSMu sync.RWMutex
)

// SetDefaultFS sets the default filesystem implementation.
// This should be called during application initialization.
func SetDefaultFS(fs infra.FileSystem) {
	defaultFSMu.Lock()
	defer defaultFSMu.Unlock()
	defaultFS = fs
}

// GetDefaultFS returns the default filesystem implementation.
// If no filesystem was set, it returns nil (caller should handle this).
func GetDefaultFS() infra.FileSystem {
	defaultFSMu.RLock()
	defer defaultFSMu.RUnlock()
	return defaultFS
}

// getOrCreateDefaultFS returns the default filesystem or creates a new ErrorFileSystem.
// This provides lazy initialization for tests and backward compatibility.
func getOrCreateDefaultFS() infra.FileSystem {
	defaultFSMu.RLock()
	fs := defaultFS
	defaultFSMu.RUnlock()

	if fs != nil {
		return fs
	}

	// Lazy initialization - return error filesystem to prevent implicit I/O
	return infra.NewErrorFileSystem()
}

type Package struct {
	fileName     string
	fs           infra.FileSystem
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

type PackageEngines struct {
	Delphi    string   `json:"delphi,omitempty"`
	Compiler  string   `json:"compiler,omitempty"`
	Platforms []string `json:"platforms,omitempty"`
}

type PackageToolchain struct {
	Delphi   string `json:"delphi,omitempty"`
	Compiler string `json:"compiler,omitempty"`
	Platform string `json:"platform,omitempty"`
	Path     string `json:"path,omitempty"`
	Strict   bool   `json:"strict,omitempty"`
}

// NewPackage creates a new Package with the given file path.
func NewPackage(filePath string) *Package {
	return &Package{
		fileName:     filePath,
		Dependencies: make(map[string]string),
		Projects:     []string{},
	}
}

// Save persists the package to disk and returns the marshaled bytes.
// Note: This method only saves the package file, not the lock file.
// Use lock.Service.Save() to persist the lock file separately.
func (p *Package) Save() []byte {
	marshal, _ := parser.JSONMarshal(p, true)
	_ = p.getFS().WriteFile(p.fileName, marshal, 0600)
	return marshal
}

// getFS returns the filesystem to use, defaulting to getOrCreateDefaultFS.
func (p *Package) getFS() infra.FileSystem {
	if p.fs == nil {
		return getOrCreateDefaultFS()
	}
	return p.fs
}

// SetFS sets the filesystem implementation for testing.
func (p *Package) SetFS(filesystem infra.FileSystem) {
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

func getNewWithFS(file string, filesystem infra.FileSystem) *Package {
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
	return LoadPackageWithFS(createNew, getOrCreateDefaultFS())
}

// LoadPackageWithFS loads the package using the specified filesystem.
func LoadPackageWithFS(createNew bool, filesystem infra.FileSystem) (*Package, error) {
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
	return LoadPackageOtherWithFS(path, getOrCreateDefaultFS())
}

// LoadPackageOtherWithFS loads a package from a specified path using the given filesystem.
func LoadPackageOtherWithFS(path string, filesystem infra.FileSystem) (*Package, error) {
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
