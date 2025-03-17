package librarypath

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"slices"

	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/models"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/utils"
)

func UpdateLibraryPath(pkg *models.Package) {
	if env.GetGlobal() {
		updateGlobalLibraryPath()
	} else {
		updateDprojLibraryPath(pkg)
		updateGlobalBrowsingPath(pkg)
	}
}

func cleanPath(paths []string, fullPath bool) []string {
	prefix := env.GetModulesDir()
	var processedPaths []string
	if !fullPath {
		prefix, _ = filepath.Rel(env.GetCurrentDir(), prefix)
	}

	for key := range paths {
		if strings.HasPrefix(paths[key], prefix) {
			continue
		}
		if !utils.Contains(processedPaths, paths[key]) {
			processedPaths = append(processedPaths, paths[key])
		}
	}
	return processedPaths
}

func GetNewBrowsingPaths(paths []string, fullPath bool, rootPath string, setReadOnly bool) []string {
	paths = cleanPath(paths, fullPath)
	var path = env.GetModulesDir()

	matches, _ := os.ReadDir(path)

	for _, value := range matches {
		paths = processBrowsingPath(value, paths, path, fullPath, rootPath, setReadOnly)
	}
	return paths
}

func processBrowsingPath(
	value os.DirEntry,
	paths []string,
	basePath string,
	fullPath bool,
	rootPath string,
	setReadOnly bool,
) []string {
	var packagePath = filepath.Join(basePath, value.Name(), consts.FilePackage)
	if _, err := os.Stat(packagePath); !os.IsNotExist(err) {
		other, _ := models.LoadPackageOther(packagePath)
		if other.BrowsingPath != "" {
			dir := filepath.Join(basePath, value.Name(), other.BrowsingPath)
			paths = getNewBrowsingPathsFromDir(dir, paths, fullPath, rootPath)
			if setReadOnly {
				setReadOnlyProperty(dir)
			}
		}
	}
	return paths
}

func setReadOnlyProperty(dir string) {
	readonlybat := filepath.Join(dir, "readonly.bat")
	readFileStr := fmt.Sprintf(`attrib +r "%s" /s /d`, filepath.Join(dir, "*"))
	err := os.WriteFile(readonlybat, []byte(readFileStr), 0600)
	if err != nil {
		msg.Warn("  - error on create build file")
	}

	cmd := exec.Command(readonlybat)

	_, err = cmd.Output()
	if err != nil {
		msg.Err("  - Failed to set readonly property to folder", dir, " - ", err)
	} else {
		os.Remove(readonlybat)
	}
}

func GetNewPaths(paths []string, fullPath bool, rootPath string) []string {
	paths = cleanPath(paths, fullPath)
	var path = env.GetModulesDir()

	matches, _ := os.ReadDir(path)

	for _, value := range matches {
		var packagePath = filepath.Join(path, value.Name(), consts.FilePackage)
		if _, err := os.Stat(packagePath); !os.IsNotExist(err) {
			other, _ := models.LoadPackageOther(packagePath)
			paths = getNewPathsFromDir(filepath.Join(path, value.Name(), other.MainSrc), paths, fullPath, rootPath)
		} else {
			paths = getNewPathsFromDir(filepath.Join(path, value.Name()), paths, fullPath, rootPath)
		}
	}
	return paths
}

func getDefaultPath(fullPath bool, rootPath string) []string {
	var paths []string

	if !fullPath {
		fullPath := filepath.Join(env.GetCurrentDir(), consts.FolderDependencies, consts.DcpFolder)

		dir, err := filepath.Rel(rootPath, fullPath)
		if err == nil {
			paths = append(paths, dir)
		}

		fullPath = filepath.Join(env.GetCurrentDir(), consts.FolderDependencies, consts.DcuFolder)
		dir, err = filepath.Rel(rootPath, fullPath)
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
			paths = slices.Delete(paths, index, index+1)
		}
	}
	return paths
}

func getNewBrowsingPathsFromDir(path string, paths []string, fullPath bool, rootPath string) []string {
	_, e := os.Stat(path)
	if os.IsNotExist(e) {
		return paths
	}

	_ = filepath.Walk(path, func(path string, info os.FileInfo, _ error) error {
		matched, _ := regexp.MatchString(consts.RegexArtifacts, info.Name())
		if matched {
			dir, _ := filepath.Split(path)

			if !fullPath {
				dir, _ = filepath.Rel(rootPath, dir)
			}
			if !utils.Contains(paths, dir) {
				paths = append(paths, dir)
			}
		}
		return nil
	})
	return cleanEmpty(paths)
}

func getNewPathsFromDir(path string, paths []string, fullPath bool, rootPath string) []string {
	_, e := os.Stat(path)
	if os.IsNotExist(e) {
		return paths
	}

	_ = filepath.Walk(path, func(path string, info os.FileInfo, _ error) error {
		matched, _ := regexp.MatchString(consts.RegexArtifacts, info.Name())
		if matched {
			dir, _ := filepath.Split(path)
			if !fullPath {
				dir, _ = filepath.Rel(rootPath, dir)
			}
			if !utils.Contains(paths, dir) {
				paths = append(paths, dir)
			}
		}
		return nil
	})

	for _, path := range getDefaultPath(fullPath, rootPath) {
		if !strings.HasPrefix(path, "$") {
			if !utils.Contains(paths, path) {
				paths = append(paths, path)
			}
		}
	}
	return cleanEmpty(paths)
}
