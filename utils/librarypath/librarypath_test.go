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

func TestGetNewPaths_MultipleMainSrc(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("BOSS_BASE_DIR", tempDir)
	t.Chdir(tempDir)

	var err error

	// Create modules directory
	modulesDir := filepath.Join(tempDir, "modules")
	libDir := filepath.Join(modulesDir, "my_lib")
	err = os.MkdirAll(filepath.Join(libDir, "src"), 0755)
	if err != nil {
		t.Fatalf("Failed to create src dir: %v", err)
	}
	err = os.MkdirAll(filepath.Join(libDir, "lib"), 0755)
	if err != nil {
		t.Fatalf("Failed to create lib dir: %v", err)
	}

	// Create a dummy source file inside both src and lib so they are walked and recognized
	err = os.WriteFile(filepath.Join(libDir, "src", "my_lib.pas"), []byte("unit my_lib;"), 0600)
	if err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}
	err = os.WriteFile(filepath.Join(libDir, "lib", "helper.pas"), []byte("unit helper;"), 0600)
	if err != nil {
		t.Fatalf("Failed to create dummy file: %v", err)
	}

	// Create a boss.json with multiple mainsrc paths
	bossJSONContent := `{
		"name": "my_lib",
		"mainsrc": "src;lib"
	}`
	err = os.WriteFile(filepath.Join(libDir, "boss.json"), []byte(bossJSONContent), 0600)
	if err != nil {
		t.Fatalf("Failed to write boss.json: %v", err)
	}

	paths := []string{}
	result := GetNewPaths(paths, true, tempDir)

	// Should contain both src and lib paths
	foundSrc := false
	foundLib := false
	for _, p := range result {
		if filepath.Base(p) == "src" {
			foundSrc = true
		}
		if filepath.Base(p) == "lib" {
			foundLib = true
		}
	}

	if !foundSrc {
		t.Error("Expected to find 'src' path in results")
	}
	if !foundLib {
		t.Error("Expected to find 'lib' path in results")
	}
}
