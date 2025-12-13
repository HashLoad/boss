package installer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	goGit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/hashload/boss/internal/adapters/secondary/filesystem"
	git "github.com/hashload/boss/internal/adapters/secondary/git"
	"github.com/hashload/boss/internal/adapters/secondary/repository"
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/compiler"
	lockService "github.com/hashload/boss/internal/core/services/lock"
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
	progress         *ProgressTracker
	lockSvc          *lockService.Service
	modulesDir       string
	options          InstallOptions
	warnings         []string
}

func newInstallContext(pkg *domain.Package, options InstallOptions, progress *ProgressTracker) *installContext {
	fs := filesystem.NewOSFileSystem()
	lockRepo := repository.NewFileLockRepository(fs)
	lockSvc := lockService.NewService(lockRepo, fs)

	return &installContext{
		rootLocked:       &pkg.Lock,
		root:             pkg,
		useLockedVersion: options.LockedVersion,
		processed:        consts.DefaultPaths(),
		progress:         progress,
		lockSvc:          lockSvc,
		modulesDir:       env.GetModulesDir(),
		options:          options,
		warnings:         make([]string, 0),
	}
}

func DoInstall(options InstallOptions, pkg *domain.Package) {
	msg.Info("Analyzing dependencies...\n")

	deps := collectAllDependencies(pkg)

	if len(deps) == 0 {
		msg.Info("No dependencies to install")
		return
	}

	progress := NewProgressTracker(deps)
	installContext := newInstallContext(pkg, options, progress)

	msg.Info("Installing %d dependencies:\n", len(deps))

	if err := progress.Start(); err != nil {
		msg.Warn("Could not start progress tracker: %s", err)
	} else {
		msg.SetQuietMode(true)
		msg.SetProgressTracker(progress)
	}

	dependencies, err := installContext.ensureDependencies(pkg)
	if err != nil {
		msg.SetQuietMode(false)
		msg.SetProgressTracker(nil)
		progress.Stop()
		msg.Err("  Installation failed: %s", err)
		os.Exit(1)
	}

	msg.SetQuietMode(false)
	msg.SetProgressTracker(nil)
	progress.Stop()

	paths.EnsureCleanModulesDir(dependencies, pkg.Lock)

	pkg.Lock.CleanRemoved(dependencies)
	pkg.Save()
	installContext.lockSvc.Save(&pkg.Lock, env.GetCurrentDir())

	librarypath.UpdateLibraryPath(pkg)

	compiler.Build(pkg, options.Compiler, options.Platform)
	pkg.Save()
	installContext.lockSvc.Save(&pkg.Lock, env.GetCurrentDir())

	if len(installContext.warnings) > 0 {
		msg.Warn("\nInstallation Warnings:")
		for _, warning := range installContext.warnings {
			msg.Warn("  - %s", warning)
		}
		fmt.Println("")
	}

	msg.Success("✓ Installation completed successfully!")
}

func (ic *installContext) addWarning(warning string) {
	ic.warnings = append(ic.warnings, warning)
}

// collectAllDependencies makes a dry-run to collect all dependencies without installing.
func collectAllDependencies(pkg *domain.Package) []domain.Dependency {
	if pkg.Dependencies == nil {
		return []domain.Dependency{}
	}

	return pkg.GetParsedDependencies()
}

func (ic *installContext) ensureDependencies(pkg *domain.Package) ([]domain.Dependency, error) {
	if pkg.Dependencies == nil {
		return []domain.Dependency{}, nil
	}
	deps := pkg.GetParsedDependencies()

	if err := ic.ensureModules(pkg, deps); err != nil {
		return nil, err
	}

	otherDeps, err := ic.processOthers()
	if err != nil {
		return nil, err
	}
	deps = append(deps, otherDeps...)

	return deps, nil
}

