//nolint:testpackage // Testing internal Lazarus utility functions
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

const mockLpiContent = `<?xml version="1.0" encoding="UTF-8"?>
<CONFIG>
  <ProjectOptions>
    <Version Value="12"/>
    <General>
      <Title Value="my_lazarus_project"/>
    </General>
    <BuildModes>
      <Item Name="Default" Default="True"/>
      <Item Name="Debug">
        <CompilerOptions>
          <SearchPaths>
            <OtherUnitFiles Value="src"/>
          </SearchPaths>
        </CompilerOptions>
      </Item>
    </BuildModes>
  </ProjectOptions>
  <CompilerOptions>
    <SearchPaths>
      <OtherUnitFiles Value="src"/>
    </SearchPaths>
  </CompilerOptions>
</CONFIG>
`

const mockLpkContent = `<?xml version="1.0" encoding="UTF-8"?>
<CONFIG>
  <Package Name="MyPackage">
    <CompilerOptions>
      <SearchPaths>
        <OtherUnitFiles Value="src"/>
      </SearchPaths>
    </CompilerOptions>
  </Package>
</CONFIG>
`

// TestLazarusPathInjection verifies that other unit search paths are correctly injected in both .lpi and .lpk files.
func TestLazarusPathInjection(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	var err error

	// Initialize package manager
	fs := filesystem.NewOSFileSystem()
	packageRepo := repository.NewFilePackageRepository(fs)
	lockRepo := repository.NewFileLockRepository(fs)
	packageService := packages.NewPackageService(packageRepo, lockRepo)
	pkgmanager.SetInstance(packageService)

	// Set up a mock dependency
	modulesDir := filepath.Join(tempDir, "modules")
	err = os.MkdirAll(filepath.Join(modulesDir, "horse", "src"), 0755)
	if err != nil {
		t.Fatalf("Failed to create modules: %v", err)
	}

	// Create boss.json for dependency
	bossJSONContent := `{"name": "horse", "mainsrc": "src"}`
	err = os.WriteFile(filepath.Join(modulesDir, "horse", "boss.json"), []byte(bossJSONContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write dependency boss.json: %v", err)
	}

	// Create a dummy source file in the dependency so it matches RegexArtifacts
	err = os.WriteFile(filepath.Join(modulesDir, "horse", "src", "dummy.pas"), []byte("unit dummy;"), 0644)
	if err != nil {
		t.Fatalf("Failed to write dummy.pas: %v", err)
	}

	// 1. Test LPI (Project) Path Injection
	lpiPath := filepath.Join(tempDir, "project.lpi")
	err = os.WriteFile(lpiPath, []byte(mockLpiContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write mock LPI: %v", err)
	}

	updateOtherUnitFilesProject(lpiPath)

	// Verify LPI XML contents
	doc := etree.NewDocument()
	if err = doc.ReadFromFile(lpiPath); err != nil {
		t.Fatalf("Failed to read updated LPI: %v", err)
	}

	root := doc.Root()
	compOpts := root.SelectElement("CompilerOptions")
	if compOpts == nil {
		t.Fatal("CompilerOptions element not found in LPI")
	}
	searchPaths := compOpts.SelectElement("SearchPaths")
	if searchPaths == nil {
		t.Fatal("SearchPaths element not found in LPI")
	}
	otherUnitFiles := searchPaths.SelectElement("OtherUnitFiles")
	if otherUnitFiles == nil {
		t.Fatal("OtherUnitFiles element not found in LPI")
	}
	valAttr := otherUnitFiles.SelectAttr("Value")
	if valAttr == nil {
		t.Fatal("Value attribute not found in LPI OtherUnitFiles")
	}

	expectedPath := filepath.Clean("modules/horse/src")
	if !strings.Contains(filepath.Clean(valAttr.Value), expectedPath) {
		t.Errorf("Expected path %q to be injected in LPI, got %q", expectedPath, valAttr.Value)
	}

	// 2. Test LPK (Package) Path Injection
	lpkPath := filepath.Join(tempDir, "package.lpk")
	err = os.WriteFile(lpkPath, []byte(mockLpkContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write mock LPK: %v", err)
	}

	updateOtherUnitFilesProject(lpkPath)

	// Verify LPK XML contents
	docLpk := etree.NewDocument()
	if err = docLpk.ReadFromFile(lpkPath); err != nil {
		t.Fatalf("Failed to read updated LPK: %v", err)
	}

	rootLpk := docLpk.Root()
	pkgOpts := rootLpk.SelectElement("Package")
	if pkgOpts == nil {
		t.Fatal("Package element not found in LPK")
	}
	compOptsLpk := pkgOpts.SelectElement("CompilerOptions")
	if compOptsLpk == nil {
		t.Fatal("CompilerOptions element not found in LPK")
	}
	searchPathsLpk := compOptsLpk.SelectElement("SearchPaths")
	if searchPathsLpk == nil {
		t.Fatal("SearchPaths element not found in LPK")
	}
	otherUnitFilesLpk := searchPathsLpk.SelectElement("OtherUnitFiles")
	if otherUnitFilesLpk == nil {
		t.Fatal("OtherUnitFiles element not found in LPK")
	}
	valAttrLpk := otherUnitFilesLpk.SelectAttr("Value")
	if valAttrLpk == nil {
		t.Fatal("Value attribute not found in LPK OtherUnitFiles")
	}

	if !strings.Contains(filepath.Clean(valAttrLpk.Value), expectedPath) {
		t.Errorf("Expected path %q to be injected in LPK, got %q", expectedPath, valAttrLpk.Value)
	}
}
