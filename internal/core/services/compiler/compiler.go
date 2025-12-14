// Package compiler provides functionality for building Delphi projects and their dependencies.
// It handles dependency graph resolution, build order determination, and compilation execution.
package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashload/boss/internal/adapters/secondary/filesystem"
	"github.com/hashload/boss/internal/core/services/compilerselector"
	"github.com/hashload/boss/pkg/pkgmanager"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/tracker"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
)

// Build compiles the package and its dependencies.
func Build(pkg *domain.Package, compilerVersion, platform string) {
	ctx := compilerselector.SelectionContext{
		Package:            pkg,
		CliCompilerVersion: compilerVersion,
		CliPlatform:        platform,
	}
	selected, err := compilerselector.SelectCompiler(ctx)
	if err != nil {
		msg.Warn("Compiler selection failed: %s. Falling back to default.", err)
	} else {
		msg.Info("🛠️ Using compiler:")
		msg.Info("   Version: %s", selected.Version)
		msg.Info("   Platform: %s", selected.Arch)
		msg.Info("   Binary: %s", selected.Path)
	}

	buildOrderedPackages(pkg, selected)
	graph := LoadOrderGraphAll(pkg)
	if err := saveLoadOrder(graph); err != nil {
		msg.Warn("⚠️ Failed to save build order: %v", err)
	}
}

func saveLoadOrder(queue *domain.NodeQueue) error {
	var projects = ""
	for {
		if queue.IsEmpty() {
			break
		}
		node := queue.Dequeue()
		dependencyPath := filepath.Join(env.GetModulesDir(), node.Dep.Name(), consts.FilePackage)
		if dependencyPackage, err := pkgmanager.LoadPackageOther(dependencyPath); err == nil {
			for _, value := range dependencyPackage.Projects {
				projects += strings.TrimSuffix(filepath.Base(value), filepath.Ext(value)) + consts.FileExtensionBpl + "\n"
			}
		}
	}
	outDir := filepath.Join(env.GetModulesDir(), consts.BplFolder, consts.FileBplOrder)

	if err := os.WriteFile(outDir, []byte(projects), 0600); err != nil {
		return fmt.Errorf("failed to save build load order to %s: %w", outDir, err)
	}
	return nil
}

func buildOrderedPackages(pkg *domain.Package, selectedCompiler *compilerselector.SelectedCompiler) {
	_ = pkgmanager.SavePackageCurrent(pkg)
	queue := loadOrderGraph(pkg)
	packageNames := extractPackageNames(pkg)

	trackerPtr := initializeBuildTracker(packageNames)
	if len(packageNames) == 0 {
		msg.Info("📄 No packages to compile.\n")
		return
	}

	processPackageQueue(pkg, queue, trackerPtr, selectedCompiler)

	msg.SetQuietMode(false)
	trackerPtr.Stop()
}

func extractPackageNames(pkg *domain.Package) []string {
	var packageNames []string
	tempQueue := loadOrderGraph(pkg)
	for !tempQueue.IsEmpty() {
		node := tempQueue.Dequeue()
		packageNames = append(packageNames, node.Dep.Name())
	}
	return packageNames
}

func initializeBuildTracker(packageNames []string) *BuildTracker {
	var trackerPtr *BuildTracker
	if msg.IsDebugMode() {
		trackerPtr = &BuildTracker{
			Tracker: tracker.NewNull[BuildStatus](),
		}
	} else {
		trackerPtr = NewBuildTracker(packageNames)
	}

	if len(packageNames) > 0 {
		msg.Info("📦 Compiling %d packages:\n", len(packageNames))
		if !msg.IsDebugMode() {
			if err := trackerPtr.Start(); err != nil {
				msg.Warn("⚠️ Could not start build tracker: %s", err)
			} else {
				msg.SetQuietMode(true)
			}
		} else {
			msg.Debug("Debug mode: progress tracker disabled\n")
		}
	}
	return trackerPtr
}

