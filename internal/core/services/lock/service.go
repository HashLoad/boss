// Package lock provides functionality for managing package lock files (boss.lock.json).
// It tracks installed dependencies and their versions to ensure consistent installations.
package lock

import (
	"path/filepath"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/ports"
	"github.com/hashload/boss/internal/infra"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/utils"
)

// Service provides lock file management operations.
// It orchestrates domain entities, repositories, and filesystem operations.
type Service struct {
	repo ports.LockRepository
	fs   infra.FileSystem
}

// NewService creates a new lock service.
func NewService(repo ports.LockRepository, fs infra.FileSystem) *Service {
	return &Service{
		repo: repo,
		fs:   fs,
	}
}

// Save persists the lock file.
func (s *Service) Save(lock *domain.PackageLock, packageDir string) error {
	lockPath := filepath.Join(packageDir, consts.FilePackageLock)
	return s.repo.Save(lock, lockPath)
}

// NeedUpdate checks if a dependency needs to be updated.
func (s *Service) NeedUpdate(lock *domain.PackageLock, dep domain.Dependency, version, modulesDir string) bool {
	key := dep.GetKey()
	locked, ok := lock.Installed[key]
	if !ok {
		return true
	}

	// Check if dependency directory exists
	depDir := filepath.Join(modulesDir, dep.Name())
	if !s.fs.Exists(depDir) {
		return true
	}

	// Check if hash changed (files were modified)
	currentHash := utils.HashDir(depDir)
	if locked.Hash != currentHash {
		return true
	}

	// Check if version update is needed
	if domain.NeedsVersionUpdate(locked.Version, version) {
		return true
	}

	// Check if all artifacts exist
	if !s.checkArtifacts(locked, modulesDir) {
		return true
	}

	return false
}

// AddDependency adds a dependency to the lock with computed hash.
func (s *Service) AddDependency(lock *domain.PackageLock, dep domain.Dependency, version, modulesDir string) {
	depDir := filepath.Join(modulesDir, dep.Name())
	hash := utils.HashDir(depDir)

	key := dep.GetKey()
	if existing, ok := lock.Installed[key]; !ok {
		lock.Installed[key] = domain.LockedDependency{
			Name:    dep.Name(),
			Version: version,
			Hash:    hash,
			Changed: true,
			Artifacts: domain.DependencyArtifacts{
				Bin: []string{},
				Bpl: []string{},
				Dcp: []string{},
				Dcu: []string{},
			},
		}
	} else {
		existing.Version = version
		existing.Hash = hash
		lock.Installed[key] = existing
	}
}

// checkArtifacts verifies that all artifacts exist on disk.
func (s *Service) checkArtifacts(locked domain.LockedDependency, modulesDir string) bool {
	checks := []struct {
		folder    string
		artifacts []string
	}{
		{consts.BplFolder, locked.Artifacts.Bpl},
		{consts.BinFolder, locked.Artifacts.Bin},
		{consts.DcpFolder, locked.Artifacts.Dcp},
		{consts.DcuFolder, locked.Artifacts.Dcu},
	}

	for _, check := range checks {
		dir := filepath.Join(modulesDir, check.folder)
		for _, artifact := range check.artifacts {
			artifactPath := filepath.Join(dir, artifact)
			if !s.fs.Exists(artifactPath) {
				return false
			}
		}
	}

	return true
}
