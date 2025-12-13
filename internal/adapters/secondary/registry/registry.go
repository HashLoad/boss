package registryadapter

import (
	"path/filepath"
	"strings"

	"github.com/hashload/boss/pkg/env"
)

// DelphiInstallation represents a Delphi installation found in the registry.
type DelphiInstallation struct {
	Version string
	Path    string
	Arch    string // Use consts.PlatformWin32 or consts.PlatformWin64
}

// GetDelphiPaths returns a list of paths to Delphi installations.
func GetDelphiPaths() []string {
	var paths []string
	for _, path := range getDelphiVersionFromRegistry() {
		paths = append(paths, filepath.Dir(path))
	}
	return paths
}

// GetDetectedDelphis returns a list of detected Delphi installations.
func GetDetectedDelphis() []DelphiInstallation {
	return getDetectedDelphisFromRegistry()
}

// GetCurrentDelphiVersion returns the version of the currently configured Delphi installation.
func GetCurrentDelphiVersion() string {
	for version, path := range getDelphiVersionFromRegistry() {
		if strings.HasPrefix(strings.ToLower(path), strings.ToLower(env.GlobalConfiguration().DelphiPath)) {
			return version
		}
	}
	return ""
}
