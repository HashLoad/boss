package cache

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/consts"
)

// MockFileSystem implements infra.FileSystem for testing.
type MockFileSystem struct {
	files map[string][]byte
	dirs  map[string]bool
}

func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
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

func (m *MockFileSystem) MkdirAll(path string, _ os.FileMode) error {
	m.dirs[path] = true
	return nil
}

func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	if _, ok := m.files[name]; ok {
		return nil, nil
	}
	if _, ok := m.dirs[name]; ok {
		return nil, nil
	}
	return nil, errors.New("not found")
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

func (m *MockFileSystem) IsDir(name string) bool {
	_, ok := m.dirs[name]
	return ok
}

func TestService_SaveAndLoadRepositoryDetails(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("BOSS_HOME", tempDir)

	// Create the boss home folder structure
	fs := NewMockFileSystem()
	service := NewCacheService(fs)

	dep := domain.ParseDependency("github.com/hashload/horse", "^1.0.0")
	versions := []string{"1.0.0", "1.1.0", "1.2.0"}

	// Save repository details
	err := service.SaveRepositoryDetails(dep, versions)
	if err != nil {
		t.Fatalf("SaveRepositoryDetails() error = %v", err)
	}

	// Verify a file was written
	if len(fs.files) == 0 {
		t.Error("SaveRepositoryDetails() should write a file")
	}

	// Load the data back
	hashName := dep.HashName()
	info, err := service.LoadRepositoryData(hashName)
	if err != nil {
		t.Fatalf("LoadRepositoryData() error = %v", err)
	}

	if info.Name != "horse" {
		t.Errorf("LoadRepositoryData().Name = %q, want %q", info.Name, "horse")
	}

	if len(info.Versions) != 3 {
		t.Errorf("LoadRepositoryData().Versions count = %d, want 3", len(info.Versions))
	}
}

func TestService_LoadRepositoryData_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("BOSS_HOME", tempDir)

	fs := NewMockFileSystem()
	service := NewCacheService(fs)

	_, err := service.LoadRepositoryData("nonexistent")
	if err == nil {
		t.Error("LoadRepositoryData() should return error for non-existent key")
	}
}

// Ensure consts is used (to avoid unused import error)
var _ = consts.FolderBossHome
