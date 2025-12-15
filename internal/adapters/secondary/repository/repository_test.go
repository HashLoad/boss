//nolint:testpackage // Testing internal implementation details
package repository

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/infra"
)

// MockFileSystem implements infra.FileSystem for testing.
type MockFileSystem struct {
	files   map[string][]byte
	renamed map[string]string
}

func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files:   make(map[string][]byte),
		renamed: make(map[string]string),
	}
}

func (m *MockFileSystem) ReadFile(name string) ([]byte, error) {
	if data, ok := m.files[name]; ok {
		return data, nil
	}
	return nil, errors.New("file not found")
}

func (m *MockFileSystem) WriteFile(name string, data []byte, _ os.FileMode) error {
	m.files[name] = data
	return nil
}

func (m *MockFileSystem) MkdirAll(_ string, _ os.FileMode) error {
	return nil
}

func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	if _, ok := m.files[name]; ok {
		//nolint:nilnil // Mock for testing
		return nil, nil
	}
	return nil, errors.New("file not found")
}

func (m *MockFileSystem) Remove(name string) error {
	delete(m.files, name)
	return nil
}

func (m *MockFileSystem) RemoveAll(_ string) error {
	return nil
}

func (m *MockFileSystem) Rename(oldpath, newpath string) error {
	if data, ok := m.files[oldpath]; ok {
		m.files[newpath] = data
		delete(m.files, oldpath)
		m.renamed[oldpath] = newpath
		return nil
	}
	return errors.New("file not found")
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
	_, ok := m.files[name]
	return ok
}

func (m *MockFileSystem) IsDir(_ string) bool {
	return false
}

func TestFileLockRepository_Load_Success(t *testing.T) {
	fs := NewMockFileSystem()

	lockData := domain.PackageLock{
		Hash:    "testhash",
		Updated: time.Now().Format(time.RFC3339),
		Installed: map[string]domain.LockedDependency{
			"github.com/test/repo": {
				Name:    "repo",
				Version: "1.0.0",
				Hash:    "dephash",
			},
		},
	}

	data, _ := json.Marshal(lockData)
	fs.files["/project/boss-lock.json"] = data

	repo := NewFileLockRepository(fs)

	loaded, err := repo.Load("/project/boss-lock.json")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if loaded.Hash != "testhash" {
		t.Errorf("expected hash 'testhash', got '%s'", loaded.Hash)
	}

	if len(loaded.Installed) != 1 {
		t.Errorf("expected 1 installed dependency, got %d", len(loaded.Installed))
	}
}

func TestFileLockRepository_Load_FileNotFound(t *testing.T) {
	fs := NewMockFileSystem()
	repo := NewFileLockRepository(fs)

	lock, err := repo.Load("/nonexistent/boss-lock.json")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if lock == nil {
		t.Error("expected empty lock to be returned, got nil")
		return
	}
	if lock.Installed == nil {
		t.Error("expected Installed map to be initialized")
	}
}

func TestFileLockRepository_Load_InvalidJSON(t *testing.T) {
	fs := NewMockFileSystem()
	fs.files["/project/boss-lock.json"] = []byte("invalid json{")

	repo := NewFileLockRepository(fs)

	_, err := repo.Load("/project/boss-lock.json")

	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestFileLockRepository_Save_Success(t *testing.T) {
	fs := NewMockFileSystem()
	repo := NewFileLockRepository(fs)

	lock := &domain.PackageLock{
		Hash:      "savehash",
		Updated:   time.Now().Format(time.RFC3339),
		Installed: make(map[string]domain.LockedDependency),
	}

	err := repo.Save(lock, "/project/boss-lock.json")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, ok := fs.files["/project/boss-lock.json"]; !ok {
		t.Error("expected file to be saved")
	}
}

func TestFileLockRepository_MigrateOldFormat_FileExists(t *testing.T) {
	fs := NewMockFileSystem()
	fs.files["/project/boss.lock"] = []byte(`{"hash":"oldhash"}`)

	repo := NewFileLockRepository(fs)

	err := repo.MigrateOldFormat("/project/boss.lock", "/project/boss-lock.json")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, ok := fs.files["/project/boss-lock.json"]; !ok {
		t.Error("expected file to be renamed to new path")
	}

	if _, ok := fs.files["/project/boss.lock"]; ok {
		t.Error("expected old file to be removed")
	}
}

func TestFileLockRepository_MigrateOldFormat_FileDoesNotExist(t *testing.T) {
	fs := NewMockFileSystem()
	repo := NewFileLockRepository(fs)

	err := repo.MigrateOldFormat("/project/boss.lock", "/project/boss-lock.json")

	if err != nil {
		t.Errorf("expected no error when file doesn't exist, got %v", err)
	}
}
