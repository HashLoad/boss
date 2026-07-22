//nolint:testpackage // Testing internal command implementation
package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashload/boss/internal/adapters/secondary/filesystem"
	"github.com/hashload/boss/internal/adapters/secondary/repository"
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/packages"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/pkgmanager"
	"github.com/spf13/cobra"
)

// TestNewCommandRegistration tests that the new command registers correctly.
func TestNewCommandRegistration(t *testing.T) {
	root := &cobra.Command{Use: "boss"}
	newCmdRegister(root)

	var newCmd *cobra.Command
	for _, cmd := range root.Commands() {
		if cmd.Use == "new [project_name]" {
			newCmd = cmd
			break
		}
	}

	if newCmd == nil {
		t.Fatal("New command not found")
	}

	if newCmd.Short == "" {
		t.Error("New command should have a short description")
	}

	typeFlag := newCmd.Flags().Lookup("type")
	if typeFlag == nil {
		t.Fatal("New command should have --type flag")
	}
	if typeFlag.DefValue != "app" {
		t.Errorf("New command --type flag default value should be 'app', got %s", typeFlag.DefValue)
	}

	quietFlag := newCmd.Flags().Lookup("quiet")
	if quietFlag == nil {
		t.Error("New command should have --quiet flag")
	}
}

// TestDoCreateProject_App tests bootstrapping an application project.
func TestDoCreateProject_App(t *testing.T) {
	tempDir := t.TempDir()

	// Redirect the working directory; t.Chdir restores it when the test ends.
	t.Chdir(tempDir)

	// Initialize package manager
	fs := filesystem.NewOSFileSystem()
	packageRepo := repository.NewFilePackageRepository(fs)
	lockRepo := repository.NewFileLockRepository(fs)
	packageService := packages.NewPackageService(packageRepo, lockRepo)
	pkgmanager.SetInstance(packageService)

	projectName := "testapp"
	doCreateProject(projectName, "app", "delphi", true)

	projectPath := filepath.Join(tempDir, projectName)
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Fatalf("Project directory was not created")
	}

	// Check folders
	if _, err := os.Stat(filepath.Join(projectPath, "src")); os.IsNotExist(err) {
		t.Error("src directory was not created")
	}
	if _, err := os.Stat(filepath.Join(projectPath, "tests")); os.IsNotExist(err) {
		t.Error("tests directory was not created")
	}

	// Check boss.json
	bossJSONPath := filepath.Join(projectPath, consts.FilePackage)
	if _, err := os.Stat(bossJSONPath); os.IsNotExist(err) {
		t.Fatal("boss.json was not created")
	}

	bossBytes, readErr := os.ReadFile(bossJSONPath)
	if readErr != nil {
		t.Fatalf("Failed to read boss.json: %v", readErr)
	}

	var pkg domain.Package
	if err := json.Unmarshal(bossBytes, &pkg); err != nil {
		t.Fatalf("Failed to parse boss.json: %v", err)
	}

	if pkg.Name != projectName {
		t.Errorf("Expected package name %q, got %q", projectName, pkg.Name)
	}
	if pkg.Version != "1.0.0" {
		t.Errorf("Expected package version '1.0.0', got %q", pkg.Version)
	}
	if pkg.MainSrc != "src" {
		t.Errorf("Expected mainsrc 'src', got %q", pkg.MainSrc)
	}

	// Check .dpr file
	dprPath := filepath.Join(projectPath, projectName+".dpr")
	if _, err := os.Stat(dprPath); os.IsNotExist(err) {
		t.Fatal(".dpr file was not created")
	}

	dprBytes, readErr := os.ReadFile(dprPath)
	if readErr != nil {
		t.Fatalf("Failed to read .dpr file: %v", readErr)
	}
	dprContent := string(dprBytes)
	if !strings.Contains(dprContent, "program "+projectName) {
		t.Errorf("Expected .dpr file to contain program declaration")
	}

	// Check .dproj file
	dprojPath := filepath.Join(projectPath, projectName+".dproj")
	if _, err := os.Stat(dprojPath); os.IsNotExist(err) {
		t.Fatal(".dproj file was not created")
	}

	dprojBytes, readErr := os.ReadFile(dprojPath)
	if readErr != nil {
		t.Fatalf("Failed to read .dproj file: %v", readErr)
	}
	dprojContent := string(dprojBytes)
	if !strings.Contains(dprojContent, "<ProjectGuid>") {
		t.Error("Expected .dproj to contain ProjectGuid")
	}
	if !strings.Contains(dprojContent, "<AppType>Console</AppType>") {
		t.Error("Expected .dproj to have AppType Console")
	}
}

// TestDoCreateProject_Pkg tests bootstrapping a package project.
func TestDoCreateProject_Pkg(t *testing.T) {
	tempDir := t.TempDir()

	// Redirect the working directory; t.Chdir restores it when the test ends.
	t.Chdir(tempDir)

	// Initialize package manager
	fs := filesystem.NewOSFileSystem()
	packageRepo := repository.NewFilePackageRepository(fs)
	lockRepo := repository.NewFileLockRepository(fs)
	packageService := packages.NewPackageService(packageRepo, lockRepo)
	pkgmanager.SetInstance(packageService)

	projectName := "testpkg"
	doCreateProject(projectName, "pkg", "delphi", true)

	projectPath := filepath.Join(tempDir, projectName)
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Fatalf("Project directory was not created")
	}

	// Check .dpk file
	dpkPath := filepath.Join(projectPath, projectName+".dpk")
	if _, err := os.Stat(dpkPath); os.IsNotExist(err) {
		t.Fatal(".dpk file was not created")
	}

	dpkBytes, readErr := os.ReadFile(dpkPath)
	if readErr != nil {
		t.Fatalf("Failed to read .dpk file: %v", readErr)
	}
	dpkContent := string(dpkBytes)
	if !strings.Contains(dpkContent, "package "+projectName) {
		t.Errorf("Expected .dpk file to contain package declaration")
	}

	// Check .dproj file
	dprojPath := filepath.Join(projectPath, projectName+".dproj")
	if _, err := os.Stat(dprojPath); os.IsNotExist(err) {
		t.Fatal(".dproj file was not created")
	}

	dprojBytes, readErr := os.ReadFile(dprojPath)
	if readErr != nil {
		t.Fatalf("Failed to read .dproj file: %v", readErr)
	}
	dprojContent := string(dprojBytes)
	if !strings.Contains(dprojContent, "<AppType>Package</AppType>") {
		t.Error("Expected .dproj to have AppType Package")
	}
}

