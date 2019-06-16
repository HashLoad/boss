package compiler

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/xlab/treeprint"
	"path/filepath"
)

var tree = treeprint.New()

func dependencyOrder(pkg *models.Package) {
	rawDeps := pkg.Dependencies.(map[string]interface{})
	master := tree.AddBranch(pkg.Name + ":")
	deps := models.GetDependencies(rawDeps)
	printDeps(nil, deps, pkg.Lock, master)
	print(tree.String())
}

func printDeps(dep *models.Dependency, deps []models.Dependency, lock models.PackageLock, tree treeprint.Tree) {
	var localTree treeprint.Tree

	if dep != nil {
		localTree = printSingleDependency(dep, lock, tree)
	} else {
		localTree = tree
	}

	for _, dep := range deps {
		pkgModule, err := models.LoadPackageOther(filepath.Join(env.GetModulesDir(), dep.GetName(), consts.FilePackage))
		if err != nil {
			printSingleDependency(&dep, lock, localTree)
		} else {
			rawDeps := pkgModule.Dependencies.(map[string]interface{})
			deps := models.GetDependencies(rawDeps)
			printDeps(&dep, deps, lock, localTree)
		}
	}
}

func printSingleDependency(dep *models.Dependency, lock models.PackageLock, tree treeprint.Tree) treeprint.Tree {
	var output = dep.GetName()
	output += "@"
	output += dep.GetVersion()
	output += " --> "
	output += lock.GetInstalled(*dep).Version

	return tree.AddBranch(output)
}
