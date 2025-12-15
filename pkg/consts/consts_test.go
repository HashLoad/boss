package consts_test

import (
	"path/filepath"
	"testing"

	"github.com/hashload/boss/pkg/consts"
)

func TestConstants_FileNames(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"FilePackage", consts.FilePackage, "boss.json"},
		{"FilePackageLock", consts.FilePackageLock, "boss-lock.json"},
		{"FileBplOrder", consts.FileBplOrder, "bpl_order.txt"},
		{"FilePackageLockOld", consts.FilePackageLockOld, "boss.lock"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestConstants_FileExtensions(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"FileExtensionBpl", consts.FileExtensionBpl, ".bpl"},
		{"FileExtensionDcp", consts.FileExtensionDcp, ".dcp"},
		{"FileExtensionDpk", consts.FileExtensionDpk, ".dpk"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestConstants_Folders(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"FolderDependencies", consts.FolderDependencies, "modules"},
		{"FolderEnv", consts.FolderEnv, "env"},
		{"FolderBossHome", consts.FolderBossHome, ".boss"},
		{"BinFolder", consts.BinFolder, ".bin"},
		{"BplFolder", consts.BplFolder, ".bpl"},
		{"DcpFolder", consts.DcpFolder, ".dcp"},
		{"DcuFolder", consts.DcuFolder, ".dcu"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestConstants_EnvFolders(t *testing.T) {
	sep := string(filepath.Separator)

	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"FolderEnvBpl", consts.FolderEnvBpl, "env" + sep + "bpl"},
		{"FolderEnvDcp", consts.FolderEnvDcp, "env" + sep + "dcp"},
		{"FolderEnvDcu", consts.FolderEnvDcu, "env" + sep + "dcu"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestConstants_Config(t *testing.T) {
	if consts.BossConfigFile != "boss.cfg.json" {
		t.Errorf("BossConfigFile = %q, want %q", consts.BossConfigFile, "boss.cfg.json")
	}

	if consts.MinimalDependencyVersion != ">0.0.0" {
		t.Errorf("MinimalDependencyVersion = %q, want %q", consts.MinimalDependencyVersion, ">0.0.0")
	}
}

func TestConstants_XMLTags(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"XMLTagNameProperty", consts.XMLTagNameProperty, "PropertyGroup"},
		{"XMLTagNameLibraryPath", consts.XMLTagNameLibraryPath, "DCC_UnitSearchPath"},
		{"XMLTagNameCompilerOptions", consts.XMLTagNameCompilerOptions, "CompilerOptions"},
		{"XMLTagNameSearchPaths", consts.XMLTagNameSearchPaths, "SearchPaths"},
		{"XMLTagNameOtherUnitFiles", consts.XMLTagNameOtherUnitFiles, "OtherUnitFiles"},
		{"XMLTagNameProjectOptions", consts.XMLTagNameProjectOptions, "ProjectOptions"},
		{"XMLTagNameBuildModes", consts.XMLTagNameBuildModes, "BuildModes"},
		{"XMLTagNameItem", consts.XMLTagNameItem, "Item"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestConstants_BossInternal(t *testing.T) {
	if consts.BossInternalDir != "internal." {
		t.Errorf("BossInternalDir = %q, want %q", consts.BossInternalDir, "internal.")
	}

	if consts.BossInternalDirOld != "{internal}" {
		t.Errorf("BossInternalDirOld = %q, want %q", consts.BossInternalDirOld, "{internal}")
	}
}

func TestConstants_RegexArtifacts(t *testing.T) {
	expected := "(.*.inc$|.*.pas$|.*.dfm$|.*.fmx$|.*.dcu$|.*.bpl$|.*.dcp$|.*.res$)"
	if consts.RegexArtifacts != expected {
		t.Errorf("RegexArtifacts = %q, want %q", consts.RegexArtifacts, expected)
	}
}

func TestDefaultPaths(t *testing.T) {
	paths := consts.DefaultPaths()

	if len(paths) != 4 {
		t.Errorf("DefaultPaths() returned %d items, want 4", len(paths))
	}

	expectedPaths := map[string]bool{
		".bpl": false,
		".dcu": false,
		".dcp": false,
		".bin": false,
	}

	for _, path := range paths {
		if _, exists := expectedPaths[path]; exists {
			expectedPaths[path] = true
		} else {
			t.Errorf("Unexpected path in DefaultPaths(): %q", path)
		}
	}

	for path, found := range expectedPaths {
		if !found {
			t.Errorf("Expected path %q not found in DefaultPaths()", path)
		}
	}
}
