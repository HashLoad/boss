package domain

import (
	//nolint:gosec // We are not using this for security purposes
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashload/boss/internal/infra"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/msg"
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
type PackageLock struct {
	fileName  string
	fs        infra.FileSystem
	Hash      string                      `json:"hash"`
	Updated   time.Time                   `json:"updated"`
	Installed map[string]LockedDependency `json:"installedModules"`
}

// SetFS sets the filesystem implementation for testing.
func (p *PackageLock) SetFS(filesystem infra.FileSystem) {
	p.fs = filesystem
}

// GetFileName returns the lock file path.
func (p *PackageLock) GetFileName() string {
	return p.fileName
}

func removeOldWithFS(parentPackage *Package, filesystem infra.FileSystem) {
	oldFileName := filepath.Join(filepath.Dir(parentPackage.fileName), consts.FilePackageLockOld)
	newFileName := filepath.Join(filepath.Dir(parentPackage.fileName), consts.FilePackageLock)
	if filesystem.Exists(oldFileName) {
		err := filesystem.Rename(oldFileName, newFileName)
		if err != nil {
			msg.Warn("⚠️ Failed to rename old lock file: %v", err)
		}
	}
}

// LoadPackageLockWithFS loads the package lock file using the specified filesystem.
func LoadPackageLockWithFS(parentPackage *Package, filesystem infra.FileSystem) PackageLock {
	removeOldWithFS(parentPackage, filesystem)
	packageLockPath := filepath.Join(filepath.Dir(parentPackage.fileName), consts.FilePackageLock)
	fileBytes, err := filesystem.ReadFile(packageLockPath)
	if err != nil {
		//nolint:gosec // We are not using this for security purposes
		hash := md5.New()
		if _, err := io.WriteString(hash, parentPackage.Name); err != nil {
			msg.Warn("⚠️ Failed on write machine id to hash")
		}

		return PackageLock{
			fileName:  packageLockPath,
			fs:        filesystem,
			Updated:   time.Now(),
			Hash:      hex.EncodeToString(hash.Sum(nil)),
			Installed: map[string]LockedDependency{},
		}
	}

	lockfile := PackageLock{
		fileName:  packageLockPath,
		fs:        filesystem,
		Updated:   time.Now(),
		Installed: map[string]LockedDependency{},
	}

	if err := json.Unmarshal(fileBytes, &lockfile); err != nil {
		msg.Die("❌ Error parsing lock file %s: %s", packageLockPath, err.Error())
	}
	return lockfile
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

// CheckArtifactsExist verifies if all artifacts exist in the given directory.
func (p *LockedDependency) CheckArtifactsExist(directory string, artifacts []string, fs infra.FileSystem) bool {
	for _, artifact := range artifacts {
		path := filepath.Join(directory, artifact)
		if !fs.Exists(path) {
			return false
		}
	}
	return true
}
