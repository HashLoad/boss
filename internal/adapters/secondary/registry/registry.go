package registryadapter

import (
	"path/filepath"
	"strings"

	"github.com/hashload/boss/pkg/env"
)

func GetDelphiPaths() []string {
	var paths []string
	for _, path := range getDelphiVersionFromRegistry() {
		paths = append(paths, filepath.Dir(path))
	}
	return paths
}

func GetCurrentDelphiVersion() string {
	for version, path := range getDelphiVersionFromRegistry() {
		if strings.HasPrefix(strings.ToLower(path), strings.ToLower(env.GlobalConfiguration().DelphiPath)) {
			return version
		}
	}
	return ""
}
