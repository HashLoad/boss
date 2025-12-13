package compiler

import (
	"path/filepath"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/compiler/graphs"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
)

func loadOrderGraph(pkg *domain.Package) *graphs.NodeQueue {
	var graph graphs.GraphItem
	deps := pkg.GetParsedDependencies()
	loadGraph(&graph, nil, deps, nil)
	return graph.Queue(pkg, false)
}

// LoadOrderGraphAll loads the dependency graph for all dependencies.
func LoadOrderGraphAll(pkg *domain.Package) *graphs.NodeQueue {
	var graph graphs.GraphItem
	deps := pkg.GetParsedDependencies()
	loadGraph(&graph, nil, deps, nil)
	return graph.Queue(pkg, true)
}

func loadGraph(graph *graphs.GraphItem, dep *domain.Dependency, deps []domain.Dependency, father *graphs.Node) {
	var localFather *graphs.Node
	if dep != nil {
		localFather = graphs.NewNode(dep)
		graph.AddNode(localFather)
	}

	if father != nil {
		graph.AddEdge(father, localFather)
	}

	for _, dep := range deps {
		pkgModule, err := domain.LoadPackageOther(filepath.Join(env.GetModulesDir(), dep.Name(), consts.FilePackage))
		if err != nil {
			node := graphs.NewNode(&dep)
			graph.AddNode(node)
			if localFather != nil {
				graph.AddEdge(localFather, node)
			}
		} else {
			loadGraph(graph, &dep, pkgModule.GetParsedDependencies(), localFather)
		}
	}
}
