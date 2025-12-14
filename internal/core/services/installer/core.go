package installer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashload/boss/pkg/pkgmanager"

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
	"github.com/hashload/boss/internal/core/services/tracker"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/utils"
	"github.com/hashload/boss/utils/librarypath"
)

type installContext struct {
	config           env.ConfigProvider
	rootLocked       *domain.PackageLock
	root             *domain.Package
	processed        []string
	visited          map[string]bool
	useLockedVersion bool
	progress         *ProgressTracker
	lockSvc          *lockService.LockService
	modulesDir       string
	options          InstallOptions
	warnings         []string
	depManager       *DependencyManager
	requestedDeps    map[string]bool // Track which dependencies were explicitly requested
}

func newInstallContext(config env.ConfigProvider, pkg *domain.Package, options InstallOptions, progress *ProgressTracker) *installContext {
	fs := filesystem.NewOSFileSystem()
	lockRepo := repository.NewFileLockRepository(fs)
	lockSvc := lockService.NewLockService(lockRepo, fs)

	requestedDeps := make(map[string]bool)
	if len(options.Args) > 0 {
		for _, arg := range options.Args {
			normalized := ParseDependency(arg)
			requestedDeps[normalized] = true
		}
	}

	return &installContext{
		config:           config,
		rootLocked:       &pkg.Lock,
		root:             pkg,
		useLockedVersion: options.LockedVersion,
		processed:        consts.DefaultPaths(),
		visited:          make(map[string]bool),
		progress:         progress,
		lockSvc:          lockSvc,
		modulesDir:       env.GetModulesDir(),
		options:          options,
		warnings:         make([]string, 0),
		depManager:       NewDefaultDependencyManager(config),
		requestedDeps:    requestedDeps,
	}
}

// DoInstall performs the installation of dependencies.
func DoInstall(config env.ConfigProvider, options InstallOptions, pkg *domain.Package) error {
	msg.Info("🔍 Analyzing dependencies...\n")

	deps := collectDependenciesToInstall(pkg, options.Args)

	if len(deps) == 0 {
		msg.Info("📄 No dependencies to install")
		return nil
	}

	var progress *ProgressTracker
	if msg.IsDebugMode() {
		progress = &ProgressTracker{
			Tracker: tracker.NewNull[DependencyStatus](),
		}
	} else {
		progress = NewProgressTracker(deps)
	}
	installContext := newInstallContext(config, pkg, options, progress)

	msg.Info("✨ Installing %d dependencies:\n", len(deps))

	if !msg.IsDebugMode() {
		if err := progress.Start(); err != nil {
			msg.Warn("⚠️ Could not start progress tracker: %s", err)
		} else {
			msg.SetQuietMode(true)
			msg.SetProgressTracker(progress)
		}
	}

	dependencies, err := installContext.ensureDependencies(pkg)
	if err != nil {
		msg.SetQuietMode(false)
		msg.SetProgressTracker(nil)
		progress.Stop()
		return fmt.Errorf("❌ Installation failed: %w", err)
	}

	msg.SetQuietMode(false)
	msg.SetProgressTracker(nil)
	progress.Stop()

	paths.EnsureCleanModulesDir(dependencies, pkg.Lock)

	pkg.Lock.CleanRemoved(dependencies)
	pkgmanager.SavePackageCurrent(pkg)
	if err := installContext.lockSvc.Save(&pkg.Lock, env.GetCurrentDir()); err != nil {
		msg.Warn("⚠️ Failed to save lock file: %v", err)
	}

	librarypath.UpdateLibraryPath(pkg)

	compiler.Build(pkg, options.Compiler, options.Platform)
	pkgmanager.SavePackageCurrent(pkg)
	if err := installContext.lockSvc.Save(&pkg.Lock, env.GetCurrentDir()); err != nil {
		msg.Warn("⚠️ Failed to save lock file: %v", err)
	}

	if len(installContext.warnings) > 0 {
		msg.Warn("⚠️ Installation Warnings:")
		for _, warning := range installContext.warnings {
			msg.Warn("   - %s", warning)
		}
	}

	msg.Success("✅ Installation completed successfully!")
	return nil
}

func (ic *installContext) addWarning(warning string) {
	ic.warnings = append(ic.warnings, warning)
}

// collectDependenciesToInstall collects dependencies to install based on args filter.
// If args is empty, returns all dependencies. Otherwise, returns only specified ones.
func collectDependenciesToInstall(pkg *domain.Package, args []string) []domain.Dependency {
	if pkg.Dependencies == nil {
		return []domain.Dependency{}
	}

	allDeps := pkg.GetParsedDependencies()

	if len(args) == 0 {
		return allDeps
	}

	var filtered []domain.Dependency
	for _, arg := range args {
		normalized := ParseDependency(arg)
		for _, dep := range allDeps {
			if dep.Repository == normalized {
				filtered = append(filtered, dep)
				break
			}
		}
	}

	return filtered
}

// collectAllDependencies makes a dry-run to collect all dependencies without installing.
// Deprecated: Use collectDependenciesToInstall instead.
func collectAllDependencies(pkg *domain.Package) []domain.Dependency {
	return collectDependenciesToInstall(pkg, []string{})
}

