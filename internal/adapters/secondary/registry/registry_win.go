//go:build windows
// +build windows

package registryadapter

import (
	"os"

	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/utils"
	"golang.org/x/sys/windows/registry"
)

func getDelphiVersionFromRegistry() map[string]string {
	var result = make(map[string]string)

	delphiVersions, err := registry.OpenKey(registry.CURRENT_USER, consts.RegistryBasePath, registry.ALL_ACCESS)
	if err != nil {
		return result
	}

	keyInfo, err := delphiVersions.Stat()
	if err != nil {
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

func getDetectedDelphisFromRegistry() []DelphiInstallation {
	var result []DelphiInstallation

	delphiVersions, err := registry.OpenKey(registry.CURRENT_USER, consts.RegistryBasePath, registry.ALL_ACCESS)
	if err != nil {
		return result
	}
	defer delphiVersions.Close()

	keyInfo, err := delphiVersions.Stat()
	if err != nil {
		return result
	}

	names, err := delphiVersions.ReadSubKeyNames(int(keyInfo.SubKeyCount))
	utils.HandleError(err)

	for _, version := range names {
		delphiInfo, err := registry.OpenKey(registry.CURRENT_USER, consts.RegistryBasePath+version, registry.QUERY_VALUE)
		if err != nil {
			continue
		}

		appPath, _, err := delphiInfo.GetStringValue("App")
		if err == nil && appPath != "" {
			result = append(result, DelphiInstallation{
				Version: version,
				Path:    appPath,
				Arch:    "Win32",
			})
		}

		appPath64, _, err := delphiInfo.GetStringValue("App x64")
		if err == nil && appPath64 != "" {
			result = append(result, DelphiInstallation{
				Version: version,
				Path:    appPath64,
				Arch:    "Win64",
			})
		}
		delphiInfo.Close()
	}
	return result
}
