package models_test

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashload/boss/pkg/models"
)

func TestPackage_AddDependency(t *testing.T) {
	tests := []struct {
		name         string
		initialDeps  map[string]string
		addDep       string
		addVer       string
		expectedDeps map[string]string
	}{
		{
			name:        "add new dependency to empty map",
			initialDeps: map[string]string{},
			addDep:      "github.com/hashload/boss",
			addVer:      "1.0.0",
			expectedDeps: map[string]string{
				"github.com/hashload/boss": "1.0.0",
			},
		},
		{
			name: "add new dependency to existing map",
			initialDeps: map[string]string{
				"github.com/existing/repo": "1.0.0",
			},
			addDep: "github.com/hashload/boss",
			addVer: "2.0.0",
			expectedDeps: map[string]string{
				"github.com/existing/repo": "1.0.0",
				"github.com/hashload/boss": "2.0.0",
			},
		},
		{
			name: "update existing dependency - exact match",
			initialDeps: map[string]string{
				"github.com/hashload/boss": "1.0.0",
			},
			addDep: "github.com/hashload/boss",
			addVer: "2.0.0",
			expectedDeps: map[string]string{
				"github.com/hashload/boss": "2.0.0",
			},
		},
		{
			name: "update existing dependency - case insensitive",
			initialDeps: map[string]string{
				"github.com/HashLoad/Boss": "1.0.0",
			},
			addDep: "github.com/hashload/boss",
			addVer: "2.0.0",
			expectedDeps: map[string]string{
				"github.com/HashLoad/Boss": "2.0.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg := &models.Package{
				Dependencies: tt.initialDeps,
			}

			pkg.AddDependency(tt.addDep, tt.addVer)

			if len(pkg.Dependencies) != len(tt.expectedDeps) {
				t.Errorf("Dependencies count = %d, want %d", len(pkg.Dependencies), len(tt.expectedDeps))
			}

			for key, expectedVer := range tt.expectedDeps {
				if actualVer, exists := pkg.Dependencies[key]; !exists {
					t.Errorf("Dependency %q not found", key)
				} else if actualVer != expectedVer {
					t.Errorf("Dependency %q version = %q, want %q", key, actualVer, expectedVer)
				}
			}
		})
	}
}

func TestPackage_AddProject(t *testing.T) {
	tests := []struct {
		name            string
		initialProjects []string
		addProject      string
		expectedCount   int
	}{
		{
			name:            "add to empty projects",
			initialProjects: []string{},
			addProject:      "project1.dproj",
			expectedCount:   1,
		},
		{
			name:            "add to existing projects",
			initialProjects: []string{"project1.dproj"},
			addProject:      "project2.dproj",
			expectedCount:   2,
		},
		{
			name:            "add nil initial projects",
			initialProjects: nil,
			addProject:      "project1.dproj",
			expectedCount:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg := &models.Package{
				Projects: tt.initialProjects,
			}

			pkg.AddProject(tt.addProject)

			if len(pkg.Projects) != tt.expectedCount {
				t.Errorf("Projects count = %d, want %d", len(pkg.Projects), tt.expectedCount)
			}

			found := false
			for _, p := range pkg.Projects {
				if p == tt.addProject {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Project %q not found in Projects list", tt.addProject)
			}
		})
	}
}

