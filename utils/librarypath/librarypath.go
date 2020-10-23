package librarypath

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/utils"
)

func UpdateLibraryPath(pkg *models.Package) {
	if env.Global {
		updateGlobalLibraryPath()
	} else {
		updateDprojLibraryPath(pkg)
	}

}

func cleanPath(paths []string, fullPath bool) []string {
	prefix := env.GetModulesDir()
	var processedPaths []string
	if !fullPath {
		prefix, _ = filepath.Rel(env.GetCurrentDir(), prefix)
	}

	for key := 0; key < len(paths); key++ {
		if strings.HasPrefix(paths[key], prefix) {
			continue
		}
		if !utils.Contains(processedPaths, paths[key]) {
			processedPaths = append(processedPaths, paths[key])
		}
	}
	return processedPaths
}

func GetNewPaths(paths []string, fullPath bool) []string {
	paths = cleanPath(paths, fullPath)
	var path = env.GetModulesDir()

	matches, _ := ioutil.ReadDir(path)

	for _, value := range matches {

		var packagePath = filepath.Join(path, value.Name(), consts.FilePackage)
		if _, err := os.Stat(packagePath); !os.IsNotExist(err) {

			other, _ := models.LoadPackageOther(packagePath)
			paths = getNewPathsFromDir(filepath.Join(path, value.Name(), other.MainSrc), paths, fullPath)

		} else {
			paths = getNewPathsFromDir(filepath.Join(path, value.Name()), paths, fullPath)
		}
	}
	return paths
}

func getDefaultPath(fullPath bool) []string {
	var paths []string

	if !fullPath {
		fullPath := filepath.Join(env.GetCurrentDir(), consts.FolderDependencies, consts.DcpFolder)

		dir, err := filepath.Rel(env.GetCurrentDir(), fullPath)
		if err == nil {
			paths = append(paths, dir)
		}

		fullPath = filepath.Join(env.GetCurrentDir(), consts.FolderDependencies, consts.DcuFolder)
		dir, err = filepath.Rel(env.GetCurrentDir(), fullPath)
		if err == nil {
			paths = append(paths, dir)
		}
	} else {
		paths = append(paths, filepath.Join(env.GetCurrentDir(), consts.FolderDependencies, consts.DcpFolder))
		paths = append(paths, filepath.Join(env.GetCurrentDir(), consts.FolderDependencies, consts.DcuFolder))
	}

	if isLazarus() {
		return paths
	}

	return append(paths, "$(DCC_UnitSearchPath)")
}

func cleanEmpty(paths []string) []string {
	for index, value := range paths {

		if value == "" {
			paths = append(paths[:index], paths[index+1:]...)
		}
	}
	return paths
}

func getNewPathsFromDir(path string, paths []string, fullPath bool) []string {
	_, e := os.Stat(path)
	if os.IsNotExist(e) {
		return paths
	}

	_ = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		matched, _ := regexp.MatchString(consts.RegexArtifacts, info.Name())
		if matched {
			dir, _ := filepath.Split(path)
			if !fullPath {
				dir, _ = filepath.Rel(env.GetCurrentDir(), dir)
			}
			if !utils.Contains(paths, dir) {
				paths = append(paths, dir)
			}
		}
		return nil
	})

	for _, path := range getDefaultPath(fullPath) {
		if !utils.Contains(paths, path) {
			paths = append(paths, path)
		}
	}
	return cleanEmpty(paths)
}
