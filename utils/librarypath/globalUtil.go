package librarypath

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"golang.org/x/sys/windows/registry"
	"strings"
)

const SearchPathRegistry = "Search Path"

func updateGlobalLibraryPath() {
	ideVersion := env.GetCurrentDelphiVersionFromRegistry()
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
		newSplitPaths := GetNewPaths(splitPaths, true)
		newPaths := strings.Join(newSplitPaths, ";")
		err = delphiPlatform.SetStringValue(SearchPathRegistry, newPaths)
		utils.HandleError(err)
	}

}
