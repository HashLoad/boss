package compiler

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/core/compiler/graphs"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"path/filepath"
)

//Use to print graph
func _() {
	pkg, err := models.LoadPackage(false)
	if err != nil {
		msg.Die(err.Error())
	}
	msg.Info("Compiling order:")
	queue := loadOrderGraph(pkg)
	index := 1
	for {
		if queue.IsEmpty() {
			break
		}
		msg.Info("  %d. %s", index, queue.Dequeue().Value)
		index++
	}
}

func loadOrderGraph(pkg *models.Package) graphs.NodeQueue {
	var graph graphs.GraphItem
	rawDeps := pkg.Dependencies.(map[string]interface{})
	deps := models.GetDependencies(rawDeps)
	loadGraph(&graph, nil, deps, nil)
	return graph.Queue(pkg)
}

func loadGraph(graph *graphs.GraphItem, dep *models.Dependency, deps []models.Dependency, father *graphs.Node) {
	var localFather *graphs.Node
	if dep != nil {
		localFather = graphs.NewNode(dep)
		graph.AddNode(localFather)
	}
	if father != nil {
		graph.AddEdge(father, localFather)
	}

	for _, dep := range deps {
		pkgModule, err := models.LoadPackageOther(filepath.Join(env.GetModulesDir(), dep.GetName(), consts.FilePackage))
		if err != nil {
			node := graphs.NewNode(&dep)
			graph.AddNode(node)
			if localFather != nil {
				graph.AddEdge(localFather, node)
			}
		} else {
			rawDeps := pkgModule.Dependencies.(map[string]interface{})
			deps := models.GetDependencies(rawDeps)
			loadGraph(graph, &dep, deps, localFather)
		}
	}
}
