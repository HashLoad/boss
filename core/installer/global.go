package installer

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"golang.org/x/sys/windows/registry"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GlobalInstall(args []string, pkg *models.Package, lockedVersion bool) {
	EnsureDependencyOfArgs(pkg, args)
	DoInstall(pkg, lockedVersion)
	DoInstallPackages()
}

func find(array []string, value string) int {
	for key, item := range array {
		if item == value {
			return key
		}
	}
	return -1
}

func addPathBpl(ideVersion string) {
	idePath, err := registry.OpenKey(registry.CURRENT_USER, consts.RegistryBasePath+ideVersion+`\Environment Variables`,
		registry.ALL_ACCESS)
	if err != nil {
		msg.Err("Cannot add automatic bpl path dir")
		return
	}
	value, _, err := idePath.GetStringValue("PATH")
	utils.HandleError(err)

	currentPath := filepath.Join(env.GetCurrentDir(), consts.FolderDependencies, consts.BplFolder)

	paths := strings.Split(value, ";")
	if utils.Contains(paths, currentPath) {
		return
	}

	paths = append(paths, currentPath)
	err = idePath.SetStringValue("PATH", strings.Join(paths, ";"))
	utils.HandleError(err)
}

func DoInstallPackages() {
	var ideVersion = env.GetCurrentDelphiVersionFromRegistry()
	var bplDir = filepath.Join(env.GetModulesDir(), consts.BplFolder)

	addPathBpl(ideVersion)

	knowPackages, err := registry.OpenKey(registry.CURRENT_USER, consts.RegistryBasePath+ideVersion+`\Known Packages`,
		registry.ALL_ACCESS)

	if err != nil {
		msg.Err("Cannot open registry to add packages in IDE")
		return
	}

	keyStat, err := knowPackages.Stat()
	utils.HandleError(err)

	keys, err := knowPackages.ReadValueNames(int(keyStat.ValueCount))
	utils.HandleError(err)

	var existingBpls []string

	err = filepath.Walk(bplDir, func(path string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(strings.ToLower(path), ".bpl") {
			return nil
		}

		if !isDesignTimeBpl(path) {
			return nil
		}

		if find(keys, path) == -1 {
			utils.HandleError(knowPackages.SetStringValue(path, path))
		}
		existingBpls = append(existingBpls, path)

		return nil
	})

	for _, key := range keys {
		if find(existingBpls, key) != -1 {
			continue
		}

		if strings.HasPrefix(key, env.GetModulesDir()) {
			err := knowPackages.DeleteValue(key)
			utils.HandleError(err)
		}
	}
}

func isDesignTimeBpl(bplPath string) bool {

	command := exec.Command(filepath.Join(env.GetInternalGlobalDir(), consts.FolderDependencies, consts.BinFolder, consts.BplIdentifierName), bplPath)
	_ = command.Run()
	return command.ProcessState.ExitCode() == 0
}
