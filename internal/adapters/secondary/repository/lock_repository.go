// Package repository provides implementations for domain repositories.
package repository

import (
	//nolint:gosec // We are not using this for security purposes
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"path/filepath"
	"time"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/ports"
	"github.com/hashload/boss/internal/infra"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/msg"
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
	if err := r.MigrateOldFormat(lockPath, lockPath); err != nil {
		msg.Warn("⚠️ Failed to migrate old lock file: %v", err)
	}

	data, err := r.fs.ReadFile(lockPath)
	if err != nil {
		return r.createEmptyLock(""), err
	}

	lock := &domain.PackageLock{
		Updated:   time.Now().Format(time.RFC3339),
		Installed: make(map[string]domain.LockedDependency),
	}

	if err := json.Unmarshal(data, lock); err != nil {
		return nil, err
	}

	return lock, nil
}

// createEmptyLock creates a new empty lock with a hash based on the package name.
func (r *FileLockRepository) createEmptyLock(packageName string) *domain.PackageLock {
	//nolint:gosec // We are not using this for security purposes
	hash := md5.New()
	if _, err := io.WriteString(hash, packageName); err != nil {
		msg.Warn("⚠️ Failed on write machine id to hash")
	}

	return &domain.PackageLock{
		Updated:   time.Now().Format(time.RFC3339),
		Hash:      hex.EncodeToString(hash.Sum(nil)),
		Installed: map[string]domain.LockedDependency{},
	}
}

// Save persists the lock file to the given path.
func (r *FileLockRepository) Save(lock *domain.PackageLock, lockPath string) error {
	lock.Updated = time.Now().Format(time.RFC3339)

	data, err := json.MarshalIndent(lock, "", "\t")
	if err != nil {
		return err
	}

	return r.fs.WriteFile(lockPath, data, 0600)
}

// MigrateOldFormat migrates from old lock file format if needed.
func (r *FileLockRepository) MigrateOldFormat(_, newPath string) error {
	dir := filepath.Dir(newPath)
	oldFileName := filepath.Join(dir, consts.FilePackageLockOld)
	newFileName := filepath.Join(dir, consts.FilePackageLock)

	if r.fs.Exists(oldFileName) && oldFileName != newFileName {
		return r.fs.Rename(oldFileName, newFileName)
	}

	return nil
}
