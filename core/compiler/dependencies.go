package compiler

import (
	"path/filepath"

	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/core/compiler/graphs"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
)

func loadOrderGraph(pkg *models.Package) *graphs.NodeQueue {
	var graph graphs.GraphItem
	deps := pkg.GetParsedDependencies()
	loadGraph(&graph, nil, deps, nil)
	return graph.Queue(pkg, false)
}
func LoadOrderGraphAll(pkg *models.Package) *graphs.NodeQueue {
	var graph graphs.GraphItem
	deps := pkg.GetParsedDependencies()
	loadGraph(&graph, nil, deps, nil)
	return graph.Queue(pkg, true)
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
			deps := pkgModule.GetParsedDependencies()
			loadGraph(graph, &dep, deps, localFather)
		}
	}
}