func (ic *installContext) processOthers() ([]domain.Dependency, error) {
	infos, err := os.ReadDir(env.GetModulesDir())
	var lenProcessedInitial = len(ic.processed)
	var result []domain.Dependency
	if err != nil {
		msg.Err("Error on try load dir of modules: %s", err)
		return result, err
	}

	for _, info := range infos {
		if !info.IsDir() {
			continue
		}

		moduleName := info.Name()

		if utils.Contains(ic.processed, moduleName) {
			continue
		}

		ic.processed = append(ic.processed, moduleName)

		if !ic.progress.IsEnabled() {
			msg.Info("Processing module %s", moduleName)
		}

		fileName := filepath.Join(env.GetModulesDir(), moduleName, consts.FilePackage)

		_, err := os.Stat(fileName)
		if os.IsNotExist(err) {
			continue
		}

		if packageOther, err := domain.LoadPackageOther(fileName); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			msg.Err("  Error on try load package %s: %s", fileName, err)
		} else {
			childDeps := packageOther.GetParsedDependencies()
			for _, childDep := range childDeps {
				ic.progress.AddDependency(childDep.Name())
			}
			deps, err := ic.ensureDependencies(packageOther)
			if err != nil {
				return nil, err
			}
			result = append(result, deps...)
		}
	}
	if lenProcessedInitial > len(ic.processed) {
		deps, err := ic.processOthers()
		if err != nil {
			return nil, err
		}
		result = append(result, deps...)
	}

	return result, nil
}