func TestPackage_UninstallDependency(t *testing.T) {
	tests := []struct {
		name          string
		initialDeps   map[string]string
		uninstallDep  string
		expectedCount int
	}{
		{
			name: "uninstall existing dependency",
			initialDeps: map[string]string{
				"github.com/hashload/boss":  "1.0.0",
				"github.com/hashload/horse": "2.0.0",
			},
			uninstallDep:  "github.com/hashload/boss",
			expectedCount: 1,
		},
		{
			name: "uninstall non-existing dependency",
			initialDeps: map[string]string{
				"github.com/hashload/boss": "1.0.0",
			},
			uninstallDep:  "github.com/hashload/notexists",
			expectedCount: 1,
		},
		{
			name: "uninstall case insensitive",
			initialDeps: map[string]string{
				"github.com/HashLoad/Boss": "1.0.0",
			},
			uninstallDep:  "github.com/hashload/boss",
			expectedCount: 0,
		},
		{
			name:          "uninstall from empty map",
			initialDeps:   map[string]string{},
			uninstallDep:  "github.com/hashload/boss",
			expectedCount: 0,
		},
		{
			name:          "uninstall from nil map",
			initialDeps:   nil,
			uninstallDep:  "github.com/hashload/boss",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg := &models.Package{
				Dependencies: tt.initialDeps,
			}

			pkg.UninstallDependency(tt.uninstallDep)

			actualCount := 0
			if pkg.Dependencies != nil {
				actualCount = len(pkg.Dependencies)
			}

			if actualCount != tt.expectedCount {
				t.Errorf("Dependencies count after uninstall = %d, want %d", actualCount, tt.expectedCount)
			}
		})
	}
}

func TestPackage_GetParsedDependencies(t *testing.T) {
	tests := []struct {
		name          string
		pkg           *models.Package
		expectedCount int
	}{
		{
			name:          "nil package",
			pkg:           nil,
			expectedCount: 0,
		},
		{
			name: "empty dependencies",
			pkg: &models.Package{
				Dependencies: map[string]string{},
			},
			expectedCount: 0,
		},
		{
			name: "nil dependencies",
			pkg: &models.Package{
				Dependencies: nil,
			},
			expectedCount: 0,
		},
		{
			name: "with dependencies",
			pkg: &models.Package{
				Dependencies: map[string]string{
					"github.com/hashload/boss":  "1.0.0",
					"github.com/hashload/horse": "^2.0.0",
				},
			},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pkg.GetParsedDependencies()
			if len(result) != tt.expectedCount {
				t.Errorf("GetParsedDependencies() returned %d, want %d", len(result), tt.expectedCount)
			}
		})
	}
}

func TestLoadPackageOther_ValidPackage(t *testing.T) {
	tempDir := t.TempDir()

	pkgContent := map[string]any{
		"name":        "test-package",
		"description": "A test package",
		"version":     "1.0.0",
		"mainsrc":     "./src",
		"dependencies": map[string]string{
			"github.com/hashload/boss": "1.0.0",
		},
	}

	data, err := json.Marshal(pkgContent)
	if err != nil {
		t.Fatalf("Failed to marshal package: %v", err)
	}

	pkgPath := filepath.Join(tempDir, "boss.json")
	err = os.WriteFile(pkgPath, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write package file: %v", err)
	}

	pkg, err := models.LoadPackageOther(pkgPath)
	if err != nil {
		t.Fatalf("LoadPackageOther() error = %v", err)
	}

	if pkg.Name != "test-package" {
		t.Errorf("Name = %q, want %q", pkg.Name, "test-package")
	}
	if pkg.Description != "A test package" {
		t.Errorf("Description = %q, want %q", pkg.Description, "A test package")
	}
	if pkg.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", pkg.Version, "1.0.0")
	}
	if len(pkg.Dependencies) != 1 {
		t.Errorf("Dependencies count = %d, want 1", len(pkg.Dependencies))
	}
}

func TestLoadPackageOther_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()

	pkg, err := models.LoadPackageOther(filepath.Join(tempDir, "nonexistent.json"))
	if err == nil {
		t.Error("LoadPackageOther() should return error for non-existent file")
	}
	if pkg == nil {
		t.Error("LoadPackageOther() should return a new package even on error")
	}
}

