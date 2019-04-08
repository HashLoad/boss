package installer

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/models"
	"github.com/hashload/boss/msg"
	"golang.org/x/sys/windows/registry"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GlobalInstall(args []string) {
	pkg, _ := models.LoadPackage(true)
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

func handleError(err error) {
	if err != nil {
		msg.Err(err.Error())
	}
}

func DoInstallPackages() {
	var ideVersion = env.GetDelphiVersionFromRegisty()
	var bplDir = filepath.Join(env.GetModulesDir(), consts.BplFolder)

	knowPackages, err := registry.OpenKey(registry.CURRENT_USER, `Software\Embarcadero\BDS\`+ideVersion+`\Known Packages`,
		registry.ALL_ACCESS)

	if err != nil {
		msg.Err("Cannot open registry to add packages in IDE")
	}

	keyStat, err := knowPackages.Stat()
	handleError(err)

	keys, err := knowPackages.ReadValueNames(int(keyStat.ValueCount))
	handleError(err)

	existingBpls := []string{}

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
			handleError(knowPackages.SetStringValue(path, consts.BossInstalled+path))
		}
		existingBpls = append(existingBpls, path)

		return nil
	})

	for _, key := range keys {
		if Find(existingBpls, key) != -1 {
			continue
		}

		stringValue, _, err := knowPackages.GetStringValue(key)
		if err != nil {
			msg.Warn("Ignoring: %s", err.Error())
		}
		if strings.HasPrefix(stringValue, consts.BossInstalled) {
			err := knowPackages.DeleteValue(key)
			handleError(err)
		}
	}
}

func isDesignTimeBpl(bplPath string) bool {

	command := exec.Command(filepath.Join(env.GetInternalGlobalDir(), consts.FolderDependencies, consts.BinFolder, consts.BplIdentifierName), bplPath)
	out, _ := command.Output()
	return command.ProcessState.ExitCode() == 0 || strings.HasPrefix(strings.ToLower(string(out)), "design")
}
