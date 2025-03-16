package cmd

import (
	"os"
	"path/filepath"

	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/installer"
	"github.com/hashload/boss/pkg/models"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/utils"
	"github.com/masterminds/semver"
	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"
)

const (
	updated   = 0
	outdated  = 1
	usingMain = 2
)

var showVersion bool
var tree = treeprint.New()

var dependenciesCmd = &cobra.Command{
	Use:     "dependencies",
	Short:   "Print all project dependencies",
	Long:    "Print all project dependencies with or without version control",
	Aliases: []string{"dep", "ls", "list", "ll", "la"},
	Example: `  Listing all dependencies:
  boss dependencies

  Listing all dependencies with version control:
  boss dependencies --version

  List package dependencies:
  boss dependencies <pkg>

  List package dependencies with version control:
  boss dependencies <pkg> --version`,
	Run: func(cmd *cobra.Command, args []string) {
		printDependencies(showVersion)
	},
}

func init() {
	root.AddCommand(dependenciesCmd)
	dependenciesCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "show dependency version")
}

func printDependencies(showVersion bool) {
	pkg, err := models.LoadPackage(false)
	if err != nil {
		if os.IsNotExist(err) {
			msg.Die("boss.json not exists in " + env.GetCurrentDir())
		} else {
			msg.Die("Fail on open dependencies file: %s", err)
		}
	}

	main := tree.AddBranch(pkg.Name + ":")
	deps := pkg.GetParsedDependencies()
	printDeps(nil, deps, pkg.Lock, main, showVersion)
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
	case usingMain:
		output += " using main"
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
			return usingMain
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