func processPackageQueue(
	pkg *domain.Package,
	queue *domain.NodeQueue,
	trackerPtr *BuildTracker,
	selectedCompiler *compilerselector.SelectedCompiler,
) {
	fs := filesystem.NewOSFileSystem()
	artifactMgr := NewDefaultArtifactManager(fs)

	for !queue.IsEmpty() {
		node := queue.Dequeue()
		processPackageNode(pkg, node, trackerPtr, selectedCompiler, artifactMgr)
	}
}

func processPackageNode(
	pkg *domain.Package,
	node *domain.Node,
	trackerPtr *BuildTracker,
	selectedCompiler *compilerselector.SelectedCompiler,
	artifactMgr *DefaultArtifactManager,
) {
	dependencyPath := filepath.Join(env.GetModulesDir(), node.Dep.Name())
	dependency := pkg.Lock.GetInstalled(node.Dep)

	reportBuildStart(trackerPtr, node.Dep.Name())

	dependency.Changed = false
	dependencyPackage, err := pkgmanager.LoadPackageOther(filepath.Join(dependencyPath, consts.FilePackage))

	if err != nil {
		reportNoBossJSON(trackerPtr, node.Dep.Name())
		pkg.Lock.SetInstalled(node.Dep, dependency)
		return
	}

	if len(dependencyPackage.Projects) == 0 {
		reportNoProjects(trackerPtr, node.Dep.Name())
		pkg.Lock.SetInstalled(node.Dep, dependency)
		return
	}

	hasFailed := buildProjectsForDependency(
		&dependency,
		node.Dep,
		dependencyPackage.Projects,
		trackerPtr,
		selectedCompiler,
		pkg.Lock,
	)

	artifactMgr.EnsureArtifacts(&dependency, node.Dep, env.GetModulesDir())
	artifactMgr.MoveArtifacts(node.Dep, env.GetModulesDir())

	reportBuildResult(trackerPtr, node.Dep.Name(), hasFailed)
	pkg.Lock.SetInstalled(node.Dep, dependency)
}

func buildProjectsForDependency(
	dependency *domain.LockedDependency,
	dep domain.Dependency,
	projects []string,
	trackerPtr *BuildTracker,
	selectedCompiler *compilerselector.SelectedCompiler,
	lock domain.PackageLock,
) bool {
	hasFailed := false
	for _, dproj := range projects {
		dprojPath, _ := filepath.Abs(filepath.Join(env.GetModulesDir(), dep.Name(), dproj))

		if trackerPtr.IsEnabled() {
			trackerPtr.SetBuilding(dep.Name(), filepath.Base(dproj))
		} else {
			msg.Info("  🔥 Compiling project: %s", filepath.Base(dproj))
		}

		if !compile(dprojPath, &dep, lock, trackerPtr, selectedCompiler) {
			dependency.Failed = true
			hasFailed = true
		}
	}
	return hasFailed
}

func reportBuildStart(trackerPtr *BuildTracker, depName string) {
	if trackerPtr.IsEnabled() {
		trackerPtr.SetBuilding(depName, "")
	} else {
		msg.Info("  🔨 Building %s", depName)
	}
}

func reportBuildResult(trackerPtr *BuildTracker, depName string, hasFailed bool) {
	if trackerPtr.IsEnabled() {
		if hasFailed {
			trackerPtr.SetFailed(depName, consts.StatusMsgBuildError)
		} else {
			trackerPtr.SetSuccess(depName)
		}
	} else {
		if hasFailed {
			msg.Err("  ❌ Build failed for %s", depName)
		} else {
			msg.Info("  ✅ %s built successfully", depName)
		}
	}
}

func reportNoProjects(trackerPtr *BuildTracker, depName string) {
	if trackerPtr.IsEnabled() {
		trackerPtr.SetSkipped(depName, consts.StatusMsgNoProjects)
	} else {
		msg.Info("  ⏭️ %s has no projects to build", depName)
	}
}

func reportNoBossJSON(trackerPtr *BuildTracker, depName string) {
	if trackerPtr.IsEnabled() {
		trackerPtr.SetSkipped(depName, consts.StatusMsgNoBossJSON)
	} else {
		msg.Info("  ⏭️ %s has no boss.json", depName)
	}
}
