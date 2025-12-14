package compiler

import (
	"path/filepath"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/pkgmanager"
)

func loadOrderGraph(pkg *domain.Package) *domain.NodeQueue {
	var graph domain.GraphItem
	deps := pkg.GetParsedDependencies()
	loadGraph(&graph, nil, deps, nil)
	return graph.Queue(pkg, false)
}

// LoadOrderGraphAll loads the dependency graph for all dependencies.
func LoadOrderGraphAll(pkg *domain.Package) *domain.NodeQueue {
	var graph domain.GraphItem
	deps := pkg.GetParsedDependencies()
	loadGraph(&graph, nil, deps, nil)
	return graph.Queue(pkg, true)
}

func loadGraph(graph *domain.GraphItem, dep *domain.Dependency, deps []domain.Dependency, father *domain.Node) {
	var localFather *domain.Node
	if dep != nil {
		localFather = domain.NewNode(dep)
		graph.AddNode(localFather)
	}

	if father != nil {
		graph.AddEdge(father, localFather)
	}

	for _, dep := range deps {
		pkgModule, err := pkgmanager.LoadPackageOther(filepath.Join(env.GetModulesDir(), dep.Name(), consts.FilePackage))
		if err != nil {
			node := domain.NewNode(&dep)
			graph.AddNode(node)
			if localFather != nil {
				graph.AddEdge(localFather, node)
			}
		} else {
			loadGraph(graph, &dep, pkgModule.GetParsedDependencies(), localFather)
		}
	}
}
