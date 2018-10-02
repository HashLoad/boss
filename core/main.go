package core

import (
	"github.com/Masterminds/semver"
	"github.com/hashload/boss/core/git"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	git2 "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"path/filepath"
)

func EnsureDependencies(pkg *models.Package) {
	if pkg.Dependencies == nil {
		return
	}
	rawDeps := pkg.Dependencies.(map[string]interface{})

	deps := models.GetDependencies(rawDeps)

	makeCache(deps)

	ensureModules(deps)
}

func makeCache(deps []models.Dependency) {
	msg.Info("Building cache files..")

	for _, dep := range deps {
		GetDependency(dep)
	}
}

func ensureModules(deps []models.Dependency) {
	msg.Info("Installing modules in project patch")
	for _, dep := range deps {
		repository := OpenRepository(dep)
		versions := git.GetVersions(repository)
		constraints, e := semver.NewConstraint(dep.GetVersion())
		if e != nil {
			msg.Err("Version type not supported! %s", e)
		}
		var bestMatch *plumbing.Reference
		hasMatch := false
		for _, version := range versions {
			short := version.Name().Short()
			newVersion, err := semver.NewVersion(short)
			if err != nil {
				msg.Warn("Erro to parse version %s: '%s' in dependency %s", short, err, dep.Repository)
				continue
			}
			validate, _ := constraints.Validate(newVersion)
			if validate {
				//msg.Debug("Dependency %s with version %s is %s", dep.Repository, newVersion.String(), validate)
				hasMatch = true
				bestMatch = version
			}
		}
		if !hasMatch {
			msg.Die("No candidate to version %s", dep.GetVersion())
		} else {
			msg.Info("For %s using version %s", dep.Repository, bestMatch.Name().Short())
		}

		worktree, _ := repository.Worktree()
		worktree.Filesystem.TempFile(filepath.Join(env.GetCacheDir(), "tmp"), "tpt")
		err := worktree.Checkout(&git2.CheckoutOptions{
			Force: true,
			Hash:  bestMatch.Hash(),
		})
		if err != nil {
			msg.Die("Error on switch to needed version from dependency: %s", dep.Repository)
		}

	}

}
