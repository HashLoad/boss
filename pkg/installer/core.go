package installer

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	goGit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashload/boss/pkg/compiler"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/git"
	"github.com/hashload/boss/pkg/models"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/pkg/paths"
	"github.com/hashload/boss/utils"
	"github.com/hashload/boss/utils/librarypath"
	"github.com/masterminds/semver"
)

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

func EnsureDependencies(
	rootLock models.PackageLock,
	pkg *models.Package,
	lockedVersion bool,
	processed ...string) []models.Dependency {
	if pkg.Dependencies == nil {
		return []models.Dependency{}
	}
	deps := pkg.GetParsedDependencies()

	ensureModules(rootLock, pkg, deps, lockedVersion)
	if len(processed) == 0 {
		processed = append(processed, consts.DefaultPaths()...)
	}

	deps = append(deps, processOthers(rootLock, lockedVersion, processed)...)

	return deps
}

func processOthers(rootLock models.PackageLock, lockedVersion bool, processed []string) []models.Dependency {
	infos, err := os.ReadDir(env.GetModulesDir())
	var lenProcessedInitial = len(processed)
	var result []models.Dependency
	if err != nil {
		msg.Err("Error on try load dir of modules: %s", err)
	}

	for _, info := range infos {
		if !info.IsDir() {
			continue
		}

		if utils.Contains(processed, info.Name()) {
			continue
		}

		processed = append(processed, info.Name())

		msg.Info("Processing module %s", info.Name())

		fileName := filepath.Join(env.GetModulesDir(), info.Name(), consts.FilePackage)

		_, err := os.Stat(fileName)
		if os.IsNotExist(err) {
			msg.Warn("  boss.json not exists in %s", info.Name())
		}

		if packageOther, err := models.LoadPackageOther(fileName); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			msg.Err("  Error on try load package %s: %s", fileName, err)
		} else {
			result = append(result, EnsureDependencies(rootLock, packageOther, lockedVersion, processed...)...)
		}
	}
	if lenProcessedInitial > len(processed) {
		result = append(result, processOthers(rootLock, lockedVersion, processed)...)
	}

	return result
}

func ensureModules(rootLock models.PackageLock, pkg *models.Package, deps []models.Dependency, lockedVersion bool) {
	msg.Info("Installing modules")
	for _, dep := range deps {
		msg.Info("Processing dependency %s", dep.GetName())

		if shouldSkipDependency(rootLock, dep, lockedVersion) {
			msg.Info("Dependency %s already installed", dep.GetName())
			continue
		}

		GetDependency(dep)
		repository := git.GetRepository(dep)
		referenceName := getReferenceName(rootLock, pkg, dep, repository)

		if !rootLock.NeedUpdate(dep, referenceName.Short()) {
			msg.Info("  %s already updated", dep.GetName())
			continue
		}

		checkoutAndUpdate(rootLock, dep, repository, referenceName)
	}
}

func shouldSkipDependency(rootLock models.PackageLock, dep models.Dependency, lockedVersion bool) bool {
	if !lockedVersion {
		return false
	}

	installed, exists := rootLock.Installed[strings.ToLower(dep.GetURL())]
	if !exists {
		return false
	}

	depv := strings.NewReplacer("^", "", "~", "").Replace(dep.GetVersion())
	requiredVersion, err := semver.NewVersion(depv)
	if err != nil {
		msg.Warn("  Error '%s' on get required version. Updating...", err)
		return false
	}

	installedVersion, err := semver.NewVersion(installed.Version)
	if err != nil {
		msg.Warn("  Error '%s' on get installed version. Updating...", err)
		return false
	}

	return !installedVersion.LessThan(requiredVersion)
}

func getReferenceName(
	rootLock models.PackageLock,
	pkg *models.Package,
	dep models.Dependency,
	repository *goGit.Repository) plumbing.ReferenceName {
	bestMatch := getVersion(rootLock, dep, repository, false)
	var referenceName plumbing.ReferenceName

	if bestMatch == nil {
		if mainBranchReference, err := git.GetMain(repository); err == nil {
			referenceName = plumbing.NewBranchReferenceName(mainBranchReference.Name)
		}
	} else {
		referenceName = bestMatch.Name()
		if dep.GetVersion() == consts.MinimalDependencyVersion {
			pkg.Dependencies[dep.Repository] = "^" + referenceName.Short()
		}
	}

	return referenceName
}

func checkoutAndUpdate(
	rootLock models.PackageLock,
	dep models.Dependency,
	repository *goGit.Repository,
	referenceName plumbing.ReferenceName) {
	worktree, _ := repository.Worktree()

	err := worktree.Checkout(&goGit.CheckoutOptions{
		Force:  true,
		Branch: referenceName,
	})

	rootLock.AddInstalled(dep, referenceName.Short())

	if err != nil {
		msg.Die("  Error on switch to needed version from dependency %s\n%s", dep.Repository, err)
	}

	err = worktree.Pull(&goGit.PullOptions{
		Force: true,

		Auth: env.GlobalConfiguration().GetAuth(dep.GetURLPrefix()),
	})

	if err != nil && !errors.Is(err, goGit.NoErrAlreadyUpToDate) {
		msg.Warn("  Error on pull from dependency %s\n%s", dep.Repository, err)
	}
}

func getVersion(
	rootLock models.PackageLock,
	dep models.Dependency,
	repository *goGit.Repository,
	locked bool,
) *plumbing.Reference {
	versions := git.GetVersions(repository, dep)
	constraints, err := semver.NewConstraint(dep.GetVersion())
	if err != nil {
		msg.Warn("  Error on get constraints %s, using branch format now...", err)
		for _, version := range versions {
			if version.Name().Short() == dep.GetVersion() {
				return version
			}
		}

		return nil
	}

	return getVersionSemantic(
		rootLock,
		dep,
		repository,
		versions,
		constraints,
		locked)
}

func getVersionSemantic(
	rootLock models.PackageLock,
	dep models.Dependency,
	repository *goGit.Repository,
	versions []*plumbing.Reference,
	contraint *semver.Constraints,
	locked bool) *plumbing.Reference {
	if locked {
		lockedDependency := rootLock.GetInstalled(dep)
		if tag := git.GetByTag(repository, lockedDependency.Version); tag != nil {
			return tag
		}

		msg.Warn("Tag not found %s, using semantic now...", lockedDependency.Version)
		return getVersion(rootLock, dep, repository, false)
	}

	var bestVersion *semver.Version

	for _, version := range versions {
		short := version.Name().Short()
		newVersion, err := semver.NewVersion(short)
		if err != nil {
			continue
		}
		if contraint.Check(newVersion) {
			if bestVersion != nil && newVersion.GreaterThan(bestVersion) {
				bestVersion = newVersion
			}

			if bestVersion == nil {
				bestVersion = newVersion
			}
		}
	}
	return nil
}
