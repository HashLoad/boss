package dcp

import (
	"path/filepath"
	"strings"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/consts"
)

func getRequiresList(pkg *domain.Package, rootLock domain.PackageLock) []string {
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

func getDcpListFromDep(dependency domain.Dependency, lock domain.PackageLock) []string {
	var dcpList []string
	installedMetadata := lock.GetInstalled(dependency)
	for _, dcp := range installedMetadata.Artifacts.Dcp {
		if strings.ToLower(filepath.Ext(dcp)) == consts.FileExtensionDcp {
			dcpList = append(dcpList, dcp)
		}
	}
	return dcpList
}
