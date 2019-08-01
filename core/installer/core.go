package installer

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/core/compiler"
	"github.com/hashload/boss/core/gitWrapper"
	"github.com/hashload/boss/core/paths"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"github.com/hashload/boss/utils/librarypath"
	"github.com/masterminds/semver"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"io/ioutil"
	"os"
	"path/filepath"
)

var processed = []string{consts.BplFolder, consts.BinFolder, consts.DcpFolder, consts.DcuFolder}

func DoInstall(pkg *models.Package, lockedVersion bool) {
	msg.Info("Installing modules in project path")

	dependencies := EnsureDependencies(pkg.Lock, pkg, lockedVersion)
	paths.EnsureCleanModulesDir(dependencies, pkg.Lock)

	pkg.Lock.CleanRemoved(dependencies)
	pkg.Save()

	librarypath.UpdateLibraryPath(pkg)
	msg.Info("Compiling units")
	compiler.Build(pkg)
	pkg.Save()
	msg.Info("Success!")
}

func EnsureDependencies(rootLock models.PackageLock, pkg *models.Package, lockedVersion bool) []models.Dependency {
	if pkg.Dependencies == nil {
		return []models.Dependency{}
	}
	deps := pkg.GetParsedDependencies()

	makeCache(deps)

	ensureModules(rootLock, pkg, deps, lockedVersion)

	deps = append(deps, processOthers(rootLock, lockedVersion)...)

	return deps
}

func makeCache(deps []models.Dependency) {
	msg.Info("Building cache files...")
	for _, dep := range deps {
		GetDependency(dep)
	}
}

func processOthers(rootLock models.PackageLock, lockedVersion bool) []models.Dependency {
	infos, e := ioutil.ReadDir(env.GetModulesDir())
	var lenProcessedInitial = len(processed)
	var result []models.Dependency
	if e != nil {
		msg.Err("Error on try load dir of modules: %s", e)
	}

	for _, info := range infos {
		if !info.IsDir() {
			continue
		}

		if utils.Contains(processed, info.Name()) {
			continue
		} else {
			processed = append(processed, info.Name())
		}
		msg.Info("Processing module %s", info.Name())

		fileName := filepath.Join(env.GetModulesDir(), info.Name(), consts.FilePackage)

		_, i := os.Stat(fileName)
		if os.IsNotExist(i) {
			msg.Warn("  boss.json not exists in %s", info.Name())
		}

		if packageOther, e := models.LoadPackageOther(fileName); e != nil {
			if os.IsNotExist(e) {
				continue
			}
			msg.Err("  Error on try load package %s: %s", fileName, e)
		} else {
			result = append(result, EnsureDependencies(rootLock, packageOther, lockedVersion)...)
		}
	}
	if lenProcessedInitial > len(processed) {
		result = append(result, processOthers(rootLock, lockedVersion)...)
	}

	return result
}

func ensureModules(rootLock models.PackageLock, pkg *models.Package, deps []models.Dependency, lockedVersion bool) {
	msg.Info("Installing modules")
	for _, dep := range deps {
		msg.Info("Processing dependency %s", dep.GetName())
		repository := gitWrapper.GetRepository(dep)
		hasMatch, bestMatch := getVersion(rootLock, dep, repository, lockedVersion)
		var referenceName plumbing.ReferenceName

		worktree, _ := repository.Worktree()

		if !hasMatch {
			if masterReference := gitWrapper.GetMaster(repository); masterReference != nil {
				referenceName = plumbing.NewBranchReferenceName(masterReference.Name)
			}
		} else {
			referenceName = bestMatch.Name()
			if dep.GetVersion() == consts.MinimalDependencyVersion {
				pkg.Dependencies.(map[string]interface{})[dep.Repository] = "^" + referenceName.Short()
			}
		}

		if !rootLock.NeedUpdate(dep, referenceName.Short()) {
			msg.Info("  %s already updated", dep.GetName())
			continue
		} else if !hasMatch {
			msg.Warn("  No candidate to version for %s. Using master branch", dep.GetVersion())
		} else {
			msg.Info("  Detected semantic version. Using version %s", bestMatch.Name().Short())
		}

		err := worktree.Checkout(&git.CheckoutOptions{
			Force:  true,
			Branch: referenceName,
		})

		rootLock.AddInstalled(dep, referenceName.Short())

		if err != nil {
			msg.Die("  Error on switch to needed version from dependency %s\n%s", dep.Repository, err)
		}
	}
}

func getVersion(rootLock models.PackageLock, dep models.Dependency, repository *git.Repository, lockedVersion bool) (bool, *plumbing.Reference) {
	versions := gitWrapper.GetVersions(repository)
	constraints, e := semver.NewConstraint(dep.GetVersion())
	if e != nil {
		msg.Err("  Version type not supported! %s", e)
	}
	var bestVersion *semver.Version
	hasMatch := false
	var bestMatch *plumbing.Reference = nil
	if !lockedVersion {

		for _, version := range versions {
			short := version.Name().Short()
			newVersion, err := semver.NewVersion(short)
			if err != nil {
				continue
			}
			if constraints.Check(newVersion) {
				hasMatch = true
				if bestVersion == nil || newVersion.GreaterThan(bestVersion) {
					bestMatch = version
					bestVersion = newVersion
				}
			}
		}
	} else {
		lockedDependency := rootLock.GetInstalled(dep)
		if tag := gitWrapper.GetByTag(repository, lockedDependency.Version); tag != nil {
			return true, tag
		} else {
			msg.Warn("Tag not found %s, using semantic now...", lockedDependency.Version)
			return getVersion(rootLock, dep, repository, false)
		}

	}

	return hasMatch, bestMatch
}
