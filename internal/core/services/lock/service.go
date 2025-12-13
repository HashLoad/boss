// Package lock provides services for managing package lock files.
// It contains business logic that was previously mixed with domain entities.
package lock

import (
	"path/filepath"
	"strings"

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

// LoadForPackage loads the lock file for a given package.
func (s *Service) LoadForPackage(packageDir, packageName string) (*domain.PackageLock, error) {
	// Handle migration from old format
	oldPath := filepath.Join(packageDir, consts.FilePackageLockOld)
	newPath := filepath.Join(packageDir, consts.FilePackageLock)

	if err := s.repo.MigrateOldFormat(oldPath, newPath); err != nil {
		return nil, err
	}

	lock, err := s.repo.Load(newPath)
	if err != nil {
		// Create new lock if file doesn't exist
		hash := domain.ComputeMD5Hash(packageName)
		return &domain.PackageLock{
			Hash:      hash,
			Installed: make(map[string]domain.LockedDependency),
		}, nil
	}

	return lock, nil
}

// Save persists the lock file.
func (s *Service) Save(lock *domain.PackageLock, packageDir string) error {
	lockPath := filepath.Join(packageDir, consts.FilePackageLock)
	return s.repo.Save(lock, lockPath)
}

// NeedUpdate checks if a dependency needs to be updated.
func (s *Service) NeedUpdate(lock *domain.PackageLock, dep domain.Dependency, version, modulesDir string) bool {
	key := strings.ToLower(dep.Repository)
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

// MarkNeedUpdate marks a dependency as needing update and returns whether update is needed.
func (s *Service) MarkNeedUpdate(lock *domain.PackageLock, dep domain.Dependency, version, modulesDir string) bool {
	needUpdate := s.NeedUpdate(lock, dep, version, modulesDir)

	if needUpdate {
		key := strings.ToLower(dep.Repository)
		if locked, ok := lock.Installed[key]; ok {
			locked.Changed = true
			locked.Failed = false
			lock.Installed[key] = locked
		}
	}

	return needUpdate
}

// AddDependency adds a dependency to the lock with computed hash.
func (s *Service) AddDependency(lock *domain.PackageLock, dep domain.Dependency, version, modulesDir string) {
	depDir := filepath.Join(modulesDir, dep.Name())
	hash := utils.HashDir(depDir)

	key := strings.ToLower(dep.Repository)
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

// SetInstalled updates an installed dependency with computed hash.
func (s *Service) SetInstalled(lock *domain.PackageLock, dep domain.Dependency, locked domain.LockedDependency, modulesDir string) {
	depDir := filepath.Join(modulesDir, dep.Name())
	hash := utils.HashDir(depDir)
	locked.Hash = hash
	lock.Installed[strings.ToLower(dep.Repository)] = locked
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

// CheckArtifactsExist checks if specific artifacts exist.
func (s *Service) CheckArtifactsExist(directory string, artifacts []string) bool {
	for _, artifact := range artifacts {
		path := filepath.Join(directory, artifact)
		if !s.fs.Exists(path) {
			return false
		}
	}
	return true
}
