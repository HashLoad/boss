package installer

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"github.com/hashload/boss/utils/librarypath"
	"golang.org/x/sys/windows/registry"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GlobalInstall(args []string, pkg *models.Package) {
	EnsureDependencyOfArgs(pkg, args)
	DoInstall(pkg)
	DoInstallPackages()
}

func Find(array []string, value string) int {
	for key, item := range array {
		if item == value {
			return key
		}
	}
	return -1
}

func addPathBpl(ideVersion string) {
	idePath, err := registry.OpenKey(registry.CURRENT_USER, `Software\Embarcadero\BDS\`+ideVersion+`\Environment Variables`,
		registry.ALL_ACCESS)
	if err != nil {
		msg.Err("Cannot add automatic bpl path dir")
		return
	}
	value, _, err := idePath.GetStringValue("PATH")
	utils.HandleError(err)

	currentPath := filepath.Join(env.GetCurrentDir(), consts.FolderDependencies, consts.BplFolder)

	paths := strings.Split(value, ";")
	if librarypath.Contains(paths, currentPath) {
		return
	}

	paths = append(paths, currentPath)
	err = idePath.SetStringValue("PATH", strings.Join(paths, ";"))
	utils.HandleError(err)
}

func DoInstallPackages() {
	var ideVersion = env.GetCurrentDelphiVersionFromRegisty()
	var bplDir = filepath.Join(env.GetModulesDir(), consts.BplFolder)

	addPathBpl(ideVersion)

	knowPackages, err := registry.OpenKey(registry.CURRENT_USER, `Software\Embarcadero\BDS\`+ideVersion+`\Known Packages`,
		registry.ALL_ACCESS)

	if err != nil {
		msg.Err("Cannot open registry to add packages in IDE")
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

		if Find(keys, path) == -1 {
			utils.HandleError(knowPackages.SetStringValue(path, path))
		}
		existingBpls = append(existingBpls, path)

		return nil
	})

	for _, key := range keys {
		if Find(existingBpls, key) != -1 {
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
	out, _ := command.Output()
	return strings.HasPrefix(strings.ToLower(string(out)), "design")
}
