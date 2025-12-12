//nolint:testpackage // Testing internal function hasCache
package installer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashload/boss/pkg/models"
)

// TestDependencyManager_HasCache_NotExists tests hasCache when directory doesn't exist.
func TestDependencyManager_HasCache_NotExists(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("BOSS_CACHE_DIR", tempDir)

	dm := NewDefaultDependencyManager()
	dm.cacheDir = tempDir

	dep := models.Dependency{
		Repository: "github.com/test/nonexistent-repo-12345",
	}

	result := dm.hasCache(dep)

	if result {
		t.Error("hasCache() should return false for non-existent cache")
	}
}

// TestDependencyManager_HasCache_Exists tests hasCache when directory exists.
func TestDependencyManager_HasCache_Exists(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("BOSS_CACHE_DIR", tempDir)

	dm := NewDefaultDependencyManager()
	dm.cacheDir = tempDir

	dep := models.Dependency{
		Repository: "github.com/test/repo",
	}

	// Create the cache directory
	cacheDir := filepath.Join(tempDir, dep.HashName())
	err := os.MkdirAll(cacheDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	result := dm.hasCache(dep)

	if !result {
		t.Error("hasCache() should return true when cache directory exists")
	}
}

// TestDependencyManager_HasCache_FileInsteadOfDir tests hasCache when path is a file.
func TestDependencyManager_HasCache_FileInsteadOfDir(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("BOSS_CACHE_DIR", tempDir)

	dm := NewDefaultDependencyManager()
	dm.cacheDir = tempDir

	// Create a file where directory is expected
	dep := models.Dependency{
		Repository: "github.com/test/filerepo",
	}

	filePath := filepath.Join(tempDir, dep.HashName())
	err := os.WriteFile(filePath, []byte("not a directory"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// hasCache should handle this case
	result := dm.hasCache(dep)

	// After removing the file (inside hasCache), it should return false
	if result {
		t.Error("hasCache() should return false after removing file")
	}
}

// TestResetDependencyCache tests the global reset function.
func TestResetDependencyCache(t *testing.T) {
	// Get the default manager and add some entries
	dm := getDefaultDependencyManager()

	// Mark something as updated
	dm.Cache().MarkUpdated("test-dep")

	// Reset
	ResetDependencyCache()

	// Should be empty now
	if dm.Cache().IsUpdated("test-dep") {
		t.Error("Cache should be empty after ResetDependencyCache()")
	}
}
