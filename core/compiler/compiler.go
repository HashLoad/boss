package compiler

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/core/compiler/graphs"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func Build(pkg *models.Package) {
	buildOrderedPackages(pkg)
	graph := LoadOrderGraphAll(pkg)
	saveLoadOrder(graph)
}

func saveLoadOrder(queue graphs.NodeQueue) {
	var projects = ""
	for {
		if queue.IsEmpty() {
			break
		}
		node := queue.Dequeue()
		dependencyPath := filepath.Join(env.GetModulesDir(), node.Dep.GetName(), consts.FilePackage)
		if dependencyPackage, err := models.LoadPackageOther(dependencyPath); err == nil {
			for _, value := range dependencyPackage.Projects {
				projects += strings.TrimSuffix(filepath.Base(value), filepath.Ext(value)) + consts.FileExtensionBpl + "\n"
			}
		}
	}
	outDir := filepath.Join(env.GetModulesDir(), consts.BplFolder, consts.FileBplOrder)

	utils.HandleError(ioutil.WriteFile(outDir, []byte(projects), os.ModePerm))
}

func buildAllDCUs(dependency models.LockedDependency, pkg *models.Package) {
	modulePath := filepath.Join(env.GetModulesDir(), dependency.Name)
	if pkg != nil {
		modulePath = filepath.Join(modulePath, pkg.MainSrc)
	}
	err := filepath.Walk(modulePath,
		func(path string, info os.FileInfo, err error) error {
			if os.IsNotExist(err) {
				return nil
			} else if !info.IsDir() && filepath.Ext(info.Name()) == ".pas" {

				buildDCU(path)
			}
			return nil
		})
	utils.HandleError(err)
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

		msg.Info("Building %s", node.Dep.GetName())
		dependency.Changed = false
		if dependencyPackage, err := models.LoadPackageOther(filepath.Join(dependencyPath, consts.FilePackage)); err == nil {
			dprojs := dependencyPackage.Projects
			if len(dprojs) > 0 {
				for _, dproj := range dprojs {
					s, _ := filepath.Abs(filepath.Join(env.GetModulesDir(), node.Dep.GetName(), dproj))
					if !compile(s, env.GetModulesDir(), &node.Dep) {
						dependency.Failed = true
					}
				}
				ensureArtifacts(&dependency, node.Dep, env.GetModulesDir())
				moveArtifacts(node.Dep, env.GetModulesDir())
			} else {
				buildAllDCUs(dependency, dependencyPackage)
			}
		} else {
			buildAllDCUs(dependency, nil)
		}
		pkg.Lock.SetInstalled(node.Dep, dependency)

	}
}
