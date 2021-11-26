package core

import (
	"os"
	"path/filepath"

	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/core/installer"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"github.com/masterminds/semver"
	"github.com/xlab/treeprint"
)

var tree = treeprint.New()

const (
	updated     = 0
	outdated    = 1
	usingMaster = 2
)

func PrintDependencies(showVersion bool) {
	pkg, err := models.LoadPackage(false)
	if err != nil {
		if os.IsNotExist(err) {
			msg.Die("boss.json not exists in " + env.GetCurrentDir())
		} else {
			msg.Die("Fail on open dependencies file: %s", err)
		}
	}

	master := tree.AddBranch(pkg.Name + ":")
	deps := pkg.GetParsedDependencies()
	printDeps(nil, deps, pkg.Lock, master, showVersion)
	print(tree.String())
}

func printDeps(dep *models.Dependency, deps []models.Dependency, lock models.PackageLock, tree treeprint.Tree, showVersion bool) {
	var localTree treeprint.Tree

	if dep != nil {
		localTree = printSingleDependency(dep, lock, tree, showVersion)
	} else {
		localTree = tree
	}

	for _, dep := range deps {
		pkgModule, err := models.LoadPackageOther(filepath.Join(env.GetModulesDir(), dep.GetName(), consts.FilePackage))
		if err != nil {
			printSingleDependency(&dep, lock, localTree, showVersion)
		} else {
			deps := pkgModule.GetParsedDependencies()
			printDeps(&dep, deps, lock, localTree, showVersion)
		}
	}
}

func printSingleDependency(dep *models.Dependency, lock models.PackageLock, tree treeprint.Tree, showVersion bool) treeprint.Tree {
	var output = dep.GetName()

	if showVersion {
		output += "@"
		output += lock.GetInstalled(*dep).Version
	}

	switch isOutdated(*dep, lock.GetInstalled(*dep).Version) {
	case outdated:
		output += " outdated"
		break
	case usingMaster:
		output += " using master"
		break
	}

	return tree.AddBranch(output)
}

func isOutdated(dependency models.Dependency, version string) int {
	installer.GetDependency(dependency)
	info, err := models.RepoData(dependency.GetHashName())
	if err != nil {
		utils.HandleError(err)
	} else {
		locked, err := semver.NewVersion(version)
		if err != nil {
			return usingMaster
		}
		constraint, _ := semver.NewConstraint(dependency.GetVersion())
		for _, value := range info.Versions {
			version, err := semver.NewVersion(value)
			if err == nil && version.GreaterThan(locked) && constraint.Check(version) {
				return outdated
			}
		}
	}
	return updated

}
