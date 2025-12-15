//go:build windows
// +build windows

// Package registryadapter provides Windows registry access for Delphi detection.
package registryadapter

import (
	"os"

	"github.com/hashload/boss/pkg/consts"
	"golang.org/x/sys/windows/registry"
)

// getDelphiVersionFromRegistry returns the delphi version from the registry
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
	if err != nil {
		return result
	}

	for _, value := range names {
		delphiInfo, err := registry.OpenKey(registry.CURRENT_USER, consts.RegistryBasePath+value, registry.QUERY_VALUE)
		if err != nil {
			continue
		}

		appPath, _, err := delphiInfo.GetStringValue("App")
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			continue
		}
		result[value] = appPath

	}
	return result
}

// getDetectedDelphisFromRegistry returns the detected delphi installations from the registry
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
	if err != nil {
		return result
	}

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
				Arch:    consts.PlatformWin32.String(),
			})
		}

		appPath64, _, err := delphiInfo.GetStringValue("App x64")
		if err == nil && appPath64 != "" {
			result = append(result, DelphiInstallation{
				Version: version,
				Path:    appPath64,
				Arch:    consts.PlatformWin64.String(),
			})
		}
		delphiInfo.Close()
	}
	return result
}
