package ports

import "github.com/hashload/boss/internal/core/domain"

// LockRepository defines the contract for lock file persistence.
// This interface is implemented by adapters in the infrastructure layer.
type LockRepository interface {
	// Load loads a lock file from the given path.
	// Returns an empty lock if the file doesn't exist.
	Load(lockPath string) (*domain.PackageLock, error)

	// Save persists the lock file to the given path.
	Save(lock *domain.PackageLock, lockPath string) error

	// MigrateOldFormat migrates from old lock file format if needed.
	MigrateOldFormat(oldPath, newPath string) error
}

// PackageRepository defines the contract for package file persistence.
type PackageRepository interface {
	// Load loads a package from the given path.
	Load(packagePath string) (*domain.Package, error)

	// Save persists the package to the given path.
	Save(pkg *domain.Package, packagePath string) error

	// Exists checks if a package file exists at the given path.
	Exists(packagePath string) bool
}
