package compiler

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"path/filepath"
)

func Build(pkg *models.Package) {
	buildOrderedPackages(pkg)
}

func buildOrderedPackages(pkg *models.Package) {
	pkg.Lock.Save()
	queue := loadOrderGraph(pkg)
	for {
		if queue.IsEmpty() {
			break
		}
		node := queue.Dequeue()
		dependencyPath := filepath.Join(env.GetModulesDir(), node.Dep.GetName())

		dependency := pkg.Lock.GetInstalled(node.Dep)

		if !dependency.Changed {
			continue
		} else {
			msg.Info("Building %s", node.Dep.GetName())
			dependency.Changed = false
			if dependencyPackage, err := models.LoadPackageOther(filepath.Join(dependencyPath, consts.FilePackage)); err == nil {
				dprojs := dependencyPackage.Projects
				for _, dproj := range dprojs {
					s, _ := filepath.Abs(filepath.Join(env.GetModulesDir(), node.Dep.GetName(), dproj))
					if !compile(s, env.GetModulesDir(), &node.Dep) {
						dependency.Failed = true
					}
					ensureArtifacts(&dependency, node.Dep, env.GetModulesDir())
					moveArtifacts(node.Dep, env.GetModulesDir())
				}
			}
			pkg.Lock.SetInstalled(node.Dep, dependency)
		}
	}
}
