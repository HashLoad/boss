//nolint:testpackage // Testing internal functions
package compiler

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/infra"
	"github.com/hashload/boss/pkg/consts"
)

func TestGetCompilerParameters(t *testing.T) {
	tests := []struct {
		name     string
		rootPath string
		dep      *domain.Dependency
		platform string
		wantBpl  bool
		wantDcp  bool
		wantDcu  bool
	}{
		{
			name:     "with dependency",
			rootPath: "/test/modules",
			dep:      &domain.Dependency{Repository: "github.com/test/lib"},
			platform: consts.PlatformWin32.String(),
			wantBpl:  true,
			wantDcp:  true,
			wantDcu:  true,
		},
		{
			name:     "without dependency",
			rootPath: "/test/modules",
			dep:      nil,
			platform: consts.PlatformWin64.String(),
			wantBpl:  true,
			wantDcp:  true,
			wantDcu:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCompilerParameters(tt.rootPath, tt.dep, tt.platform)

			if tt.wantBpl && !containsStr(result, "DCC_BplOutput") {
				t.Error("Expected DCC_BplOutput in parameters")
			}
			if tt.wantDcp && !containsStr(result, "DCC_DcpOutput") {
				t.Error("Expected DCC_DcpOutput in parameters")
			}
			if tt.wantDcu && !containsStr(result, "DCC_DcuOutput") {
				t.Error("Expected DCC_DcuOutput in parameters")
			}
			if !containsStr(result, tt.platform) {
				t.Errorf("Expected platform %s in parameters", tt.platform)
			}
		})
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestBuildSearchPath(t *testing.T) {
	t.Skip("Skipping test that requires pkgmanager initialization - needs integration test setup")

	tests := []struct {
		name string
		dep  *domain.Dependency
	}{
		{
			name: "nil dependency",
			dep:  nil,
		},
		{
			name: "with dependency",
			dep:  &domain.Dependency{Repository: "github.com/test/lib"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildSearchPath(tt.dep)

			if tt.dep == nil && result != "" {
				t.Error("Expected empty string for nil dependency")
			}
			if tt.dep != nil && result == "" {
				t.Error("Expected non-empty string for valid dependency")
			}
		})
	}
}

func TestMoveArtifacts(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()

	dep := domain.Dependency{Repository: "github.com/test/lib"}
	modulePath := filepath.Join(tmpDir, dep.Name())

	// Create source directories with test files (using actual consts)
	bplDir := filepath.Join(modulePath, ".bpl")
	if err := os.MkdirAll(bplDir, 0755); err != nil {
		t.Fatal(err)
	}

	testFile := filepath.Join(bplDir, "test.bpl")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		t.Fatal(err)
	}

	// Create destination directory
	destBplDir := filepath.Join(tmpDir, ".bpl")
	if err := os.MkdirAll(destBplDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Test move using the artifact manager
	fs := &OSFileSystemWrapper{}
	artifactMgr := NewDefaultArtifactManager(fs)
	artifactMgr.MoveArtifacts(dep, tmpDir)

	// Verify file was moved
	destFile := filepath.Join(destBplDir, "test.bpl")
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("Expected file to be moved to destination")
	}
}

// OSFileSystemWrapper wraps os package functions for testing.
type OSFileSystemWrapper struct{}

func (o *OSFileSystemWrapper) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (o *OSFileSystemWrapper) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

func (o *OSFileSystemWrapper) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (o *OSFileSystemWrapper) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (o *OSFileSystemWrapper) Remove(name string) error {
	return os.Remove(name)
}

func (o *OSFileSystemWrapper) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (o *OSFileSystemWrapper) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func (o *OSFileSystemWrapper) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

func (o *OSFileSystemWrapper) Create(name string) (io.WriteCloser, error) {
	return os.Create(name)
}

func (o *OSFileSystemWrapper) IsDir(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (o *OSFileSystemWrapper) ReadDir(name string) ([]infra.DirEntry, error) {
	entries, err := os.ReadDir(name)
	if err != nil {
		return nil, err
	}
	result := make([]infra.DirEntry, len(entries))
	for i, e := range entries {
		result[i] = &dirEntryWrapper{entry: e}
	}
	return result, nil
}

func (o *OSFileSystemWrapper) Exists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

type dirEntryWrapper struct {
	entry os.DirEntry
}

func (d *dirEntryWrapper) Name() string               { return d.entry.Name() }
func (d *dirEntryWrapper) IsDir() bool                { return d.entry.IsDir() }
func (d *dirEntryWrapper) Type() os.FileMode          { return d.entry.Type() }
func (d *dirEntryWrapper) Info() (os.FileInfo, error) { return d.entry.Info() }

func TestEnsureArtifacts(t *testing.T) {
	tmpDir := t.TempDir()

	dep := domain.Dependency{Repository: "github.com/test/lib"}
	modulePath := filepath.Join(tmpDir, dep.Name())

	// Create directories with test files
	bplDir := filepath.Join(modulePath, "bpl")
	os.MkdirAll(bplDir, 0755)
	os.WriteFile(filepath.Join(bplDir, "test.bpl"), []byte("test"), 0600)

	lockedDep := &domain.LockedDependency{
		Artifacts: domain.DependencyArtifacts{},
	}

	fs := &OSFileSystemWrapper{}
	artifactMgr := NewDefaultArtifactManager(fs)
	artifactMgr.EnsureArtifacts(lockedDep, dep, tmpDir)

	// The function should have collected artifacts
}
