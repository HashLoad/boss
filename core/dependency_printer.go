package core

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/core/installer"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/masterminds/semver"
	"github.com/xlab/treeprint"
	"os"
	"path/filepath"
)

var tree = treeprint.New()

func PrintDependencies() {
	pkg, err := models.LoadPackage(false)
	if err != nil {
		if os.IsNotExist(err) {
			msg.Die("boss.json not exists in " + env.GetCurrentDir())
		} else {
			msg.Die("Fail on open dependencies file: %s", err)
		}
	}

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
	output += lock.GetInstalled(*dep).Version
	if isOutdaded(*dep, lock.GetInstalled(*dep).Version) {
		output += " outdated"
	}

	return tree.AddBranch(output)
}

func isOutdaded(dependency models.Dependency, version string) bool {
	info, err := models.RepoData(dependency.GetHashName())
	if err != nil {
		installer.GetDependency(dependency)
		return isOutdaded(dependency, version)
	} else {
		locked := semver.MustParse(version)
		constraint, _ := semver.NewConstraint(dependency.GetVersion())
		for _, value := range info.Versions {
			version, err := semver.NewVersion(value)
			if err == nil && version.GreaterThan(locked) && constraint.Check(version) {
				return true
			}
		}
	}
	return false

}
