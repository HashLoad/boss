// Package installer provides Windows global installation support.
//go:build windows

package installer

import (
	"io/fs"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	bossRegistry "github.com/hashload/boss/internal/adapters/secondary/registry"
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/utils"
	"golang.org/x/sys/windows/registry"
)

// GlobalInstall installs dependencies globally (Windows implementation).
func GlobalInstall(config env.ConfigProvider, args []string, pkg *domain.Package, lockedVersion bool, noSave bool) {
	// TODO noSave
	EnsureDependency(pkg, args)
	if err := DoInstall(config, InstallOptions{
		Args:          args,
		LockedVersion: lockedVersion,
		NoSave:        noSave,
	}, pkg); err != nil {
		msg.Die("❌ %s", err)
	}
	doInstallPackages()
}

func addPathBpl(ideVersion string) {
	idePath, err := registry.OpenKey(registry.CURRENT_USER, consts.RegistryBasePath+ideVersion+`\Environment Variables`,
		registry.ALL_ACCESS)
	if err != nil {
		msg.Err("❌ Cannot add automatic bpl path dir")
		return
	}
	value, _, err := idePath.GetStringValue("PATH")
	if err != nil {
		msg.Warn("⚠️ Failed to get PATH environment variable: %v", err)
		return
	}

	currentPath := filepath.Join(env.GetCurrentDir(), consts.FolderDependencies, consts.BplFolder)

	paths := strings.Split(value, ";")
	if utils.Contains(paths, currentPath) {
		return
	}

	paths = append(paths, currentPath)
	err = idePath.SetStringValue("PATH", strings.Join(paths, ";"))
	if err != nil {
		msg.Warn("⚠️ Failed to update PATH environment variable: %v", err)
	}
}

func doInstallPackages() {
	var ideVersion = bossRegistry.GetCurrentDelphiVersion()
	var bplDir = filepath.Join(env.GetModulesDir(), consts.BplFolder)

	addPathBpl(ideVersion)

	knowPackages, err := registry.OpenKey(registry.CURRENT_USER, consts.RegistryBasePath+ideVersion+`\Known Packages`,
		registry.ALL_ACCESS)

	if err != nil {
		msg.Err("❌ Cannot open registry to add packages in IDE")
		return
	}

	keyStat, err := knowPackages.Stat()
	if err != nil {
		msg.Warn("⚠️ Failed to stat Known Packages registry key: %v", err)
		return
	}

	keys, err := knowPackages.ReadValueNames(int(keyStat.ValueCount))
	if err != nil {
		msg.Warn("⚠️ Failed to read Known Packages values: %v", err)
		return
	}

	var existingBpls []string

	_ = filepath.WalkDir(bplDir, func(path string, info fs.DirEntry, err error) error {
		if info == nil || info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(strings.ToLower(path), consts.FileExtensionBpl) {
			return nil
		}

		if !isDesignTimeBpl(path) {
			return nil
		}

		if !slices.Contains(keys, path) {
			if err := knowPackages.SetStringValue(path, path); err != nil {
				msg.Debug("Failed to register BPL %s: %v", path, err)
			}
		}
		existingBpls = append(existingBpls, path)

		return nil
	})

	for _, key := range keys {
		if slices.Contains(existingBpls, key) {
			continue
		}

		if strings.HasPrefix(key, env.GetModulesDir()) {
			err := knowPackages.DeleteValue(key)
			if err != nil {
				msg.Debug("Failed to delete obsolete BPL registry entry %s: %v", key, err)
			}
		}
	}
}

func isDesignTimeBpl(bplPath string) bool {

	command := exec.Command(filepath.Join(env.GetInternalGlobalDir(), consts.FolderDependencies, consts.BinFolder, consts.BplIdentifierName), bplPath)
	_ = command.Run()
	return command.ProcessState.ExitCode() == 0
}