// TestDoCreateProject_LazarusApp tests bootstrapping a Lazarus application project.
func TestDoCreateProject_LazarusApp(t *testing.T) {
	tempDir := t.TempDir()

	// Redirect the working directory; t.Chdir restores it when the test ends.
	t.Chdir(tempDir)

	// Initialize package manager
	fs := filesystem.NewOSFileSystem()
	packageRepo := repository.NewFilePackageRepository(fs)
	lockRepo := repository.NewFileLockRepository(fs)
	packageService := packages.NewPackageService(packageRepo, lockRepo)
	pkgmanager.SetInstance(packageService)

	projectName := "testlazapp"
	doCreateProject(projectName, "app", "lazarus", true)

	projectPath := filepath.Join(tempDir, projectName)
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Fatalf("Project directory was not created")
	}

	// Check folders
	if _, err := os.Stat(filepath.Join(projectPath, "src")); os.IsNotExist(err) {
		t.Error("src directory was not created")
	}
	if _, err := os.Stat(filepath.Join(projectPath, "tests")); os.IsNotExist(err) {
		t.Error("tests directory was not created")
	}

	// Check boss.json
	bossJSONPath := filepath.Join(projectPath, consts.FilePackage)
	if _, err := os.Stat(bossJSONPath); os.IsNotExist(err) {
		t.Fatal("boss.json was not created")
	}

	bossBytes, readErr := os.ReadFile(bossJSONPath)
	if readErr != nil {
		t.Fatalf("Failed to read boss.json: %v", readErr)
	}

	var pkg domain.Package
	if err := json.Unmarshal(bossBytes, &pkg); err != nil {
		t.Fatalf("Failed to parse boss.json: %v", err)
	}

	if pkg.Name != projectName {
		t.Errorf("Expected package name %q, got %q", projectName, pkg.Name)
	}

	// Check .lpr file
	lprPath := filepath.Join(projectPath, projectName+".lpr")
	if _, err := os.Stat(lprPath); os.IsNotExist(err) {
		t.Fatal(".lpr file was not created")
	}

	lprBytes, readErr := os.ReadFile(lprPath)
	if readErr != nil {
		t.Fatalf("Failed to read .lpr file: %v", readErr)
	}
	lprContent := string(lprBytes)
	if !strings.Contains(lprContent, "program "+projectName) {
		t.Errorf("Expected .lpr file to contain program declaration")
	}

	// Check .lpi file
	lpiPath := filepath.Join(projectPath, projectName+".lpi")
	if _, err := os.Stat(lpiPath); os.IsNotExist(err) {
		t.Fatal(".lpi file was not created")
	}

	lpiBytes, readErr := os.ReadFile(lpiPath)
	if readErr != nil {
		t.Fatalf("Failed to read .lpi file: %v", readErr)
	}
	lpiContent := string(lpiBytes)
	if !strings.Contains(lpiContent, "<ProjectOptions>") {
		t.Error("Expected .lpi to contain ProjectOptions tag")
	}
	if !strings.Contains(lpiContent, "<Filename Value=\""+projectName+".lpr\"/>") {
		t.Error("Expected .lpi to reference the .lpr main unit")
	}
}

// TestDoCreateProject_LazarusPkg tests bootstrapping a Lazarus package project.
func TestDoCreateProject_LazarusPkg(t *testing.T) {
	tempDir := t.TempDir()

	// Redirect the working directory; t.Chdir restores it when the test ends.
	t.Chdir(tempDir)

	// Initialize package manager
	fs := filesystem.NewOSFileSystem()
	packageRepo := repository.NewFilePackageRepository(fs)
	lockRepo := repository.NewFileLockRepository(fs)
	packageService := packages.NewPackageService(packageRepo, lockRepo)
	pkgmanager.SetInstance(packageService)

	projectName := "testlazpkg"
	doCreateProject(projectName, "pkg", "lazarus", true)

	projectPath := filepath.Join(tempDir, projectName)
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		t.Fatalf("Project directory was not created")
	}

	// Check .lpk file
	lpkPath := filepath.Join(projectPath, projectName+".lpk")
	if _, err := os.Stat(lpkPath); os.IsNotExist(err) {
		t.Fatal(".lpk file was not created")
	}

	lpkBytes, readErr := os.ReadFile(lpkPath)
	if readErr != nil {
		t.Fatalf("Failed to read .lpk file: %v", readErr)
	}
	lpkContent := string(lpkBytes)
	if !strings.Contains(lpkContent, "<Package Version=\"5\">") {
		t.Error("Expected .lpk to contain Package Version 5 tag")
	}
	if !strings.Contains(lpkContent, "<Name Value=\""+projectName+"\"/>") {
		t.Error("Expected .lpk to have the package name")
	}
}
