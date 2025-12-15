// Package pkgmanager provides convenient access to package operations.
// This package acts as a facade to avoid circular dependencies and provide
// easy access to package service from anywhere in the codebase.
package pkgmanager

import (
	"sync"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/packages"
)

var (
	//nolint:gochecknoglobals // Singleton pattern for package manager
	instance   *packages.PackageService
	instanceMu sync.RWMutex //nolint:gochecknoglobals // Singleton mutex
)

// SetInstance sets the global package service instance.
// This should be called during application initialization (in setup package).
func SetInstance(packageService *packages.PackageService) {
	instanceMu.Lock()
	defer instanceMu.Unlock()
	instance = packageService
}

// GetInstance returns the global package service instance.
func GetInstance() *packages.PackageService {
	instanceMu.RLock()
	defer instanceMu.RUnlock()
	return instance
}

// LoadPackage loads the current project's package file.
// This is a convenience function that uses the global service instance.
func LoadPackage() (*domain.Package, error) {
	return GetInstance().LoadCurrent()
}

// LoadPackageOther loads a package from a specific path.
// This is a convenience function that uses the global service instance.
func LoadPackageOther(path string) (*domain.Package, error) {
	return GetInstance().Load(path)
}

// SavePackage saves a package to a specific path.
func SavePackage(pkg *domain.Package, path string) error {
	return GetInstance().Save(pkg, path)
}

// SavePackageCurrent saves the current project's package file.
func SavePackageCurrent(pkg *domain.Package) error {
	return GetInstance().SaveCurrent(pkg)
}
