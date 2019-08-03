package core

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/core/installer"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"github.com/masterminds/semver"
	"github.com/xlab/treeprint"
	"os"
	"path/filepath"
)

var tree = treeprint.New()

const (
	updated     = 0
	outdated    = 1
	usingMaster = 2
)

func PrintDependencies() {
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
			deps := pkgModule.GetParsedDependencies()
			printDeps(&dep, deps, lock, localTree)
		}
	}
}

func printSingleDependency(dep *models.Dependency, lock models.PackageLock, tree treeprint.Tree) treeprint.Tree {
	var output = dep.GetName()
	output += "@"
	output += lock.GetInstalled(*dep).Version
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