func (ic *installContext) ensureDependencies(pkg *domain.Package) ([]domain.Dependency, error) {
	if pkg.Dependencies == nil {
		return []domain.Dependency{}, nil
	}

	allDeps := pkg.GetParsedDependencies()

	var deps []domain.Dependency
	if pkg == ic.root && len(ic.requestedDeps) > 0 {
		for _, dep := range allDeps {
			if ic.requestedDeps[dep.Repository] {
				deps = append(deps, dep)
			}
		}
	} else {
		deps = allDeps
	}

	if err := ic.ensureModules(pkg, deps); err != nil {
		return nil, err
	}

	var otherDeps []domain.Dependency
	if len(ic.requestedDeps) == 0 {
		var err error
		otherDeps, err = ic.processOthers()
		if err != nil {
			return nil, err
		}
	}

	deps = append(deps, otherDeps...)

	return deps, nil
}

func (ic *installContext) processOthers() ([]domain.Dependency, error) {
	infos, err := os.ReadDir(env.GetModulesDir())
	var lenProcessedInitial = len(ic.processed)
	var result []domain.Dependency
	if err != nil {
		msg.Err("  ❌ Error on try load dir of modules: %s", err)
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
			msg.Info("  ⚙️ Processing module %s", moduleName)
		}

		fileName := filepath.Join(env.GetModulesDir(), moduleName, consts.FilePackage)

		_, err := os.Stat(fileName)
		if os.IsNotExist(err) {
			continue
		}

		if packageOther, err := pkgmanager.LoadPackageOther(fileName); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			msg.Err("  ❌ Error on try load package %s: %s", fileName, err)
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
		if err := ic.ensureSingleModule(pkg, dep); err != nil {
			return err
		}
	}
	return nil
}

func (ic *installContext) ensureSingleModule(pkg *domain.Package, dep domain.Dependency) error {
	depName := dep.Name()

	if ic.visited[depName] {
		return nil
	}
	ic.visited[depName] = true
	ic.progress.AddDependency(depName)

	if ic.shouldSkipDependency(dep) {
		ic.reportSkipped(depName, consts.StatusMsgAlreadyInstalled)
		return nil
	}

	if err := ic.cloneDependency(dep, depName); err != nil {
		return err
	}

	repository := git.GetRepository(dep)
	referenceName := ic.getReferenceName(pkg, dep, repository)

	if skip, err := ic.checkIfUpToDate(dep, depName, repository, referenceName); err != nil {
		return err
	} else if skip {
		return nil
	}

	return ic.installDependency(dep, depName, repository, referenceName)
}

func (ic *installContext) cloneDependency(dep domain.Dependency, depName string) error {
	if !ic.progress.IsEnabled() {
		msg.Info("🧬 Cloning %s", depName)
	} else {
		ic.reportStatus(depName, "cloning", "🧬 Cloning")
	}

	err := GetDependencyWithProgress(dep, ic.progress)
	if err != nil {
		ic.progress.SetFailed(depName, err)
		return err
	}
	return nil
}

func (ic *installContext) checkIfUpToDate(
	dep domain.Dependency,
	depName string,
	repository *goGit.Repository,
	referenceName plumbing.ReferenceName,
) (bool, error) {
	ic.reportStatus(depName, "checking", "🔍 Checking version for")

	wt, err := repository.Worktree()
	if err != nil {
		ic.progress.SetFailed(depName, err)
		return false, err
	}

	status, err := wt.Status()
	if err != nil {
		ic.progress.SetFailed(depName, err)
		return false, err
	}

	head, err := repository.Head()
	if err != nil {
		ic.progress.SetFailed(depName, err)
		return false, err
	}

	currentRef := head.Name()
	needsUpdate := ic.lockSvc.NeedUpdate(ic.rootLocked, dep, referenceName.Short(), ic.modulesDir)

	if !needsUpdate && status.IsClean() && referenceName == currentRef {
		ic.reportSkipped(depName, consts.StatusMsgUpToDate)
		return true, nil
	}

	return false, nil
}

func (ic *installContext) installDependency(
	dep domain.Dependency,
	depName string,
	repository *goGit.Repository,
	referenceName plumbing.ReferenceName,
) error {
	ic.reportStatus(depName, "installing", "🔥 Installing")

	if err := ic.checkoutAndUpdate(dep, repository, referenceName); err != nil {
		ic.progress.SetFailed(depName, err)
		return err
	}

	warning, err := ic.verifyDependencyCompatibility(dep)
	if err != nil {
		ic.progress.SetFailed(depName, err)
		return err
	}

	ic.reportInstallResult(depName, warning)
	return nil
}

func (ic *installContext) reportStatus(depName, progressStatus, infoPrefix string) {
	if ic.progress.IsEnabled() {
		switch progressStatus {
		case "cloning":
			ic.progress.SetCloning(depName)
		case "checking":
			ic.progress.SetChecking(depName, consts.StatusMsgResolvingVer)
		case "installing":
			ic.progress.SetInstalling(depName)
		}
	} else {
		msg.Info("  %s %s...", infoPrefix, depName)
	}
}

