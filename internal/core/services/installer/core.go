package installer

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	goGit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	git "github.com/hashload/boss/internal/adapters/secondary/git"
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/compiler"
	"github.com/hashload/boss/internal/core/services/paths"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/utils"
	"github.com/hashload/boss/utils/librarypath"
)

type installContext struct {
	rootLocked       *domain.PackageLock
	root             *domain.Package
	processed        []string
	useLockedVersion bool
}

func newInstallContext(pkg *domain.Package, useLockedVersion bool) *installContext {
	return &installContext{
		rootLocked:       &pkg.Lock,
		root:             pkg,
		useLockedVersion: useLockedVersion,
		processed:        consts.DefaultPaths(),
	}
}

func DoInstall(pkg *domain.Package, lockedVersion bool) {
	msg.Info("Installing modules in project path")

	installContext := newInstallContext(pkg, lockedVersion)

	dependencies := installContext.ensureDependencies(pkg)

	paths.EnsureCleanModulesDir(dependencies, pkg.Lock)

	pkg.Lock.CleanRemoved(dependencies)
	pkg.Save()

	librarypath.UpdateLibraryPath(pkg)
	msg.Info("Compiling units")
	compiler.Build(pkg)
	pkg.Save()
	msg.Info("Success!")
}

func (ic *installContext) ensureDependencies(pkg *domain.Package) []domain.Dependency {
	if pkg.Dependencies == nil {
		return []domain.Dependency{}
	}
	deps := pkg.GetParsedDependencies()

	ic.ensureModules(pkg, deps)

	deps = append(deps, ic.processOthers()...)

	return deps
}

func (ic *installContext) processOthers() []domain.Dependency {
	infos, err := os.ReadDir(env.GetModulesDir())
	var lenProcessedInitial = len(ic.processed)
	var result []domain.Dependency
	if err != nil {
		msg.Err("Error on try load dir of modules: %s", err)
	}

	for _, info := range infos {
		if !info.IsDir() {
			continue
		}

		if utils.Contains(ic.processed, info.Name()) {
			continue
		}

		ic.processed = append(ic.processed, info.Name())

		msg.Info("Processing module %s", info.Name())

		fileName := filepath.Join(env.GetModulesDir(), info.Name(), consts.FilePackage)

		_, err := os.Stat(fileName)
		if os.IsNotExist(err) {
			msg.Warn("  boss.json not exists in %s", info.Name())
		}

		if packageOther, err := domain.LoadPackageOther(fileName); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			msg.Err("  Error on try load package %s: %s", fileName, err)
		} else {
			result = append(result, ic.ensureDependencies(packageOther)...)
		}
	}
	if lenProcessedInitial > len(ic.processed) {
		result = append(result, ic.processOthers()...)
	}

	return result
}

func (ic *installContext) ensureModules(pkg *domain.Package, deps []domain.Dependency) {
	for _, dep := range deps {
		msg.Info("Processing dependency %s", dep.Name())

		if ic.shouldSkipDependency(dep) {
			msg.Info("Dependency %s already installed", dep.Name())
			continue
		}

		GetDependency(dep)
		repository := git.GetRepository(dep)
		referenceName := ic.getReferenceName(pkg, dep, repository)

		wt, err := repository.Worktree()
		if err != nil {
			msg.Die("  Error on get worktree from repository %s\n%s", dep.Repository, err)
		}

		status, err := wt.Status()
		if err != nil {
			msg.Die("  Error on get status from worktree %s\n%s", dep.Repository, err)
		}

		head, er := repository.Head()
		if er != nil {
			msg.Die("  Error on get head from repository %s\n%s", dep.Repository, er)
		}

		currentRef := head.Name()
		if !ic.rootLocked.NeedUpdate(dep, referenceName.Short()) && status.IsClean() && referenceName == currentRef {
			msg.Info("  %s already updated", dep.Name())
			continue
		}

		ic.checkoutAndUpdate(dep, repository, referenceName)
	}
}

