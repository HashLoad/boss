//go:build windows
// +build windows

package registry

import (
	"os"

	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"golang.org/x/sys/windows/registry"
)

func getDelphiVersionFromRegistry() map[string]string {
	var result = make(map[string]string)

	delphiVersions, err := registry.OpenKey(registry.CURRENT_USER, consts.RegistryBasePath, registry.ALL_ACCESS)
	if err != nil {
		msg.Err("Cannot open registry to IDE version")
		return result
	}

	keyInfo, err := delphiVersions.Stat()
	if err != nil {
		msg.Err("Cannot open Delphi registry")
		return result
	}
	names, err := delphiVersions.ReadSubKeyNames(int(keyInfo.SubKeyCount))
	utils.HandleError(err)

	for _, value := range names {
		delphiInfo, err := registry.OpenKey(registry.CURRENT_USER, consts.RegistryBasePath+value, registry.QUERY_VALUE)
		utils.HandleError(err)

		appPath, _, err := delphiInfo.GetStringValue("App")
		if os.IsNotExist(err) {
			continue
		}
		utils.HandleError(err)
		result[value] = appPath

	}
	return result
}
