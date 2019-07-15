package models

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"github.com/masterminds/semver"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type DependencyArtifacts struct {
	Bin []string `json:"bin,omitempty"`
	Dcp []string `json:"dcp,omitempty"`
	Dcu []string `json:"dcu,omitempty"`
	Bpl []string `json:"bpl,omitempty"`
}

type LockedDependency struct {
	Name      string              `json:"name"`
	Version   string              `json:"version"`
	Hash      string              `json:"hash"`
	Artifacts DependencyArtifacts `json:"artifacts"`
	Failed    bool                `json:"failed"`
	Changed   bool                `json:"changed"`
}

type PackageLock struct {
	fileName  string
	Hash      string                      `json:"hash"`
	Updated   time.Time                   `json:"updated"`
	Installed map[string]LockedDependency `json:"installedModules"`
}

func removeOld(parentPackage *Package) {
	var oldFileName = filepath.Join(filepath.Dir(parentPackage.fileName), consts.FilePackageLockOld)
	var newFileName = filepath.Join(filepath.Dir(parentPackage.fileName), consts.FilePackageLock)
	if _, err := os.Stat(oldFileName); err == nil {
		err := os.Rename(oldFileName, newFileName)
		utils.HandleError(err)
	}
}

func LoadPackageLock(parentPackage *Package) PackageLock {
	removeOld(parentPackage)
	packageLockPath := filepath.Join(filepath.Dir(parentPackage.fileName), consts.FilePackageLock)
	if fileBytes, err := ioutil.ReadFile(packageLockPath); err != nil {
		hash := md5.New()
		if _, err := io.WriteString(hash, parentPackage.Name); err != nil {
			msg.Warn("Failed on  write machine id to hash")
		}

		return PackageLock{
			fileName: packageLockPath,
			Updated:  time.Now(),
			Hash:     hex.EncodeToString(hash.Sum(nil)),

			Installed: map[string]LockedDependency{},
		}
	} else {
		lockfile := PackageLock{
			fileName:  packageLockPath,
			Updated:   time.Now(),
			Installed: map[string]LockedDependency{},
		}

		if err := json.Unmarshal(fileBytes, &lockfile); err != nil {
			utils.HandleError(err)
		}
		return lockfile
	}
}

func (p PackageLock) Save() {
	marshal, err := json.MarshalIndent(&p, "", "\t")
	if err != nil {
		log.Fatalf("error %v", err)
	}

	_ = ioutil.WriteFile(p.fileName, marshal, 664)
}

func (p PackageLock) AddInstalled(dep Dependency, version string) {

	dependencyDir := filepath.Join(env.GetCurrentDir(), consts.FolderDependencies, dep.GetName())

	hash := utils.HashDir(dependencyDir)

	if locked, ok := p.Installed[strings.ToLower(dep.Repository)]; !ok {
		p.Installed[strings.ToLower(dep.Repository)] = LockedDependency{
			Name:    dep.GetName(),
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
		p.Installed[strings.ToLower(dep.Repository)] = locked
	}
}

func (p Dependency) internalNeedUpdate(lockedDependency LockedDependency, version string) bool {
	if lockedDependency.Failed {
		return true
	}

	dependencyDir := filepath.Join(env.GetCurrentDir(), consts.FolderDependencies, p.GetName())

	if _, err := os.Stat(dependencyDir); os.IsNotExist(err) {
		return true
	}
	hash := utils.HashDir(dependencyDir)

	if lockedDependency.Hash != hash {
		return true
	}

	parsedNewVersion, err := semver.NewVersion(version)
	if err != nil {
		return version != lockedDependency.Version
	}

	parsedVersion, err := semver.NewVersion(lockedDependency.Version)
	if err != nil {
		return version != lockedDependency.Version
	}
	return parsedNewVersion.GreaterThan(parsedVersion)

}

func (p *DependencyArtifacts) Clean() {
	p.Bin = []string{}
	p.Bpl = []string{}
	p.Dcp = []string{}
	p.Dcu = []string{}
}
func (p LockedDependency) checkArtifactsType(directory string, artifacts []string) bool {
	for _, value := range artifacts {
		bpl := filepath.Join(directory, value)
		_, err := os.Stat(bpl)
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func (p LockedDependency) checkArtifacts(lock PackageLock) bool {
	baseModulesDir := filepath.Join(filepath.Dir(lock.fileName), consts.FolderDependencies)

	if !p.checkArtifactsType(filepath.Join(baseModulesDir, consts.BplFolder), p.Artifacts.Bpl) {
		return false
	}

	if !p.checkArtifactsType(filepath.Join(baseModulesDir, consts.BinFolder), p.Artifacts.Bin) {
		return false
	}

	if !p.checkArtifactsType(filepath.Join(baseModulesDir, consts.DcpFolder), p.Artifacts.Dcp) {
		return false
	}

	if !p.checkArtifactsType(filepath.Join(baseModulesDir, consts.DcuFolder), p.Artifacts.Dcu) {
		return false
	}

	return true
}

func (p PackageLock) NeedUpdate(dep Dependency, version string) bool {
	if lockedDependency, ok := p.Installed[strings.ToLower(dep.Repository)]; !ok {
		return true
	} else {
		needUpdate := dep.internalNeedUpdate(lockedDependency, version) || !lockedDependency.checkArtifacts(p)
		lockedDependency.Changed = needUpdate || lockedDependency.Changed

		if lockedDependency.Changed {
			lockedDependency.Failed = false
			lockedDependency.Artifacts.Bin = []string{}
			lockedDependency.Artifacts.Bpl = []string{}
			lockedDependency.Artifacts.Dcp = []string{}
			lockedDependency.Artifacts.Dcu = []string{}
		}
		p.Installed[strings.ToLower(dep.Repository)] = lockedDependency
		return needUpdate
	}
}

func (p PackageLock) GetInstalled(dep Dependency) LockedDependency {
	return p.Installed[strings.ToLower(dep.Repository)]
}

func (p PackageLock) SetInstalled(dep Dependency, locked LockedDependency) {
	dependencyDir := filepath.Join(env.GetCurrentDir(), consts.FolderDependencies, dep.GetName())
	hash := utils.HashDir(dependencyDir)
	locked.Hash = hash

	p.Installed[strings.ToLower(dep.Repository)] = locked
}

func (p PackageLock) CleanRemoved(deps []Dependency) {
	var repositories []string
	for _, dep := range deps {
		repositories = append(repositories, strings.ToLower(dep.Repository))
	}

	for key := range p.Installed {
		if !utils.Contains(repositories, strings.ToLower(key)) {
			delete(p.Installed, key)
		}
	}
}

func (p PackageLock) GetArtifactList() []string {
	var result []string

	for _, installed := range p.Installed {
		result = append(result, installed.GetArtifacts()...)
	}
	return result
}

func (p LockedDependency) GetArtifacts() []string {
	var result []string
	result = append(result, p.Artifacts.Dcp...)
	result = append(result, p.Artifacts.Dcu...)
	result = append(result, p.Artifacts.Bin...)
	result = append(result, p.Artifacts.Bpl...)
	return result
}
