package filesystem_test

import (
	"os"
	"path/filepath"
	"testing"

	fs "github.com/hashload/boss/internal/adapters/secondary/filesystem"
)

func TestOSFileSystem_ReadWriteFile(t *testing.T) {
	osfs := fs.NewOSFileSystem()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test.txt")
	content := []byte("hello world")

	// Write file
	err := osfs.WriteFile(filePath, content, 0644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Read file
	read, err := osfs.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if string(read) != string(content) {
		t.Errorf("ReadFile() = %q, want %q", string(read), string(content))
	}
}

func TestOSFileSystem_MkdirAll(t *testing.T) {
	osfs := fs.NewOSFileSystem()
	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "a", "b", "c")

	err := osfs.MkdirAll(nestedDir, 0755)
	if err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	if !osfs.IsDir(nestedDir) {
		t.Error("MkdirAll() did not create directory")
	}
}

func TestOSFileSystem_Stat(t *testing.T) {
	osfs := fs.NewOSFileSystem()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "stat_test.txt")

	err := osfs.WriteFile(filePath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	info, err := osfs.Stat(filePath)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}

	if info.Name() != "stat_test.txt" {
		t.Errorf("Stat().Name() = %q, want %q", info.Name(), "stat_test.txt")
	}
}

func TestOSFileSystem_Remove(t *testing.T) {
	osfs := fs.NewOSFileSystem()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "remove_test.txt")

	err := osfs.WriteFile(filePath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err = osfs.Remove(filePath)
	if err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	if osfs.Exists(filePath) {
		t.Error("Remove() did not delete file")
	}
}

func TestOSFileSystem_RemoveAll(t *testing.T) {
	osfs := fs.NewOSFileSystem()
	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "removeall", "nested")

	err := osfs.MkdirAll(nestedDir, 0755)
	if err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	err = osfs.WriteFile(filepath.Join(nestedDir, "file.txt"), []byte("test"), 0644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err = osfs.RemoveAll(filepath.Join(tempDir, "removeall"))
	if err != nil {
		t.Fatalf("RemoveAll() error = %v", err)
	}

	if osfs.Exists(filepath.Join(tempDir, "removeall")) {
		t.Error("RemoveAll() did not delete directory tree")
	}
}

func TestOSFileSystem_Rename(t *testing.T) {
	osfs := fs.NewOSFileSystem()
	tempDir := t.TempDir()
	oldPath := filepath.Join(tempDir, "old.txt")
	newPath := filepath.Join(tempDir, "new.txt")

	err := osfs.WriteFile(oldPath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err = osfs.Rename(oldPath, newPath)
	if err != nil {
		t.Fatalf("Rename() error = %v", err)
	}

	if osfs.Exists(oldPath) {
		t.Error("Rename() did not remove old file")
	}

	if !osfs.Exists(newPath) {
		t.Error("Rename() did not create new file")
	}
}

func TestOSFileSystem_OpenCreate(t *testing.T) {
	osfs := fs.NewOSFileSystem()
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "open_create.txt")

	// Create
	writer, err := osfs.Create(filePath)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	_, err = writer.Write([]byte("created content"))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	writer.Close()

	// Open
	reader, err := osfs.Open(filePath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer reader.Close()

	buf := make([]byte, 100)
	n, err := reader.Read(buf)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if string(buf[:n]) != "created content" {
		t.Errorf("Read() = %q, want %q", string(buf[:n]), "created content")
	}
}

func TestOSFileSystem_Exists(t *testing.T) {
	osfs := fs.NewOSFileSystem()
	tempDir := t.TempDir()

	existingFile := filepath.Join(tempDir, "exists.txt")
	err := osfs.WriteFile(existingFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if !osfs.Exists(existingFile) {
		t.Error("Exists() = false for existing file")
	}

	if osfs.Exists(filepath.Join(tempDir, "nonexistent.txt")) {
		t.Error("Exists() = true for non-existent file")
	}
}

func TestOSFileSystem_IsDir(t *testing.T) {
	osfs := fs.NewOSFileSystem()
	tempDir := t.TempDir()

	// Create a file
	filePath := filepath.Join(tempDir, "file.txt")
	err := osfs.WriteFile(filePath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Create a directory
	dirPath := filepath.Join(tempDir, "subdir")
	err = osfs.MkdirAll(dirPath, 0755)
	if err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	if osfs.IsDir(filePath) {
		t.Error("IsDir() = true for file")
	}

	if !osfs.IsDir(dirPath) {
		t.Error("IsDir() = false for directory")
	}

	if osfs.IsDir(filepath.Join(tempDir, "nonexistent")) {
		t.Error("IsDir() = true for non-existent path")
	}
}

func TestDefaultFileSystem(t *testing.T) {
	if fs.Default == nil {
		t.Error("Default filesystem should not be nil")
	}

	// Test that Default works
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "default_test.txt")

	err := fs.Default.WriteFile(filePath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Default.WriteFile() error = %v", err)
	}

	content, err := fs.Default.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Default.ReadFile() error = %v", err)
	}

	if string(content) != "test" {
		t.Errorf("Default.ReadFile() = %q, want %q", string(content), "test")
	}
}

// MockFileSystem is a mock implementation for testing.
type MockFileSystem struct {
	Files map[string][]byte
	Dirs  map[string]bool
}

func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Files: make(map[string][]byte),
		Dirs:  make(map[string]bool),
	}
}

