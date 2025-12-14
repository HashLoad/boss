package lock

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/hashload/boss/internal/core/domain"
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

func (m *MockFileSystem) ReadFile(name string) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (m *MockFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return nil
}

func (m *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return nil
}

func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	return nil, errors.New("not implemented")
}

func (m *MockFileSystem) Remove(name string) error {
	return nil
}

func (m *MockFileSystem) RemoveAll(path string) error {
	return nil
}

func (m *MockFileSystem) Rename(oldpath, newpath string) error {
	return nil
}

func (m *MockFileSystem) Open(name string) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}

func (m *MockFileSystem) Create(name string) (io.WriteCloser, error) {
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

func (m *MockLockRepository) Load(lockPath string) (*domain.PackageLock, error) {
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	return m.lock, nil
}

func (m *MockLockRepository) Save(lock *domain.PackageLock, lockPath string) error {
	m.lock = lock
	return m.saveErr
}

func (m *MockLockRepository) MigrateOldFormat(oldPath, newPath string) error {
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
