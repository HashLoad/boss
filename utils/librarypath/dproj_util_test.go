//nolint:testpackage // Testing internal dproj utility functions
package librarypath

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/beevik/etree"
	"github.com/hashload/boss/internal/adapters/secondary/filesystem"
	"github.com/hashload/boss/internal/adapters/secondary/repository"
	"github.com/hashload/boss/internal/core/services/packages"
	"github.com/hashload/boss/pkg/pkgmanager"
)

const mockDprojContent = `<?xml version="1.0" encoding="utf-8"?>
<Project xmlns="http://schemas.microsoft.com/developer/msbuild/2003">
  <PropertyGroup Condition="'$(Base)'!=''">
    <DCC_UnitSearchPath>$(DCC_UnitSearchPath)</DCC_UnitSearchPath>
  </PropertyGroup>
</Project>
`

// TestUpdateLibraryPathProject_SubdirectoryProject verifies that a .dproj located in a
// subdirectory of the boss.json root gets paths relative to its OWN directory, not to
// the root. dprojName is always an absolute, OS-native path (built via filepath.Join),
// so resolving its parent directory must use filepath.Dir rather than path.Dir -- the
// latter only understands "/" and silently collapses to "." on a Windows backslash path.
func TestUpdateLibraryPathProject_SubdirectoryProject(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	fs := filesystem.NewOSFileSystem()
	packageRepo := repository.NewFilePackageRepository(fs)
	lockRepo := repository.NewFileLockRepository(fs)
	packageService := packages.NewPackageService(packageRepo, lockRepo)
	pkgmanager.SetInstance(packageService)

	depSrcDir := filepath.Join(tempDir, "modules", "mydep", "src")
	if err := os.MkdirAll(depSrcDir, 0755); err != nil {
		t.Fatalf("Failed to create dependency src dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(depSrcDir, "dummy.pas"), []byte("unit dummy;"), 0600); err != nil {
		t.Fatalf("Failed to write dummy.pas: %v", err)
	}
	depBossJSON := `{"name": "mydep", "mainsrc": "src"}`
	depBossJSONPath := filepath.Join(tempDir, "modules", "mydep", "boss.json")
	if err := os.WriteFile(depBossJSONPath, []byte(depBossJSON), 0600); err != nil {
		t.Fatalf("Failed to write dependency boss.json: %v", err)
	}

	projectDir := filepath.Join(tempDir, "app")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}
	dprojPath := filepath.Join(projectDir, "project.dproj")
	if err := os.WriteFile(dprojPath, []byte(mockDprojContent), 0600); err != nil {
		t.Fatalf("Failed to write mock dproj: %v", err)
	}

	updateLibraryPathProject(dprojPath)

	doc := etree.NewDocument()
	if err := doc.ReadFromFile(dprojPath); err != nil {
		t.Fatalf("Failed to read updated dproj: %v", err)
	}

	var searchPath string
	for _, group := range doc.Root().FindElements("PropertyGroup") {
		if el := group.SelectElement("DCC_UnitSearchPath"); el != nil {
			searchPath = el.Text()
		}
	}
	if searchPath == "" {
		t.Fatal("DCC_UnitSearchPath not found in updated dproj")
	}

	expected := filepath.Join("..", "modules", "mydep", "src")
	if !strings.Contains(filepath.Clean(searchPath), expected) {
		t.Errorf("expected DCC_UnitSearchPath to contain %q (relative to the project's own directory), got %q",
			expected, searchPath)
	}

	wrong := filepath.Join("modules", "mydep", "src")
	for _, entry := range strings.Split(searchPath, ";") {
		if filepath.Clean(entry) == wrong {
			t.Errorf("DCC_UnitSearchPath contains %q, which is relative to the boss.json root instead of "+
				"the project's own directory -- rootPath was computed wrong", entry)
		}
	}
}
