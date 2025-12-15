// Package cli provides command-line interface implementation for Boss.
package cli

import (
	"os"
	"path/filepath"

	"github.com/hashload/boss/pkg/pkgmanager"

	"github.com/Masterminds/semver/v3"
	"github.com/hashload/boss/internal/adapters/secondary/filesystem"
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/cache"
	"github.com/hashload/boss/internal/core/services/installer"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/spf13/cobra"
	"github.com/xlab/treeprint"
)

type dependencyStatus int

const (
	updated dependencyStatus = iota
	outdated
	usingBranch
	branchOutdated
)

// dependenciesCmdRegister registers the dependencies command.
func dependenciesCmdRegister(root *cobra.Command) {
	var showVersion bool

	var dependenciesCmd = &cobra.Command{
		Use:     "dependencies",
		Short:   "Print all project dependencies",
		Long:    "Print all project dependencies with or without version control",
		Aliases: []string{"dep", "ls", "list", "ll", "la", "dependency"},
		Example: `  Listing all dependencies:
  boss dependencies

  Listing all dependencies with version control:
  boss dependencies --version

  List package dependencies:
  boss dependencies <pkg>

  List package dependencies with version control:
  boss dependencies <pkg> --version`,
		Run: func(_ *cobra.Command, _ []string) {
			printDependencies(showVersion)
		},
	}

	root.AddCommand(dependenciesCmd)
	dependenciesCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "show dependency version")
}

// printDependencies prints the dependencies.
func printDependencies(showVersion bool) {
	var tree = treeprint.New()
	pkg, err := pkgmanager.LoadPackage()
	if err != nil {
		if os.IsNotExist(err) {
			msg.Die(consts.FilePackage + " not exists in " + env.GetCurrentDir())
		} else {
			msg.Die("Fail on open dependencies file: %s", err)
		}
	}

	main := tree.AddBranch(pkg.Name + ":")
	deps := pkg.GetParsedDependencies()
	printDeps(nil, deps, pkg.Lock, main, showVersion)
	msg.Info(tree.String())
}

// printDeps prints the dependencies recursively.
func printDeps(dep *domain.Dependency,
	deps []domain.Dependency,
	lock domain.PackageLock,
	tree treeprint.Tree,
	showVersion bool) {
	var localTree treeprint.Tree

	if dep != nil {
		localTree = printSingleDependency(dep, lock, tree, showVersion)
	} else {
		localTree = tree
	}

	for _, dep := range deps {
		pkgModule, err := pkgmanager.LoadPackageOther(filepath.Join(env.GetModulesDir(), dep.Name(), consts.FilePackage))
		if err != nil {
			printSingleDependency(&dep, lock, localTree, showVersion)
		} else {
			subDeps := pkgModule.GetParsedDependencies()
			printDeps(&dep, subDeps, lock, localTree, showVersion)
		}
	}
}

// printSingleDependency prints a single dependency.
func printSingleDependency(
	dep *domain.Dependency,
	lock domain.PackageLock,
	tree treeprint.Tree,
	showVersion bool) treeprint.Tree {
	var output = dep.Name()

	if showVersion {
		output += "@"
		output += lock.GetInstalled(*dep).Version
	}

	status, version := isOutdated(*dep, lock.GetInstalled(*dep).Version)

	switch status {
	case outdated:
		output += " <- outdated (" + version + ")"
	case usingBranch:
		output += " <- branch based"
	case branchOutdated:
		output += " <- branch outdated"
	case updated:
		// Already up to date, no suffix needed
	}

	return tree.AddBranch(output)
}

// isOutdated checks if the dependency is outdated.
func isOutdated(dependency domain.Dependency, version string) (dependencyStatus, string) {
	if err := installer.GetDependency(dependency); err != nil {
		return updated, ""
	}
	cacheService := cache.NewCacheService(filesystem.NewOSFileSystem())
	info, err := cacheService.LoadRepositoryData(dependency.HashName())
	if err != nil {
		return updated, ""
	}
	//TODO: Check if the branch is outdated by comparing the hash
	locked, err := semver.NewVersion(version)
	if err != nil {
		return usingBranch, ""
	}
	constraint, _ := semver.NewConstraint(dependency.GetVersion())
	for _, value := range info.Versions {
		version, err := semver.NewVersion(value)
		if err == nil && version.GreaterThan(locked) && constraint.Check(version) {
			return outdated, version.String()
		}
	}
	return updated, ""
}
