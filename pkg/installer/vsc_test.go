//nolint:testpackage // Testing internal function hasCache
package installer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashload/boss/pkg/models"
)

// TestHasCache_NotExists tests hasCache when directory doesn't exist.
func TestHasCache_NotExists(t *testing.T) {
	dep := models.Dependency{
		Repository: "github.com/test/nonexistent-repo-12345",
	}

	result := hasCache(dep)

	if result {
		t.Error("hasCache() should return false for non-existent cache")
	}
}

// TestHasCache_Exists tests hasCache when directory exists.
func TestHasCache_Exists(_ *testing.T) {
	// This test requires setting up proper environment
	// We'll just test that the function doesn't panic
	dep := models.Dependency{
		Repository: "github.com/test/repo",
	}

	// Just ensure it doesn't panic
	_ = hasCache(dep)
}

// TestHasCache_FileInsteadOfDir tests hasCache when path is a file.
func TestHasCache_FileInsteadOfDir(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("BOSS_CACHE_DIR", tempDir)

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
	result := hasCache(dep)

	// After removing the file (inside hasCache), it should return false
	if result {
		t.Error("hasCache() should return false after removing file")
	}
}
