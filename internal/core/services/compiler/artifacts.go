package compiler

import (
	"path/filepath"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/infra"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/msg"
)

// ArtifactService manages build artifacts using dependency injection.
type ArtifactService struct {
	fs infra.FileSystem
}

// NewArtifactService creates a new artifact service.
func NewArtifactService(fs infra.FileSystem) *ArtifactService {
	return &ArtifactService{fs: fs}
}

func (a *ArtifactService) moveArtifacts(dep domain.Dependency, rootPath string) {
	var moduleName = dep.Name()
	a.movePath(filepath.Join(rootPath, moduleName, consts.BplFolder), filepath.Join(rootPath, consts.BplFolder))
	a.movePath(filepath.Join(rootPath, moduleName, consts.DcpFolder), filepath.Join(rootPath, consts.DcpFolder))
	a.movePath(filepath.Join(rootPath, moduleName, consts.BinFolder), filepath.Join(rootPath, consts.BinFolder))
	a.movePath(filepath.Join(rootPath, moduleName, consts.DcuFolder), filepath.Join(rootPath, consts.DcuFolder))
}

func (a *ArtifactService) movePath(oldPath string, newPath string) {
	entries, err := a.fs.ReadDir(oldPath)
	var hasError = false
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				oldFile := filepath.Join(oldPath, entry.Name())
				newFile := filepath.Join(newPath, entry.Name())
				err = a.fs.Rename(oldFile, newFile)
				if err != nil {
					hasError = true
				}
			}
		}
	}
	if !hasError {
		err = a.fs.RemoveAll(oldPath)
		if err != nil && !a.fs.Exists(oldPath) {
			msg.Debug("Non-critical: artifact cleanup failed: %v", err)
		}
	}
}

func (a *ArtifactService) ensureArtifacts(lockedDependency *domain.LockedDependency, dep domain.Dependency, rootPath string) {
	var moduleName = dep.Name()
	lockedDependency.Artifacts.Clean()

	a.collectArtifacts(&lockedDependency.Artifacts.Bpl, filepath.Join(rootPath, moduleName, consts.BplFolder))
	a.collectArtifacts(&lockedDependency.Artifacts.Dcu, filepath.Join(rootPath, moduleName, consts.DcuFolder))
	a.collectArtifacts(&lockedDependency.Artifacts.Bin, filepath.Join(rootPath, moduleName, consts.BinFolder))
	a.collectArtifacts(&lockedDependency.Artifacts.Dcp, filepath.Join(rootPath, moduleName, consts.DcpFolder))
}

func (a *ArtifactService) collectArtifacts(artifactList *[]string, path string) {
	entries, err := a.fs.ReadDir(path)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				*artifactList = append(*artifactList, entry.Name())
			}
		}
	}
}
