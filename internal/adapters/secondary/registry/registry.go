package registryadapter

import (
	"path/filepath"
	"strings"

	"github.com/hashload/boss/pkg/env"
)

type DelphiInstallation struct {
	Version string
	Path    string
	Arch    string // Use consts.PlatformWin32 or consts.PlatformWin64
}

func GetDelphiPaths() []string {
	var paths []string
	for _, path := range getDelphiVersionFromRegistry() {
		paths = append(paths, filepath.Dir(path))
	}
	return paths
}

func GetDetectedDelphis() []DelphiInstallation {
	return getDetectedDelphisFromRegistry()
}

func GetCurrentDelphiVersion() string {
	for version, path := range getDelphiVersionFromRegistry() {
		if strings.HasPrefix(strings.ToLower(path), strings.ToLower(env.GlobalConfiguration().DelphiPath)) {
			return version
		}
	}
	return ""
}
