package compilerselector_test

import (
	"path/filepath"
	"testing"

	registryadapter "github.com/hashload/boss/internal/adapters/secondary/registry"
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/compilerselector"
	"github.com/hashload/boss/pkg/env"
)

// mockRegistryAdapter is a mock for compilerselector.RegistryAdapter
type mockRegistryAdapter struct {
	installations []registryadapter.DelphiInstallation
}

func (m *mockRegistryAdapter) GetDetectedDelphis() []registryadapter.DelphiInstallation {
	return m.installations
}

// mockConfigProvider embeds env.ConfigProvider to mock only GetDelphiPath
type mockConfigProvider struct {
	env.ConfigProvider
	delphiPath string
}

func (m *mockConfigProvider) GetDelphiPath() string {
	return m.delphiPath
}

func TestSelectCompiler_CliOverrides(t *testing.T) {
	registry := &mockRegistryAdapter{
		installations: []registryadapter.DelphiInstallation{
			{Version: "35.0", Path: filepath.Join("C:", "Delphi35", "bin", "dcc32.exe"), Arch: "Win32"},
			{Version: "35.0", Path: filepath.Join("C:", "Delphi35", "bin", "dcc64.exe"), Arch: "Win64"},
			{Version: "36.0", Path: filepath.Join("C:", "Delphi36", "bin", "dcc32.exe"), Arch: "Win32"},
			{Version: "36.0", Path: filepath.Join("C:", "Delphi36", "bin", "dcc64.exe"), Arch: "Win64"},
		},
	}
	config := &mockConfigProvider{}
	service := compilerselector.NewService(registry, config)

	// CLI version and platform specified
	ctx := compilerselector.SelectionContext{
		CliCompilerVersion: "35.0",
		CliPlatform:        "Win64",
		Package: &domain.Package{
			Toolchain: &domain.PackageToolchain{
				Compiler: "36.0",
				Platform: "Win32",
			},
		},
	}

	selected, err := service.SelectCompiler(ctx)
	if err != nil {
		t.Fatalf("Failed to select compiler: %v", err)
	}

	if selected.Version != "35.0" {
		t.Errorf("Expected version 35.0, got %s", selected.Version)
	}
	if selected.Arch != "Win64" {
		t.Errorf("Expected arch Win64, got %s", selected.Arch)
	}
}

func TestSelectCompiler_ToolchainCompilerAndPlatform(t *testing.T) {
	registry := &mockRegistryAdapter{
		installations: []registryadapter.DelphiInstallation{
			{Version: "35.0", Path: filepath.Join("C:", "Delphi35", "bin", "dcc32.exe"), Arch: "Win32"},
			{Version: "35.0", Path: filepath.Join("C:", "Delphi35", "bin", "dcc64.exe"), Arch: "Win64"},
			{Version: "36.0", Path: filepath.Join("C:", "Delphi36", "bin", "dcc32.exe"), Arch: "Win32"},
			{Version: "36.0", Path: filepath.Join("C:", "Delphi36", "bin", "dcc64.exe"), Arch: "Win64"},
		},
	}
	config := &mockConfigProvider{}
	service := compilerselector.NewService(registry, config)

	// Toolchain compiler and platform specified
	ctx := compilerselector.SelectionContext{
		Package: &domain.Package{
			Toolchain: &domain.PackageToolchain{
				Compiler: "36.0",
				Platform: "Win64",
			},
		},
	}

	selected, err := service.SelectCompiler(ctx)
	if err != nil {
		t.Fatalf("Failed to select compiler: %v", err)
	}

	if selected.Version != "36.0" {
		t.Errorf("Expected version 36.0, got %s", selected.Version)
	}
	if selected.Arch != "Win64" {
		t.Errorf("Expected arch Win64, got %s", selected.Arch)
	}
}

func TestSelectCompiler_ToolchainPlatformOnly(t *testing.T) {
	registry := &mockRegistryAdapter{
		installations: []registryadapter.DelphiInstallation{
			{Version: "35.0", Path: filepath.Join("C:", "Delphi35", "bin", "dcc32.exe"), Arch: "Win32"},
			{Version: "35.0", Path: filepath.Join("C:", "Delphi35", "bin", "dcc64.exe"), Arch: "Win64"},
			{Version: "36.0", Path: filepath.Join("C:", "Delphi36", "bin", "dcc32.exe"), Arch: "Win32"},
			{Version: "36.0", Path: filepath.Join("C:", "Delphi36", "bin", "dcc64.exe"), Arch: "Win64"},
		},
	}
	config := &mockConfigProvider{}
	service := compilerselector.NewService(registry, config)

	// Only platform Win64 specified in toolchain
	ctx := compilerselector.SelectionContext{
		Package: &domain.Package{
			Toolchain: &domain.PackageToolchain{
				Platform: "Win64",
			},
		},
	}

	selected, err := service.SelectCompiler(ctx)
	if err != nil {
		t.Fatalf("Failed to select compiler: %v", err)
	}

	// Should select the latest version (36.0) matching platform Win64
	if selected.Version != "36.0" {
		t.Errorf("Expected version 36.0, got %s", selected.Version)
	}
	if selected.Arch != "Win64" {
		t.Errorf("Expected arch Win64, got %s", selected.Arch)
	}
}

func TestSelectCompiler_FallbackToGlobalPath(t *testing.T) {
	registry := &mockRegistryAdapter{
		installations: []registryadapter.DelphiInstallation{
			{Version: "35.0", Path: filepath.Join("C:", "Delphi35", "bin", "dcc32.exe"), Arch: "Win32"},
		},
	}
	globalPath := filepath.Join("C:", "CustomDelphi", "bin")
	config := &mockConfigProvider{
		delphiPath: globalPath,
	}
	service := compilerselector.NewService(registry, config)

	// Win64 requested, should fallback to dcc64.exe in globalPath
	ctx := compilerselector.SelectionContext{
		Package: &domain.Package{
			Toolchain: &domain.PackageToolchain{
				Platform: "Win64",
			},
		},
	}

	selected, err := service.SelectCompiler(ctx)
	if err != nil {
		t.Fatalf("Failed to select compiler: %v", err)
	}

	expectedPath := filepath.Join(globalPath, "dcc64.exe")
	if selected.Path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, selected.Path)
	}
	if selected.Arch != "Win64" {
		t.Errorf("Expected arch Win64, got %s", selected.Arch)
	}
}
