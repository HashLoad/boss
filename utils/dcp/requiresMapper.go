package dcp

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/models"
	"path/filepath"
	"strings"
)

func getRequiresList(pkg *models.Package, rootLock models.PackageLock) []string {
	if pkg == nil {
		return []string{}
	}
	dependencies := pkg.GetParsedDependencies()

	if len(dependencies) == 0 {
		return []string{}
	}

	var dcpList []string

	for _, dependency := range dependencies {
		dcpList = append(dcpList, getDcpListFromDep(dependency, rootLock)...)
	}

	for key, dcp := range dcpList {
		dcp = dcp[0 : len(dcp)-len(filepath.Ext(dcp))]
		dcpList[key] = dcp
	}

	return dcpList
}

func getDcpListFromDep(dependency models.Dependency, lock models.PackageLock) []string {
	var dcpList []string
	installedMetadata := lock.GetInstalled(dependency)
	for _, dcp := range installedMetadata.Artifacts.Dcp {
		if strings.ToLower(filepath.Ext(dcp)) == consts.FileExtensionDcp {
			dcpList = append(dcpList, dcp)
		}
	}
	return dcpList
}