func (ic *installContext) reportSkipped(depName, reason string) {
	if ic.progress.IsEnabled() {
		ic.progress.SetSkipped(depName, reason)
	} else {
		msg.Info("  ✅️ %s already installed", depName)
	}
}

func (ic *installContext) reportInstallResult(depName, warning string) {
	if warning != "" {
		if ic.progress.IsEnabled() {
			ic.progress.SetWarning(depName, warning)
		} else {
			msg.Warn("  ⚠️ %s: %s", depName, warning)
		}
		ic.addWarning(fmt.Sprintf("%s: %s", depName, warning))
	} else {
		if ic.progress.IsEnabled() {
			ic.progress.SetCompleted(depName)
		} else {
			msg.Info("  ✅️ %s installed successfully", depName)
		}
	}
}

func (ic *installContext) shouldSkipDependency(dep domain.Dependency) bool {
	if utils.Contains(ic.options.ForceUpdate, dep.Name()) {
		return false
	}

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
		if !ic.progress.IsEnabled() {
			msg.Warn("  ⚠️ " + warnMsg)
		}
		ic.addWarning(fmt.Sprintf("%s: %s", dep.Name(), warnMsg))
		return false
	}

	installedVersion, err := semver.NewVersion(installed.Version)
	if err != nil {
		warnMsg := fmt.Sprintf("Error '%s' on get installed version. Updating...", err)
		if !ic.progress.IsEnabled() {
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
		if !ic.progress.IsEnabled() {
			msg.Warn("  ⚠️ " + warnMsg)
		}
		ic.addWarning(fmt.Sprintf("%s: %s", dep.Name(), warnMsg))

		if mainBranchReference, err := git.GetMain(repository); err == nil {
			warnMsg := fmt.Sprintf("Falling back to main branch: %s", mainBranchReference.Name)
			if !ic.progress.IsEnabled() {
				msg.Warn("  ⚠️ %s: %s", dep.Name(), warnMsg)
			}
			ic.addWarning(fmt.Sprintf("%s: %s", dep.Name(), warnMsg))
			return plumbing.NewBranchReferenceName(mainBranchReference.Name)
		}
		msg.Die("❌ Could not find any suitable version or branch for dependency '%s'", dep.Repository)
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

	if !ic.progress.IsEnabled() {
		msg.Debug("  🔍 Checking out %s to %s", dep.Name(), referenceName.Short())
	}
	err := git.Checkout(ic.config, dep, referenceName)

	ic.lockSvc.AddDependency(ic.rootLocked, dep, referenceName.Short(), ic.modulesDir)

	if err != nil {
		return err
	}

	if !ic.progress.IsEnabled() {
		msg.Debug("  📥 Pulling latest changes for %s", dep.Name())
	}
	err = git.Pull(ic.config, dep)

	if err != nil && !errors.Is(err, goGit.NoErrAlreadyUpToDate) {
		warnMsg := fmt.Sprintf("Error on pull from dependency %s\n%s", dep.Repository, err)
		if !ic.progress.IsEnabled() {
			msg.Warn("  " + warnMsg)
		}
		ic.addWarning(fmt.Sprintf("%s: %s", dep.Name(), warnMsg))
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

	versions := git.GetVersions(ic.config, repository, dep)
	constraints, err := domain.ParseConstraint(dep.GetVersion())
	if err != nil {
		warnMsg := fmt.Sprintf("Version constraint '%s' not supported: %s", dep.GetVersion(), err)
		if !ic.progress.IsEnabled() {
			msg.Warn("  ⚠️ " + warnMsg)
		}
		ic.addWarning(fmt.Sprintf("%s: %s", dep.Name(), warnMsg))

		for _, version := range versions {
			if version.Name().Short() == dep.GetVersion() {
				return version
			}
		}
		warnMsg2 := fmt.Sprintf("No exact match found for version '%s'. Available versions: %d", dep.GetVersion(), len(versions))
		if !ic.progress.IsEnabled() {
			msg.Warn("  ⚠️ " + warnMsg2)
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

	for _, versionRef := range versions {
		short := versionRef.Name().Short()
		withoutPrefix := domain.StripVersionPrefix(short)
		newVersion, err := semver.NewVersion(withoutPrefix)
		if err != nil {
			continue
		}
		if contraint.Check(newVersion) {
			if bestVersion != nil && newVersion.GreaterThan(bestVersion) {
				bestVersion = newVersion
				bestReference = versionRef
			}

			if bestVersion == nil {
				bestVersion = newVersion
				bestReference = versionRef
			} else if bestVersion.Equal(newVersion) {
				if strings.HasPrefix(short, "v") && !strings.HasPrefix(bestReference.Name().Short(), "v") {
					bestReference = versionRef
				}
			}
		}
	}
	return bestReference
}

func (ic *installContext) verifyDependencyCompatibility(dep domain.Dependency) (string, error) {
	depPath := filepath.Join(ic.modulesDir, dep.Name())
	depPkg, err := pkgmanager.LoadPackageOther(filepath.Join(depPath, "boss.json"))
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
