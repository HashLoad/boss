package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/compiler/graphs"
	"github.com/hashload/boss/internal/core/services/compiler_selector"
	"github.com/hashload/boss/internal/core/services/tracker"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
)

// Build compiles the package and its dependencies.
func Build(pkg *domain.Package, compilerVersion, platform string) {
	ctx := compiler_selector.SelectionContext{
		Package:            pkg,
		CliCompilerVersion: compilerVersion,
		CliPlatform:        platform,
	}
	selected, err := compiler_selector.SelectCompiler(ctx)
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

func saveLoadOrder(queue *graphs.NodeQueue) error {
	var projects = ""
	for {
		if queue.IsEmpty() {
			break
		}
		node := queue.Dequeue()
		dependencyPath := filepath.Join(env.GetModulesDir(), node.Dep.Name(), consts.FilePackage)
		if dependencyPackage, err := domain.LoadPackageOther(dependencyPath); err == nil {
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

func buildOrderedPackages(pkg *domain.Package, selectedCompiler *compiler_selector.SelectedCompiler) {
	pkg.Save()
	queue := loadOrderGraph(pkg)

	var packageNames []string
	tempQueue := loadOrderGraph(pkg)
	for !tempQueue.IsEmpty() {
		node := tempQueue.Dequeue()
		packageNames = append(packageNames, node.Dep.Name())
	}

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
				msg.Warn("❌ Could not start build tracker: %s", err)
			} else {
				msg.SetQuietMode(true)
			}
		} else {
			msg.Debug("Debug mode: progress tracker disabled\n")
		}
	} else {
		msg.Info("📄 No packages to compile.\n")
	}

	for {
		if queue.IsEmpty() {
			break
		}
		node := queue.Dequeue()
		dependencyPath := filepath.Join(env.GetModulesDir(), node.Dep.Name())

		dependency := pkg.Lock.GetInstalled(node.Dep)

		if trackerPtr.IsEnabled() {
			trackerPtr.SetBuilding(node.Dep.Name(), "")
		} else {
			msg.Info("  🔨 Building %s", node.Dep.Name())
		}

		dependency.Changed = false
		if dependencyPackage, err := domain.LoadPackageOther(filepath.Join(dependencyPath, consts.FilePackage)); err == nil {
			dprojs := dependencyPackage.Projects
			if len(dprojs) > 0 {
				hasFailed := false
				for _, dproj := range dprojs {
					dprojPath, _ := filepath.Abs(filepath.Join(env.GetModulesDir(), node.Dep.Name(), dproj))
					if trackerPtr.IsEnabled() {
						trackerPtr.SetBuilding(node.Dep.Name(), filepath.Base(dproj))
					} else {
						msg.Info("  🔥 Compiling project: %s", filepath.Base(dproj))
					}
					if !compile(dprojPath, &node.Dep, pkg.Lock, trackerPtr, selectedCompiler) {
						dependency.Failed = true
						hasFailed = true
					}
				}
				ensureArtifacts(&dependency, node.Dep, env.GetModulesDir())
				moveArtifacts(node.Dep, env.GetModulesDir())

				if trackerPtr.IsEnabled() {
					if hasFailed {
						trackerPtr.SetFailed(node.Dep.Name(), consts.StatusMsgBuildError)
					} else {
						trackerPtr.SetSuccess(node.Dep.Name())
					}
				} else {
					if hasFailed {
						msg.Err("  ❌ Build failed for %s", node.Dep.Name())
					} else {
						msg.Info("  ✅ %s built successfully", node.Dep.Name())
					}
				}
			} else {
				if trackerPtr.IsEnabled() {
					trackerPtr.SetSkipped(node.Dep.Name(), consts.StatusMsgNoProjects)
				} else {
					msg.Info("  ⏭️ %s has no projects to build", node.Dep.Name())
				}
			}
		} else {
			if trackerPtr.IsEnabled() {
				trackerPtr.SetSkipped(node.Dep.Name(), consts.StatusMsgNoBossJSON)
			} else {
				msg.Info("  ⏭️ %s has no boss.json", node.Dep.Name())
			}
		}
		pkg.Lock.SetInstalled(node.Dep, dependency)
	}

	msg.SetQuietMode(false)
	trackerPtr.Stop()
}
