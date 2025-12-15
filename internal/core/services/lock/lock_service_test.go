//nolint:testpackage // Testing internal implementation details
package lock

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/infra"
)

// MockFileSystem implements infra.FileSystem for testing.
type MockFileSystem struct {
	files       map[string]bool
	directories map[string]bool
}

func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files:       make(map[string]bool),
		directories: make(map[string]bool),
	}
}

func (m *MockFileSystem) ReadFile(_ string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (m *MockFileSystem) WriteFile(_ string, _ []byte, _ os.FileMode) error {
	return nil
}

func (m *MockFileSystem) MkdirAll(_ string, _ os.FileMode) error {
	return nil
}

func (m *MockFileSystem) Stat(_ string) (os.FileInfo, error) {
	return nil, errors.New("not implemented")
}

func (m *MockFileSystem) Remove(_ string) error {
	return nil
}

func (m *MockFileSystem) RemoveAll(_ string) error {
	return nil
}

func (m *MockFileSystem) Rename(_, _ string) error {
	return nil
}

func (m *MockFileSystem) ReadDir(_ string) ([]infra.DirEntry, error) {
	return nil, nil
}

func (m *MockFileSystem) Open(_ string) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

func (m *MockFileSystem) Create(_ string) (io.WriteCloser, error) {
	return nil, errors.New("not implemented")
}

func (m *MockFileSystem) Exists(name string) bool {
	return m.files[name] || m.directories[name]
}

func (m *MockFileSystem) IsDir(name string) bool {
	return m.directories[name]
}

func (m *MockFileSystem) AddFile(path string) {
	m.files[path] = true
}

func (m *MockFileSystem) AddDir(path string) {
	m.directories[path] = true
}

// MockLockRepository implements ports.LockRepository for testing.
type MockLockRepository struct {
	lock         *domain.PackageLock
	loadErr      error
	saveErr      error
	migrateCalls int
}

func NewMockLockRepository() *MockLockRepository {
	return &MockLockRepository{}
}

func (m *MockLockRepository) Load(_ string) (*domain.PackageLock, error) {
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	return m.lock, nil
}

func (m *MockLockRepository) Save(lock *domain.PackageLock, _ string) error {
	m.lock = lock
	return m.saveErr
}

func (m *MockLockRepository) MigrateOldFormat(_, _ string) error {
	m.migrateCalls++
	return nil
}

func (m *MockLockRepository) SetLock(lock *domain.PackageLock) {
	m.lock = lock
}

func (m *MockLockRepository) SetLoadError(err error) {
	m.loadErr = err
}

func TestLockService_NeedUpdate_ReturnsTrueWhenNotInstalled(t *testing.T) {
	repo := NewMockLockRepository()
	fs := NewMockFileSystem()
	service := NewLockService(repo, fs)

	lock := &domain.PackageLock{
		Installed: make(map[string]domain.LockedDependency),
	}

	dep := domain.ParseDependency("github.com/test/repo", "1.0.0")

	needUpdate := service.NeedUpdate(lock, dep, "1.0.0", "/modules")

	if !needUpdate {
		t.Error("expected NeedUpdate to return true when dependency is not installed")
	}
}

func TestLockService_NeedUpdate_ReturnsTrueWhenDirNotExists(t *testing.T) {
	repo := NewMockLockRepository()
	fs := NewMockFileSystem()
	service := NewLockService(repo, fs)

	lock := &domain.PackageLock{
		Installed: map[string]domain.LockedDependency{
			"github.com/test/repo": {
				Name:    "repo",
				Version: "1.0.0",
				Hash:    "somehash",
			},
		},
	}

	dep := domain.ParseDependency("github.com/test/repo", "1.0.0")

	needUpdate := service.NeedUpdate(lock, dep, "1.0.0", "/modules")

	if !needUpdate {
		t.Error("expected NeedUpdate to return true when dependency dir doesn't exist")
	}
}

func TestLockService_AddDependency_CreatesNewEntry(t *testing.T) {
	repo := NewMockLockRepository()
	fs := NewMockFileSystem()
	service := NewLockService(repo, fs)

	lock := &domain.PackageLock{
		Installed: make(map[string]domain.LockedDependency),
	}

	dep := domain.ParseDependency("github.com/test/repo", "1.0.0")

	service.AddDependency(lock, dep, "1.0.0", "/modules")

	if _, ok := lock.Installed["github.com/test/repo"]; !ok {
		t.Error("expected dependency to be added to lock")
	}
}

func TestLockService_AddDependency_UpdatesExistingEntry(t *testing.T) {
	repo := NewMockLockRepository()
	fs := NewMockFileSystem()
	service := NewLockService(repo, fs)

	lock := &domain.PackageLock{
		Installed: map[string]domain.LockedDependency{
			"github.com/test/repo": {
				Name:    "repo",
				Version: "1.0.0",
				Hash:    "oldhash",
			},
		},
	}

	dep := domain.ParseDependency("github.com/test/repo", "2.0.0")

	service.AddDependency(lock, dep, "2.0.0", "/modules")

	installed := lock.Installed["github.com/test/repo"]
	if installed.Version != "2.0.0" {
		t.Errorf("expected version 2.0.0, got %s", installed.Version)
	}
}

func TestLockService_Save(t *testing.T) {
	repo := NewMockLockRepository()
	fs := NewMockFileSystem()

	service := NewLockService(repo, fs)

	lock := &domain.PackageLock{
		Hash:      "testhash",
		Installed: make(map[string]domain.LockedDependency),
	}

	err := service.Save(lock, "/project")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repo.lock != lock {
		t.Error("expected lock to be saved in repository")
	}
}
