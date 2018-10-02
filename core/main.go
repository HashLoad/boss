package core

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver"
	"github.com/hashload/boss/core/git"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	git2 "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

var processed = make([]string, 0)

var processedOld = -1

func EnsureDependencies(pkg *models.Package) {
	if pkg.Dependencies == nil {
		return
	}
	rawDeps := pkg.Dependencies.(map[string]interface{})

	deps := models.GetDependencies(rawDeps)

	makeCache(deps)

	ensureModules(deps)

	processOthers()
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
		msg.Info("Processing dependency: %s", dep.GetName())
		repository := OpenRepository(dep)
		versions := git.GetVersions(repository)
		constraints, e := semver.NewConstraint(dep.GetVersion())
		if e != nil {
			msg.Err("\tVersion type not supported! %s", e)
		}
		var bestMatch *plumbing.Reference
		var bestVersion *semver.Version
		hasMatch := false
		for _, version := range versions {
			short := version.Name().Short()
			newVersion, err := semver.NewVersion(short)
			if err != nil {
				msg.Warn("\tErro to parse version %s: '%s' in dependency %s", short, err, dep.Repository)
				continue
			}
			if constraints.Check(newVersion) {
				//msg.Info("Dependency %s with version %s", dep.Repository, newVersion.String())
				hasMatch = true
				if bestVersion == nil || newVersion.GreaterThan(bestVersion) {
					bestMatch = version
					bestVersion = newVersion
				}
			}
		}
		if !hasMatch {
			msg.Die("\tNo candidate to version %s", dep.GetVersion())
		} else {
			msg.Info("\tFor %s using version %s", dep.Repository, bestMatch.Name().Short())
		}

		worktree, _ := repository.Worktree()
		worktree.Filesystem.TempFile(filepath.Join(env.GetCacheDir(), "tmp"), "tpt")
		err := worktree.Checkout(&git2.CheckoutOptions{
			Force: true,
			Hash:  bestMatch.Hash(),
		})
		if err != nil {
			msg.Die("\tError on switch to needed version from dependency: %s", dep.Repository)
		}
	}
}

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func processOthers() {
	if len(processed) > processedOld {
		processedOld = len(processed)
		infos, e := ioutil.ReadDir(env.GetModulesDir())
		if e != nil {
			msg.Err("Error on try load dir of modules: %s", e)
		}

		for _, info := range infos {
			if !info.IsDir() {
				continue
			}
			if contains(processed, info.Name()) {
				continue
			}
			msg.Info("Proccessing module: %s", info.Name())

			fileName := filepath.Join(env.GetModulesDir(), info.Name(), "boss.json")

			_, i := os.Stat(fileName)
			if os.IsNotExist(i) {
				msg.Warn("boss.json not exists in %s", info.Name())
			}

			if packageOther, e := models.LoadPackageOther(fileName); e != nil {
				msg.Err("Error on try load package %s: %s", fileName, e)
			} else {
				EnsureDependencies(packageOther)
			}
		}

	}
}
