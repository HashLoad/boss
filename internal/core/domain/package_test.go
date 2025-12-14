package domain_test

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/infra"
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
			pkg := &domain.Package{
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
			pkg := &domain.Package{
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
			pkg := &domain.Package{
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
		pkg           *domain.Package
		expectedCount int
	}{
		{
			name:          "nil package",
			pkg:           nil,
			expectedCount: 0,
		},
		{
			name: "empty dependencies",
			pkg: &domain.Package{
				Dependencies: map[string]string{},
			},
			expectedCount: 0,
		},
		{
			name: "nil dependencies",
			pkg: &domain.Package{
				Dependencies: nil,
			},
			expectedCount: 0,
		},
		{
			name: "with dependencies",
			pkg: &domain.Package{
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

func (m *MockFileSystem) ReadDir(_ string) ([]infra.DirEntry, error) {
	return nil, nil
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

func TestDependency_GetURL_SSH(t *testing.T) {
	dep := domain.ParseDependency("github.com/hashload/horse", "^1.0.0")

	// Force SSH URL
	dep.UseSSH = true

	url := dep.GetURL()

	if url == "" {
		t.Error("GetURL() should return non-empty URL")
	}
}

func TestDependency_GetURL_HTTPS(t *testing.T) {
	dep := domain.ParseDependency("github.com/hashload/horse", "^1.0.0")

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
