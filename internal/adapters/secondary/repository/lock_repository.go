// Package repository provides implementations for domain repositories.
package repository

import (
	"encoding/json"
	"time"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/ports"
	"github.com/hashload/boss/internal/infra"
)

// Compile-time check that FileLockRepository implements ports.LockRepository.
var _ ports.LockRepository = (*FileLockRepository)(nil)

// FileLockRepository implements LockRepository using the filesystem.
type FileLockRepository struct {
	fs infra.FileSystem
}

// NewFileLockRepository creates a new FileLockRepository.
func NewFileLockRepository(fs infra.FileSystem) *FileLockRepository {
	return &FileLockRepository{fs: fs}
}

// Load loads a lock file from the given path.
func (r *FileLockRepository) Load(lockPath string) (*domain.PackageLock, error) {
	data, err := r.fs.ReadFile(lockPath)
	if err != nil {
		return nil, err
	}

	lock := &domain.PackageLock{
		Updated:   time.Now(),
		Installed: make(map[string]domain.LockedDependency),
	}

	if err := json.Unmarshal(data, lock); err != nil {
		return nil, err
	}

	return lock, nil
}

// Save persists the lock file to the given path.
func (r *FileLockRepository) Save(lock *domain.PackageLock, lockPath string) error {
	data, err := json.MarshalIndent(lock, "", "\t")
	if err != nil {
		return err
	}

	return r.fs.WriteFile(lockPath, data, 0600)
}

// MigrateOldFormat migrates from old lock file format if needed.
func (r *FileLockRepository) MigrateOldFormat(oldPath, newPath string) error {
	if r.fs.Exists(oldPath) {
		return r.fs.Rename(oldPath, newPath)
	}
	return nil
}
