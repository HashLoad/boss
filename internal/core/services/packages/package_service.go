// Package packages provides services for package operations.
package packages

import (
	"fmt"
	"path/filepath"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/ports"
	"github.com/hashload/boss/pkg/env"
)

// PackageService handles package operations using repositories.
type PackageService struct {
	packageRepo ports.PackageRepository
	lockRepo    ports.LockRepository
}

// NewPackageService creates a new package service.
func NewPackageService(packageRepo ports.PackageRepository, lockRepo ports.LockRepository) *PackageService {
	return &PackageService{
		packageRepo: packageRepo,
		lockRepo:    lockRepo,
	}
}

// LoadCurrent loads the current project's package file (boss.json).
func (s *PackageService) LoadCurrent() (*domain.Package, error) {
	bossFile := env.GetBossFile()

	if !s.packageRepo.Exists(bossFile) {
		// Return empty package if file doesn't exist
		pkg := domain.NewPackage()
		pkg.Lock = s.loadOrCreateLock(bossFile)
		return pkg, nil
	}

	pkg, err := s.packageRepo.Load(bossFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load package from %s: %w", bossFile, err)
	}

	pkg.Lock = s.loadOrCreateLock(bossFile)
	return pkg, nil
}

// Load loads a package from a specific path.
func (s *PackageService) Load(packagePath string) (*domain.Package, error) {
	pkg, err := s.packageRepo.Load(packagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load package from %s: %w", packagePath, err)
	}

	pkg.Lock = s.loadOrCreateLock(packagePath)
	return pkg, nil
}

// Save saves a package to a specific path.
func (s *PackageService) Save(pkg *domain.Package, packagePath string) error {
	if err := s.packageRepo.Save(pkg, packagePath); err != nil {
		return fmt.Errorf("failed to save package to %s: %w", packagePath, err)
	}
	return nil
}

// SaveCurrent saves the current project's package file.
func (s *PackageService) SaveCurrent(pkg *domain.Package) error {
	return s.Save(pkg, env.GetBossFile())
}

// SaveLock saves the lock file for a package.
func (s *PackageService) SaveLock(pkg *domain.Package, packagePath string) error {
	lockPath := s.getLockPath(packagePath)
	if err := s.lockRepo.Save(&pkg.Lock, lockPath); err != nil {
		return fmt.Errorf("failed to save lock file to %s: %w", lockPath, err)
	}
	return nil
}

// loadOrCreateLock loads the lock file or creates a new empty one.
func (s *PackageService) loadOrCreateLock(packagePath string) domain.PackageLock {
	lockPath := s.getLockPath(packagePath)
	lock, err := s.lockRepo.Load(lockPath)
	if err != nil || lock == nil {
		return domain.PackageLock{
			Updated:   "",
			Hash:      "",
			Installed: make(map[string]domain.LockedDependency),
		}
	}
	return *lock
}

// getLockPath returns the lock file path for a given package path.
func (s *PackageService) getLockPath(packagePath string) string {
	dir := filepath.Dir(packagePath)
	return filepath.Join(dir, "boss.lock")
}
