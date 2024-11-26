//go:build windows
// +build windows

package librarypath

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"golang.org/x/sys/windows/registry"

	bossRegistry "github.com/hashload/boss/core/registry"
)

const SearchPathRegistry = "Search Path"
const BrowsingPathRegistry = "Browsing Path"

func updateGlobalLibraryPath() {
	ideVersion := bossRegistry.GetCurrentDelphiVersion()
	if ideVersion == "" {
		msg.Err("Version not found for path %s", env.GlobalConfiguration.DelphiPath)
	}
	library, err := registry.OpenKey(registry.CURRENT_USER, consts.RegistryBasePath+ideVersion+`\Library`, registry.ALL_ACCESS)

	if err != nil {
		msg.Err(`Registry path` + consts.RegistryBasePath + ideVersion + `\Library not exists`)
		return
	}

	libraryInfo, err := library.Stat()
	if err != nil {
		msg.Err(err.Error())
		return
	}
	platforms, err := library.ReadSubKeyNames(int(libraryInfo.SubKeyCount))
	if err != nil {
		msg.Err("No platform found for delphi " + ideVersion)
		return
	}

	for _, platform := range platforms {
		delphiPlatform, err := registry.OpenKey(registry.CURRENT_USER, consts.RegistryBasePath+ideVersion+`\Library\`+platform, registry.ALL_ACCESS)
		utils.HandleError(err)
		paths, _, err := delphiPlatform.GetStringValue(SearchPathRegistry)
		if err != nil {
			msg.Debug("Failed to update library path from platform %s with delphi %s", platform, ideVersion)
			continue
		}

		splitPaths := strings.Split(paths, ";")
		newSplitPaths := GetNewPaths(splitPaths, true, env.GetCurrentDir())
		newPaths := strings.Join(newSplitPaths, ";")
		err = delphiPlatform.SetStringValue(SearchPathRegistry, newPaths)
		utils.HandleError(err)
	}

}

func updateGlobalBrowsingByProject(dprojName string, setReadOnly bool) {
	ideVersion := bossRegistry.GetCurrentDelphiVersion()
	if ideVersion == "" {
		msg.Err("Version not found for path %s", env.GlobalConfiguration.DelphiPath)
	}
	library, err := registry.OpenKey(registry.CURRENT_USER, consts.RegistryBasePath+ideVersion+`\Library`, registry.ALL_ACCESS)

	if err != nil {
		msg.Err(`Registry path` + consts.RegistryBasePath + ideVersion + `\Library not exists`)
		return
	}

	libraryInfo, err := library.Stat()
	if err != nil {
		msg.Err(err.Error())
		return
	}
	platforms, err := library.ReadSubKeyNames(int(libraryInfo.SubKeyCount))
	if err != nil {
		msg.Err("No platform found for delphi " + ideVersion)
		return
	}

	for _, platform := range platforms {
		delphiPlatform, err := registry.OpenKey(registry.CURRENT_USER, consts.RegistryBasePath+ideVersion+`\Library\`+platform, registry.ALL_ACCESS)
		utils.HandleError(err)
		paths, _, err := delphiPlatform.GetStringValue(BrowsingPathRegistry)
		if err != nil {
			msg.Debug("Failed to update library path from platform %s with delphi %s", platform, ideVersion)
			continue
		}

		splitPaths := strings.Split(paths, ";")
		rootPath := filepath.Join(env.GetCurrentDir(), path.Dir(dprojName))
		newSplitPaths := GetNewBrowsingPaths(splitPaths, false, rootPath, setReadOnly)
		newPaths := strings.Join(newSplitPaths, ";")
		err = delphiPlatform.SetStringValue(BrowsingPathRegistry, newPaths)
		utils.HandleError(err)
	}
}