func (ic *installContext) ensureModules(pkg *domain.Package, deps []domain.Dependency) error {
	for _, dep := range deps {
		depName := dep.Name()

		ic.progress.AddDependency(depName)

		if ic.shouldSkipDependency(dep) {
			if ic.progress.IsEnabled() {
				ic.progress.SetSkipped(depName, "up to date")
			} else {
				msg.Info("  %s already installed", depName)
			}
			continue
		}

		if ic.progress.IsEnabled() {
			ic.progress.SetCloning(depName)
		} else {
			msg.Info("Processing dependency %s", depName)
		}

		err := GetDependencyWithProgress(dep, ic.progress)
		if err != nil {
			ic.progress.SetFailed(depName, err)
			return err
		}
		repository := git.GetRepository(dep)

		ic.progress.SetChecking(depName, "resolving version")

		referenceName := ic.getReferenceName(pkg, dep, repository)

		wt, err := repository.Worktree()
		if err != nil {
			ic.progress.SetFailed(depName, err)
			return err
		}

		status, err := wt.Status()
		if err != nil {
			ic.progress.SetFailed(depName, err)
			return err
		}

		head, er := repository.Head()
		if er != nil {
			ic.progress.SetFailed(depName, er)
			return er
		}

		currentRef := head.Name()

		needsUpdate := ic.lockSvc.NeedUpdate(ic.rootLocked, dep, referenceName.Short(), ic.modulesDir)
		if !needsUpdate && status.IsClean() && referenceName == currentRef {
			if ic.progress.IsEnabled() {
				ic.progress.SetSkipped(depName, "already up to date")
			} else {
				msg.Info("  %s already updated", depName)
			}
			continue
		}

		ic.progress.SetInstalling(depName)

		if err := ic.checkoutAndUpdate(dep, repository, referenceName); err != nil {
			ic.progress.SetFailed(depName, err)
			return err
		}

		warning, err := ic.verifyDependencyCompatibility(dep)
		if err != nil {
			ic.progress.SetFailed(depName, err)
			return err
		}

		if warning != "" {
			ic.progress.SetWarning(depName, warning)
			ic.addWarning(fmt.Sprintf("%s: %s", depName, warning))
		} else {
			ic.progress.SetCompleted(depName)
		}
	}
	return nil
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
		warnMsg := fmt.Sprintf("Error '%s' on get required version. Updating...", err)
		if ic.progress.IsEnabled() {
			ic.progress.SetWarning(dep.Name(), warnMsg)
		} else {
			msg.Warn("  " + warnMsg)
		}
		ic.addWarning(fmt.Sprintf("%s: %s", dep.Name(), warnMsg))
		return false
	}

	installedVersion, err := semver.NewVersion(installed.Version)
	if err != nil {
		warnMsg := fmt.Sprintf("Error '%s' on get installed version. Updating...", err)
		if ic.progress.IsEnabled() {
			ic.progress.SetWarning(dep.Name(), warnMsg)
		} else {
			msg.Warn("  " + warnMsg)
		}
		ic.addWarning(fmt.Sprintf("%s: %s", dep.Name(), warnMsg))
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
		warnMsg := fmt.Sprintf("No matching version found for '%s' with constraint '%s'", dep.Repository, dep.GetVersion())
		if ic.progress.IsEnabled() {
			ic.progress.SetWarning(dep.Name(), warnMsg)
		} else {
			msg.Warn(warnMsg)
		}
		ic.addWarning(fmt.Sprintf("%s: %s", dep.Name(), warnMsg))

		if mainBranchReference, err := git.GetMain(repository); err == nil {
			if !ic.progress.IsEnabled() {
				msg.Info("Falling back to main branch: %s", mainBranchReference.Name)
			}
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
	referenceName plumbing.ReferenceName) error {
	worktree, err := repository.Worktree()
	if err != nil {
		return err
	}

	err = worktree.Checkout(&goGit.CheckoutOptions{
		Force:  true,
		Branch: referenceName,
	})

	ic.lockSvc.AddDependency(ic.rootLocked, dep, referenceName.Short(), ic.modulesDir)

	if err != nil {
		return err
	}

	err = worktree.Pull(&goGit.PullOptions{
		Force: true,
		Auth:  env.GlobalConfiguration().GetAuth(dep.GetURLPrefix()),
	})

	if err != nil && !errors.Is(err, goGit.NoErrAlreadyUpToDate) {
		if ic.progress.IsEnabled() {
			ic.progress.SetWarning(dep.Name(), fmt.Sprintf("Error on pull from dependency %s\n%s", dep.Repository, err))
		} else {
			msg.Warn("  Error on pull from dependency %s\n%s", dep.Repository, err)
		}
	}
	return nil
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
		warnMsg := fmt.Sprintf("Version constraint '%s' not supported: %s", dep.GetVersion(), err)
		if ic.progress.IsEnabled() {
			ic.progress.SetWarning(dep.Name(), warnMsg)
		} else {
			msg.Warn(warnMsg)
		}
		ic.addWarning(fmt.Sprintf("%s: %s", dep.Name(), warnMsg))

		for _, version := range versions {
			if version.Name().Short() == dep.GetVersion() {
				return version
			}
		}
		warnMsg2 := fmt.Sprintf("No exact match found for version '%s'. Available versions: %d", dep.GetVersion(), len(versions))
		if ic.progress.IsEnabled() {
			ic.progress.SetWarning(dep.Name(), warnMsg2)
		} else {
			msg.Warn(warnMsg2)
		}
		ic.addWarning(fmt.Sprintf("%s: %s", dep.Name(), warnMsg2))
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

func (ic *installContext) verifyDependencyCompatibility(dep domain.Dependency) (string, error) {
	depPath := filepath.Join(ic.modulesDir, dep.Name())
	depPkg, err := domain.LoadPackageOther(filepath.Join(depPath, "boss.json"))
	if err != nil {
		return "", nil
	}

	if depPkg.Engines == nil || len(depPkg.Engines.Platforms) == 0 {
		return "", nil
	}

	targetPlatform := ic.options.Platform
	if targetPlatform == "" && ic.root.Toolchain != nil {
		targetPlatform = ic.root.Toolchain.Platform
	}

	if targetPlatform == "" {
		return "", nil
	}

	for _, p := range depPkg.Engines.Platforms {
		if strings.EqualFold(p, targetPlatform) {
			return "", nil
		}
	}

	errorMessage := fmt.Sprintf("Dependency '%s' does not support platform '%s'. Supported: %v", dep.Name(), targetPlatform, depPkg.Engines.Platforms)

	isStrict := ic.options.Strict
	if !isStrict && ic.root.Toolchain != nil {
		isStrict = ic.root.Toolchain.Strict
	}

	if isStrict {
		return "", errors.New(errorMessage)
	}
	return errorMessage, nil
}
