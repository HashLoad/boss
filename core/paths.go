package core

import (
	"github.com/Masterminds/semver"
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/core/git"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	git2 "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"io/ioutil"
	"os"
	"path/filepath"
)

var processed = []string{consts.BplFolder, consts.BinFolder, consts.DcpFolder, consts.DcuFolder}

func EnsureDependencies(pkg *models.Package) {
	if pkg.Dependencies == nil {
		return
	}
	rawDeps := pkg.Dependencies.(map[string]interface{})

	deps := models.GetDependencies(rawDeps)

	makeCache(deps)

	ensureModules(pkg, deps)

	processOthers()
}

func makeCache(deps []models.Dependency) {
	msg.Info("Building cache files..")
	for _, dep := range deps {
		GetDependency(dep)
	}
}

func ensureModules(pkg *models.Package, deps []models.Dependency) {
	msg.Info("Installing modules")
	for _, dep := range deps {
		msg.Info("Processing dependency: %s", dep.GetName())
		repository := OpenRepository(dep)
		versions := git.GetVersions(repository)
		constraints, e := semver.NewConstraint(dep.GetVersion())
		if e != nil {
			msg.Err("  Version type not supported! %s", e)
		}
		var bestMatch *plumbing.Reference
		var bestVersion *semver.Version
		hasMatch := false
		for _, version := range versions {
			short := version.Name().Short()
			newVersion, err := semver.NewVersion(short)
			if err != nil {
				continue
			}
			if constraints.Check(newVersion) {
				hasMatch = true
				if bestVersion == nil || newVersion.GreaterThan(bestVersion) {
					bestMatch = version
					bestVersion = newVersion
				}
			}
		}

		var referenceName plumbing.ReferenceName

		if !hasMatch {
			msg.Warn("  No candidate to version for %s. Using master branch", dep.GetVersion())
			if masterReference := git.GetMaster(repository); masterReference != nil {
				referenceName = plumbing.NewBranchReferenceName(masterReference.Name)
			}
		} else {
			msg.Info("  Detected semantic version. For %s using version %s", dep.Repository, bestMatch.Name().Short())
			referenceName = bestMatch.Name()
			if dep.GetVersion() == consts.MinimalDependencyVersion {
				pkg.Dependencies.(map[string]interface{})[dep.Repository] = "^" + bestMatch.Name().Short()
			}
		}
		worktree, _ := repository.Worktree()

		err := worktree.Checkout(&git2.CheckoutOptions{
			Force:  true,
			Branch: referenceName,
		})
		if err != nil {
			msg.Die("  Error on switch to needed version from dependency: %s\n%s", dep.Repository, err)
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
		} else {
			processed = append(processed, info.Name())
		}
		msg.Info("Processing module: %s", info.Name())

		fileName := filepath.Join(env.GetModulesDir(), info.Name(), consts.FilePackage)

		_, i := os.Stat(fileName)
		if os.IsNotExist(i) {
			msg.Warn("  boss.json not exists in %s", info.Name())
		}

		if packageOther, e := models.LoadPackageOther(fileName); e != nil {
			if os.IsNotExist(e) {
				continue
			}
			msg.Err("  Error on try load package %s: %s", fileName, e)
		} else {
			EnsureDependencies(packageOther)
		}
	}
}