func TestLoadPackageOther_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()

	invalidPath := filepath.Join(tempDir, "invalid.json")
	err := os.WriteFile(invalidPath, []byte("not valid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	_, err = models.LoadPackageOther(invalidPath)
	if err == nil {
		t.Error("LoadPackageOther() should return error for invalid JSON")
	}
}

func TestLoadPackageOther_EmptyJSON(t *testing.T) {
	tempDir := t.TempDir()

	emptyPath := filepath.Join(tempDir, "empty.json")
	err := os.WriteFile(emptyPath, []byte("{}"), 0644)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	pkg, err := models.LoadPackageOther(emptyPath)
	if err != nil {
		t.Fatalf("LoadPackageOther() error = %v", err)
	}
	if pkg == nil {
		t.Error("LoadPackageOther() should return a package for empty JSON")
	}
}

// MockFileSystem is a simple mock for testing.
type MockFileSystem struct {
	Files map[string][]byte
}

func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		Files: make(map[string][]byte),
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

func (m *MockFileSystem) MkdirAll(_ string, _ os.FileMode) error {
	return nil
}

func (m *MockFileSystem) Stat(_ string) (os.FileInfo, error) {
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) Remove(name string) error {
	delete(m.Files, name)
	return nil
}

func (m *MockFileSystem) RemoveAll(_ string) error {
	return nil
}

func (m *MockFileSystem) Rename(oldpath, newpath string) error {
	if data, ok := m.Files[oldpath]; ok {
		m.Files[newpath] = data
		delete(m.Files, oldpath)
	}
	return nil
}

// mockReadCloser implements io.ReadCloser for testing.
type mockReadCloser struct {
	data   []byte
	offset int
}

func (r *mockReadCloser) Read(p []byte) (int, error) {
	if r.offset >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.offset:])
	r.offset += n
	return n, nil
}

func (r *mockReadCloser) Close() error {
	return nil
}

// mockWriteCloser implements io.WriteCloser for testing.
type mockWriteCloser struct {
	fs   *MockFileSystem
	name string
	buf  []byte
}

func (w *mockWriteCloser) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)
	return len(p), nil
}

func (w *mockWriteCloser) Close() error {
	w.fs.Files[w.name] = w.buf
	return nil
}

func (m *MockFileSystem) Open(name string) (io.ReadCloser, error) {
	if data, ok := m.Files[name]; ok {
		return &mockReadCloser{data: data}, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileSystem) Create(name string) (io.WriteCloser, error) {
	return &mockWriteCloser{fs: m, name: name}, nil
}

func (m *MockFileSystem) Exists(name string) bool {
	_, ok := m.Files[name]
	return ok
}

func (m *MockFileSystem) IsDir(_ string) bool {
	return false
}

func TestPackage_Save_WithMockFS(t *testing.T) {
	mockFS := NewMockFileSystem()

	pkg := &models.Package{
		Name:         "test-package",
		Version:      "1.0.0",
		Dependencies: map[string]string{},
	}
	pkg.SetFS(mockFS)

	// Create an empty lock to avoid nil pointer
	lock := models.PackageLock{
		Installed: map[string]models.LockedDependency{},
	}
	lock.SetFS(mockFS)
	pkg.Lock = lock

	result := pkg.Save()

	if len(result) == 0 {
		t.Error("Save() should return non-empty bytes")
	}

	// Verify the package was serialized correctly
	var parsed map[string]any
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Errorf("Save() result is not valid JSON: %v", err)
	}

	if parsed["name"] != "test-package" {
		t.Errorf("Saved name = %v, want %q", parsed["name"], "test-package")
	}
}

func TestLoadPackageOtherWithFS_ValidPackage(t *testing.T) {
	mockFS := NewMockFileSystem()

	pkgContent := map[string]any{
		"name":        "mock-package",
		"version":     "2.0.0",
		"description": "A mock package",
	}

	data, _ := json.Marshal(pkgContent)
	mockFS.Files["/test/boss.json"] = data
	mockFS.Files["/test/boss-lock.json"] = []byte("{}")

	pkg, err := models.LoadPackageOtherWithFS("/test/boss.json", mockFS)
	if err != nil {
		t.Fatalf("LoadPackageOtherWithFS() error = %v", err)
	}

	if pkg.Name != "mock-package" {
		t.Errorf("Name = %q, want %q", pkg.Name, "mock-package")
	}
	if pkg.Version != "2.0.0" {
		t.Errorf("Version = %q, want %q", pkg.Version, "2.0.0")
	}
}