func (m *MockFileSystem) ReadFile(name string) ([]byte, error) {
	if data, ok := m.Files[name]; ok {
		return data, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) WriteFile(name string, data []byte, _ os.FileMode) error {
	m.Files[name] = data
	return nil
}

func (m *MockFileSystem) MkdirAll(path string, _ os.FileMode) error {
	m.Dirs[path] = true
	return nil
}

func (m *MockFileSystem) Stat(_ string) (os.FileInfo, error) {
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) Remove(name string) error {
	delete(m.Files, name)
	delete(m.Dirs, name)
	return nil
}

func (m *MockFileSystem) RemoveAll(path string) error {
	for k := range m.Files {
		if len(k) >= len(path) && k[:len(path)] == path {
			delete(m.Files, k)
		}
	}
	for k := range m.Dirs {
		if len(k) >= len(path) && k[:len(path)] == path {
			delete(m.Dirs, k)
		}
	}
	return nil
}

func (m *MockFileSystem) Rename(oldpath, newpath string) error {
	if data, ok := m.Files[oldpath]; ok {
		m.Files[newpath] = data
		delete(m.Files, oldpath)
	}
	return nil
}

func (m *MockFileSystem) Open(_ string) (interface {
	Read([]byte) (int, error)
	Close() error
}, error) {
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) Create(_ string) (interface {
	Write([]byte) (int, error)
	Close() error
}, error) {
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) Exists(name string) bool {
	_, fileExists := m.Files[name]
	_, dirExists := m.Dirs[name]
	return fileExists || dirExists
}

func (m *MockFileSystem) IsDir(name string) bool {
	return m.Dirs[name]
}

func TestMockFileSystem(t *testing.T) {
	mockFS := NewMockFileSystem()

	// Test WriteFile and ReadFile
	err := mockFS.WriteFile("/test/file.txt", []byte("mock content"), 0644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	content, err := mockFS.ReadFile("/test/file.txt")
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if string(content) != "mock content" {
		t.Errorf("ReadFile() = %q, want %q", string(content), "mock content")
	}

	// Test Exists
	if !mockFS.Exists("/test/file.txt") {
		t.Error("Exists() should return true for written file")
	}

	// Test Remove
	err = mockFS.Remove("/test/file.txt")
	if err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	if mockFS.Exists("/test/file.txt") {
		t.Error("Exists() should return false after Remove")
	}
}
