//nolint:testpackage // Testing internal functions
package dcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashload/boss/pkg/models"
)

// TestGetDcpString tests DCP string generation.
func TestGetDcpString(t *testing.T) {
	tests := []struct {
		name     string
		dcps     []string
		contains string
	}{
		{
			name:     "single dcp",
			dcps:     []string{"/path/to/package.dcp"},
			contains: "package.dcp",
		},
		{
			name:     "multiple dcps",
			dcps:     []string{"/path/to/pkg1.dcp", "/path/to/pkg2.dcp"},
			contains: "pkg1.dcp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDcpString(tt.dcps)

			if !strings.Contains(result, tt.contains) {
				t.Errorf("getDcpString() should contain %q, got %q", tt.contains, result)
			}

			if !strings.Contains(result, CommentBoss) {
				t.Errorf("getDcpString() should contain BOSS comment marker, got %q", result)
			}
		})
	}
}

// TestGetDprDpkFromDproj_NotExists tests when dpk file doesn't exist.
func TestGetDprDpkFromDproj_NotExists(t *testing.T) {
	tempDir := t.TempDir()

	// Create a dproj without corresponding dpk
	dprojPath := filepath.Join(tempDir, "MyProject.dproj")
	err := os.WriteFile(dprojPath, []byte("<Project></Project>"), 0644)
	if err != nil {
		t.Fatalf("Failed to create dproj: %v", err)
	}

	// Change to temp dir for test
	t.Chdir(tempDir)

	result, exists := getDprDpkFromDproj("MyProject.dproj")

	if exists {
		t.Error("getDprDpkFromDproj() should return false when dpk doesn't exist")
	}

	if result != "" {
		t.Errorf("getDprDpkFromDproj() should return empty string when not exists, got %q", result)
	}
}

// TestGetDprDpkFromDproj_Exists tests when dpk file exists.
func TestGetDprDpkFromDproj_Exists(t *testing.T) {
	tempDir := t.TempDir()

	// Create both dproj and dpk
	dprojPath := filepath.Join(tempDir, "MyPackage.dproj")
	dpkPath := filepath.Join(tempDir, "MyPackage.dpk")

	err := os.WriteFile(dprojPath, []byte("<Project></Project>"), 0644)
	if err != nil {
		t.Fatalf("Failed to create dproj: %v", err)
	}

	err = os.WriteFile(dpkPath, []byte("package MyPackage;"), 0644)
	if err != nil {
		t.Fatalf("Failed to create dpk: %v", err)
	}

	// Change to temp dir for test
	t.Chdir(tempDir)

	result, exists := getDprDpkFromDproj("MyPackage.dproj")

	if !exists {
		t.Error("getDprDpkFromDproj() should return true when dpk exists")
	}

	if !strings.HasSuffix(result, ".dpk") {
		t.Errorf("getDprDpkFromDproj() should return dpk path, got %q", result)
	}
}

// TestInjectDcps_NoRequiresSection tests injection when no requires section exists.
func TestInjectDcps_NoRequiresSection(t *testing.T) {
	content := `package MyPackage;
contains
  Unit1 in 'Unit1.pas';
end.`

	dcps := []string{"rtl", "vcl"}

	result, changed := injectDcps(content, dcps)

	if changed {
		t.Error("injectDcps() should return false when no requires section exists")
	}

	if result != content {
		t.Error("injectDcps() should return original content when no requires section")
	}
}

// TestInjectDcps_WithRequiresSection tests injection with existing requires.
func TestInjectDcps_WithRequiresSection(t *testing.T) {
	content := `package MyPackage;
requires
  rtl,
  vcl;
contains
  Unit1 in 'Unit1.pas';
end.`

	dcps := []string{"newpkg"}

	result, changed := injectDcps(content, dcps)

	if !changed {
		t.Error("injectDcps() should return true when requires section is modified")
	}

	if !strings.Contains(result, "newpkg") {
		t.Error("injectDcps() should add new dcp to result")
	}

	if !strings.Contains(result, CommentBoss) {
		t.Error("injectDcps() should add BOSS comment marker")
	}
}

// TestProcessFile_EmptyDcps tests that empty dcps returns unchanged content.
func TestProcessFile_EmptyDcps(t *testing.T) {
	content := "package test;"
	dcps := []string{}

	result, changed := processFile(content, dcps)

	if changed {
		t.Error("processFile() should return false for empty dcps")
	}

	if result != content {
		t.Error("processFile() should return original content for empty dcps")
	}
}

// TestGetRequiresList_NilPackage tests handling of nil package.
func TestGetRequiresList_NilPackage(t *testing.T) {
	result := getRequiresList(nil, models.PackageLock{})

	if len(result) != 0 {
		t.Errorf("getRequiresList() should return empty list for nil package, got %v", result)
	}
}

// TestGetRequiresList_NoDependencies tests package with no dependencies.
func TestGetRequiresList_NoDependencies(t *testing.T) {
	pkg := &models.Package{
		Dependencies: map[string]string{},
	}

	result := getRequiresList(pkg, models.PackageLock{})

	if len(result) != 0 {
		t.Errorf("getRequiresList() should return empty list for no deps, got %v", result)
	}
}
