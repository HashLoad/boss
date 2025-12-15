//nolint:testpackage // Testing internal functions
package librarypath

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCleanPath tests path cleaning functionality.
func TestCleanPath(t *testing.T) {
	tests := []struct {
		name     string
		paths    []string
		fullPath bool
		wantLen  int
	}{
		{
			name:     "empty paths",
			paths:    []string{},
			fullPath: true,
			wantLen:  0,
		},
		{
			name:     "paths without modules prefix",
			paths:    []string{"/usr/lib", "/home/user/lib"},
			fullPath: true,
			wantLen:  2,
		},
		{
			name:     "duplicate paths removed",
			paths:    []string{"/usr/lib", "/usr/lib"},
			fullPath: true,
			wantLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanPath(tt.paths, tt.fullPath)

			if len(result) != tt.wantLen {
				t.Errorf("cleanPath() returned %d paths, want %d", len(result), tt.wantLen)
			}
		})
	}
}

// TestGetNewBrowsingPaths tests browsing paths retrieval.
func TestGetNewBrowsingPaths(t *testing.T) {
	tempDir := t.TempDir()

	// Set up environment
	t.Setenv("BOSS_BASE_DIR", tempDir)

	// Create modules directory
	modulesDir := filepath.Join(tempDir, "modules")
	err := os.MkdirAll(modulesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create modules dir: %v", err)
	}

	paths := []string{"/existing/path"}

	result := GetNewBrowsingPaths(paths, true, tempDir, false)

	// Should at least contain the existing path
	if len(result) == 0 {
		t.Error("GetNewBrowsingPaths() should return paths")
	}
}
