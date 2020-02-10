package dcc32

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils"
	"golang.org/x/sys/windows/registry"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GetDcc32DirByCmd() []string {
	command := exec.Command("where", "dcc32")
	output, err := command.Output()
	if err != nil {
		msg.Warn("dcc32 not found")
	}
	outputStr := strings.ReplaceAll(string(output), "\t", "")
	outputStr = strings.ReplaceAll(outputStr, "\r", "")
	if strings.HasSuffix(outputStr, "\n") {
		outputStr = outputStr[0 : len(outputStr)-1]
	}
	if len(outputStr) == 0 {
		return []string{}
	}
	installations := strings.Split(outputStr, "\n")
	for key, value := range installations {
		installations[key] = filepath.Dir(value)
	}
	return installations
}

func GetDelphiVersionFromRegistry() map[string]string {
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

func GetDelphiVersionNumberName(currentPath string) string {
	for version, path := range GetDelphiVersionFromRegistry() {
		if strings.HasPrefix(strings.ToLower(path), strings.ToLower(currentPath)) {
			return version
		}
	}
	return ""
}

func GetDelphiPathsByRegistry() []string {
	var paths []string
	for _, path := range GetDelphiVersionFromRegistry() {
		paths = append(paths, filepath.Dir(path))
	}
	return paths
}
