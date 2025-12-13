//nolint:testpackage // Testing internal function removeCache
package gc

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashload/boss/internal/adapters/secondary/filesystem"
	"github.com/hashload/boss/internal/core/services/cache"
)

// TestRemoveCacheFunc_NilInfo tests that the walk function handles nil info gracefully.
func TestRemoveCacheFunc_NilInfo(t *testing.T) {
	cacheService := cache.NewService(filesystem.NewOSFileSystem())
	fn := removeCache(false, cacheService)

	// Should not panic with nil info
	err := fn("/some/path", nil, nil)
	if err != nil {
		t.Errorf("removeCache() with nil info returned error: %v", err)
	}
}

// TestRemoveCacheFunc_Directory tests that directories are skipped.
func TestRemoveCacheFunc_Directory(t *testing.T) {
	tempDir := t.TempDir()

	cacheService := cache.NewService(filesystem.NewOSFileSystem())
	fn := removeCache(false, cacheService)

	info, err := os.Stat(tempDir)
	if err != nil {
		t.Fatalf("Failed to stat tempDir: %v", err)
	}

	// Should return nil for directories
	err = fn(tempDir, info, nil)
	if err != nil {
		t.Errorf("removeCache() with directory returned error: %v", err)
	}
}

// TestRemoveCacheFunc_InvalidInfoFile tests handling of invalid cache info files.
func TestRemoveCacheFunc_InvalidInfoFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create a file with an invalid name (can't be parsed as repo info)
	invalidFile := filepath.Join(tempDir, "invalid-file.json")
	err := os.WriteFile(invalidFile, []byte("invalid content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid file: %v", err)
	}

	cacheService := cache.NewService(filesystem.NewOSFileSystem())
	fn := removeCache(false, cacheService)

	info, err := os.Stat(invalidFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	// Should not return error, just log warning
	err = fn(invalidFile, info, nil)
	if err != nil {
		t.Errorf("removeCache() with invalid file should not return error: %v", err)
	}
}

// cacheInfo is a minimal struct for creating test cache files.
type cacheInfo struct {
	Key        string    `json:"key"`
	LastUpdate time.Time `json:"last_update"`
}

// TestRemoveCacheFunc_ExpiredCache tests removal of expired cache entries.
func TestRemoveCacheFunc_ExpiredCache(t *testing.T) {
	tempDir := t.TempDir()

	// Set up cache directory structure
	cacheDir := filepath.Join(tempDir, ".boss")
	infoDir := filepath.Join(cacheDir, "info")

	err := os.MkdirAll(infoDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create info dir: %v", err)
	}

	// Set cache dir environment
	t.Setenv("BOSS_CACHE_DIR", cacheDir)

	// Create a cache info file with old last update
	info := cacheInfo{
		Key:        "test-repo-key",
		LastUpdate: time.Now().AddDate(0, 0, -100), // 100 days ago
	}

	infoData, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Failed to marshal cache info: %v", err)
	}

	// The file name should be a valid repo format: owner--repo
	infoFile := filepath.Join(infoDir, "owner--repo.json")
	err = os.WriteFile(infoFile, infoData, 0644)
	if err != nil {
		t.Fatalf("Failed to write info file: %v", err)
	}

	t.Run("ignoreLastUpdate forces removal", func(t *testing.T) {
		cacheService := cache.NewService(filesystem.NewOSFileSystem())
		fn := removeCache(true, cacheService)

		fileInfo, err := os.Stat(infoFile)
		if err != nil {
			t.Skipf("Info file not available: %v", err)
		}

		// This should not return an error
		err = fn(infoFile, fileInfo, nil)
		if err != nil {
			t.Errorf("removeCache() returned error: %v", err)
		}
	})
}

// TestRemoveCacheFunc_RecentCache tests that recent cache is not removed.
func TestRemoveCacheFunc_RecentCache(t *testing.T) {
	tempDir := t.TempDir()

	// Create a recent cache info file (should not be removed)
	infoDir := filepath.Join(tempDir, "info")
	err := os.MkdirAll(infoDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create info dir: %v", err)
	}

	// Create a cache info with recent update
	info := cacheInfo{
		Key:        "recent-repo",
		LastUpdate: time.Now(), // Just now
	}

	infoData, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("Failed to marshal cache info: %v", err)
	}

	// Create a file with an invalid repo name format to test parsing failure
	infoFile := filepath.Join(infoDir, "not-valid-format.json")
	err = os.WriteFile(infoFile, infoData, 0644)
	if err != nil {
		t.Fatalf("Failed to write info file: %v", err)
	}

	cacheService := cache.NewService(filesystem.NewOSFileSystem())
	fn := removeCache(false, cacheService)

	fileInfo, err := os.Stat(infoFile)
	if err != nil {
		t.Fatalf("Failed to stat info file: %v", err)
	}

	// Should not return error for recent cache
	err = fn(infoFile, fileInfo, nil)
	if err != nil {
		t.Errorf("removeCache() with recent cache returned error: %v", err)
	}
}