func TestLoadPackageOtherWithFS_FileNotFound(t *testing.T) {
	mockFS := NewMockFileSystem()

	pkg, err := models.LoadPackageOtherWithFS("/nonexistent/boss.json", mockFS)
	if err == nil {
		t.Error("LoadPackageOtherWithFS() should return error for non-existent file")
	}
	if pkg == nil {
		t.Error("LoadPackageOtherWithFS() should return a new package even on error")
	}
}

func TestLoadPackageLockWithFS_NewLock(t *testing.T) {
	mockFS := NewMockFileSystem()

	pkg := &models.Package{
		Name: "test-package",
	}
	pkg.SetFS(mockFS)

	// No lock file exists
	lock := models.LoadPackageLockWithFS(pkg, mockFS)

	if lock.Hash == "" {
		t.Error("New PackageLock should have a hash")
	}
	if lock.Installed == nil {
		t.Error("New PackageLock should have non-nil Installed map")
	}
}

func TestLoadPackageLockWithFS_ExistingLock(t *testing.T) {
	mockFS := NewMockFileSystem()

	lockContent := map[string]any{
		"hash":    "abc123",
		"updated": "2025-01-01T00:00:00Z",
		"installedModules": map[string]any{
			"github.com/test/repo": map[string]any{
				"name":    "repo",
				"version": "1.0.0",
				"hash":    "def456",
			},
		},
	}
	data, _ := json.Marshal(lockContent)
	// When Package has fileName "/test/boss.json", lock path becomes "/test/boss-lock.json"
	mockFS.Files["/test/boss-lock.json"] = data

	// Create package content
	pkgContent := map[string]any{"name": "test-package"}
	pkgData, _ := json.Marshal(pkgContent)
	mockFS.Files["/test/boss.json"] = pkgData

	// Load the package first to set fileName properly
	pkg, err := models.LoadPackageOtherWithFS("/test/boss.json", mockFS)
	if err != nil {
		t.Fatalf("LoadPackageOtherWithFS() error = %v", err)
	}

	// Now the lock should be loaded from the file
	lock := models.LoadPackageLockWithFS(pkg, mockFS)

	if lock.Hash != "abc123" {
		t.Errorf("Hash = %q, want %q", lock.Hash, "abc123")
	}
	if len(lock.Installed) != 1 {
		t.Errorf("Installed count = %d, want 1", len(lock.Installed))
	}
}

func TestPackageLock_Save_WithMockFS(_ *testing.T) {
	mockFS := NewMockFileSystem()

	lock := models.PackageLock{
		Hash: "test-hash",
		Installed: map[string]models.LockedDependency{
			"github.com/test/repo": {
				Name:    "repo",
				Version: "1.0.0",
			},
		},
	}
	lock.SetFS(mockFS)

	lock.Save()

	// Since we don't have direct access to fileName, we just verify no panic occurred
	// The Save method should work without error
}

func TestDependency_GetURL_SSH(t *testing.T) {
	dep := models.ParseDependency("github.com/hashload/horse", "^1.0.0")

	// Force SSH URL
	dep.UseSSH = true

	url := dep.GetURL()

	if url == "" {
		t.Error("GetURL() should return non-empty URL")
	}
}

func TestDependency_GetURL_HTTPS(t *testing.T) {
	dep := models.ParseDependency("github.com/hashload/horse", "^1.0.0")

	// Force HTTPS URL
	dep.UseSSH = false

	url := dep.GetURL()

	if url == "" {
		t.Error("GetURL() should return non-empty URL")
	}

	// Should contain https
	if !strings.Contains(url, "https://") {
		t.Errorf("GetURL() = %q, should contain https://", url)
	}
}
