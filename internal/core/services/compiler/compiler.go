package compiler

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/compiler/graphs"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/utils"
)

func Build(pkg *domain.Package) {
	buildOrderedPackages(pkg)
	graph := LoadOrderGraphAll(pkg)
	saveLoadOrder(graph)
}

func saveLoadOrder(queue *graphs.NodeQueue) {
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

	utils.HandleError(os.WriteFile(outDir, []byte(projects), 0600))
}

func buildOrderedPackages(pkg *domain.Package) {
	pkg.Save()
	queue := loadOrderGraph(pkg)

	var packageNames []string
	tempQueue := loadOrderGraph(pkg)
	for !tempQueue.IsEmpty() {
		node := tempQueue.Dequeue()
		packageNames = append(packageNames, node.Dep.Name())
	}

	tracker := NewBuildTracker(packageNames)
	if len(packageNames) > 0 {
		msg.Info("Compiling %d packages:\n", len(packageNames))
		if err := tracker.Start(); err != nil {
			msg.Warn("Could not start build tracker: %s", err)
		} else {
			msg.SetQuietMode(true)
		}
	}

	for {
		if queue.IsEmpty() {
			break
		}
		node := queue.Dequeue()
		dependencyPath := filepath.Join(env.GetModulesDir(), node.Dep.Name())

		dependency := pkg.Lock.GetInstalled(node.Dep)

		if tracker.IsEnabled() {
			tracker.SetBuilding(node.Dep.Name(), "")
		} else {
			msg.Info("Building %s", node.Dep.Name())
		}

		dependency.Changed = false
		if dependencyPackage, err := domain.LoadPackageOther(filepath.Join(dependencyPath, consts.FilePackage)); err == nil {
			dprojs := dependencyPackage.Projects
			if len(dprojs) > 0 {
				hasFailed := false
				for _, dproj := range dprojs {
					dprojPath, _ := filepath.Abs(filepath.Join(env.GetModulesDir(), node.Dep.Name(), dproj))
					if tracker.IsEnabled() {
						tracker.SetBuilding(node.Dep.Name(), filepath.Base(dproj))
					}
					if !compile(dprojPath, &node.Dep, pkg.Lock, tracker) {
						dependency.Failed = true
						hasFailed = true
					}
				}
				ensureArtifacts(&dependency, node.Dep, env.GetModulesDir())
				moveArtifacts(node.Dep, env.GetModulesDir())

				if tracker.IsEnabled() {
					if hasFailed {
						tracker.SetFailed(node.Dep.Name(), "build error")
					} else {
						tracker.SetSuccess(node.Dep.Name())
					}
				}
			} else {
				if tracker.IsEnabled() {
					tracker.SetSkipped(node.Dep.Name(), "no projects")
				}
			}
		} else {
			if tracker.IsEnabled() {
				tracker.SetSkipped(node.Dep.Name(), "no boss.json")
			}
		}
		pkg.Lock.SetInstalled(node.Dep, dependency)
	}

	msg.SetQuietMode(false)
	tracker.Stop()
}