func (ic *installContext) shouldSkipDependency(dep domain.Dependency) bool {
	if !ic.useLockedVersion {
		return false
	}

	installed, exists := ic.rootLocked.Installed[strings.ToLower(dep.GetURL())]
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

func (ic *installContext) getReferenceName(
	pkg *domain.Package,
	dep domain.Dependency,
	repository *goGit.Repository) plumbing.ReferenceName {
	bestMatch := ic.getVersion(dep, repository)
	var referenceName plumbing.ReferenceName

	if bestMatch == nil {
		msg.Warn("No matching version found for '%s' with constraint '%s'", dep.Repository, dep.GetVersion())
		if mainBranchReference, err := git.GetMain(repository); err == nil {
			msg.Info("Falling back to main branch: %s", mainBranchReference.Name)
			return plumbing.NewBranchReferenceName(mainBranchReference.Name)
		}
		msg.Die("Could not find any suitable version or branch for dependency '%s'", dep.Repository)
	}

	referenceName = bestMatch.Name()
	if dep.GetVersion() == consts.MinimalDependencyVersion {
		pkg.Dependencies[dep.Repository] = "^" + referenceName.Short()
	}

	return referenceName
}

func (ic *installContext) checkoutAndUpdate(
	dep domain.Dependency,
	repository *goGit.Repository,
	referenceName plumbing.ReferenceName) {
	worktree, err := repository.Worktree()
	if err != nil {
		msg.Die("  Error on get worktree from repository %s\n%s", dep.Repository, err)
	}

	err = worktree.Checkout(&goGit.CheckoutOptions{
		Force:  true,
		Branch: referenceName,
	})

	ic.rootLocked.Add(dep, referenceName.Short())

	if err != nil {
		msg.Die("  Error on switch to needed version from dependency %s\n%s", dep.Repository, err)
	}

	err = worktree.Pull(&goGit.PullOptions{
		Force: true,
		Auth:  env.GlobalConfiguration().GetAuth(dep.GetURLPrefix()),
	})

	if err != nil && !errors.Is(err, goGit.NoErrAlreadyUpToDate) {
		msg.Warn("  Error on pull from dependency %s\n%s", dep.Repository, err)
	}
}

func (ic *installContext) getVersion(
	dep domain.Dependency,
	repository *goGit.Repository,
) *plumbing.Reference {
	if ic.useLockedVersion {
		lockedDependency := ic.rootLocked.GetInstalled(dep)

		if tag := git.GetByTag(repository, lockedDependency.Version); tag != nil &&
			lockedDependency.Version != dep.GetVersion() {
			return tag
		}
	}

	versions := git.GetVersions(repository, dep)
	constraints, err := ParseConstraint(dep.GetVersion())
	if err != nil {
		msg.Warn("Version constraint '%s' not supported: %s", dep.GetVersion(), err)
		// Try exact match as fallback
		for _, version := range versions {
			if version.Name().Short() == dep.GetVersion() {
				return version
			}
		}
		msg.Warn("No exact match found for version '%s'. Available versions: %d", dep.GetVersion(), len(versions))
		return nil
	}

	return ic.getVersionSemantic(
		versions,
		constraints)
}

func (ic *installContext) getVersionSemantic(
	versions []*plumbing.Reference,
	contraint *semver.Constraints) *plumbing.Reference {
	var bestVersion *semver.Version
	var bestReference *plumbing.Reference

	for _, version := range versions {
		short := version.Name().Short()
		withoutPrefix := stripVersionPrefix(short)
		newVersion, err := semver.NewVersion(withoutPrefix)
		if err != nil {
			continue
		}
		if contraint.Check(newVersion) {
			if bestVersion != nil && newVersion.GreaterThan(bestVersion) {
				bestVersion = newVersion
				bestReference = version
			}

			if bestVersion == nil {
				bestVersion = newVersion
				bestReference = version
			} else if bestVersion.Equal(newVersion) {
				if strings.HasPrefix(short, "v") && !strings.HasPrefix(bestReference.Name().Short(), "v") {
					bestReference = version
				}
			}
		}
	}
	return bestReference
}
