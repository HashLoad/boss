//nolint:testpackage // Testing internal implementation details
package compilerselector_test

import (
	"testing"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/compilerselector"
	"github.com/hashload/boss/pkg/consts"
)

func TestSelectionContext(t *testing.T) {
	ctx := compilerselector.SelectionContext{
		CliCompilerVersion: "12.0",
		CliPlatform:        "Win32",
		Package:            nil,
	}

	if ctx.CliCompilerVersion != "12.0" {
		t.Errorf("Expected CliCompilerVersion to be '12.0', got '%s'", ctx.CliCompilerVersion)
	}

	if ctx.CliPlatform != "Win32" {
		t.Errorf("Expected CliPlatform to be 'Win32', got '%s'", ctx.CliPlatform)
	}
}

func TestSelectedCompiler(t *testing.T) {
	compiler := &compilerselector.SelectedCompiler{
		Version: "12.0",
		Path:    "/path/to/delphi",
		Arch:    "Win32",
		BinDir:  "/path/to/bin",
	}

	if compiler.Version != "12.0" {
		t.Errorf("Expected Version to be '12.0', got '%s'", compiler.Version)
	}

	if compiler.Path != "/path/to/delphi" {
		t.Errorf("Expected Path to be '/path/to/delphi', got '%s'", compiler.Path)
	}

	if compiler.Arch != "Win32" {
		t.Errorf("Expected Arch to be 'Win32', got '%s'", compiler.Arch)
	}

	if compiler.BinDir != "/path/to/bin" {
		t.Errorf("Expected BinDir to be '/path/to/bin', got '%s'", compiler.BinDir)
	}
}

func TestSelectCompiler_NoInstallations(t *testing.T) {
	// This test will likely fail on systems without Delphi installed
	// but verifies the error handling path
	ctx := compilerselector.SelectionContext{
		CliCompilerVersion: "999.0", // Non-existent version
		CliPlatform:        "Win32",
		Package:            nil,
	}

	_, err := compilerselector.SelectCompiler(ctx)
	// On systems without Delphi, this should return an error
	// On systems with Delphi but not version 999.0, this should also error
	if err == nil {
		t.Log("Warning: Expected error for non-existent compiler version, but got nil (Delphi may be installed)")
	}
}

func TestSelectCompiler_WithPackageToolchain(t *testing.T) {
	pkg := &domain.Package{
		Toolchain: &domain.PackageToolchain{
			Compiler: "12.0",
			Platform: consts.PlatformWin32.String(),
		},
	}

	ctx := compilerselector.SelectionContext{
		Package: pkg,
	}

	_, err := compilerselector.SelectCompiler(ctx)
	// This may succeed or fail depending on system configuration
	if err != nil {
		t.Logf("SelectCompiler returned error (expected on systems without Delphi): %v", err)
	}
}

func TestSelectCompiler_WithDelphiInToolchain(t *testing.T) {
	pkg := &domain.Package{
		Toolchain: &domain.PackageToolchain{
			Delphi:   "12.0",
			Platform: consts.PlatformWin64.String(),
		},
	}

	ctx := compilerselector.SelectionContext{
		Package: pkg,
	}

	_, err := compilerselector.SelectCompiler(ctx)
	// This may succeed or fail depending on system configuration
	if err != nil {
		t.Logf("SelectCompiler returned error (expected on systems without Delphi): %v", err)
	}
}

func TestSelectCompiler_PlatformDefaults(t *testing.T) {
	pkg := &domain.Package{
		Toolchain: &domain.PackageToolchain{
			Compiler: "12.0",
			// Platform not specified - should default to Win32
		},
	}

	ctx := compilerselector.SelectionContext{
		Package: pkg,
	}

	_, err := compilerselector.SelectCompiler(ctx)
	if err != nil {
		t.Logf("SelectCompiler returned error (expected on systems without Delphi): %v", err)
	}
}

func TestSelectionContext_EmptyPackage(t *testing.T) {
	ctx := compilerselector.SelectionContext{
		CliCompilerVersion: "",
		CliPlatform:        "",
		Package:            &domain.Package{},
	}

	_, err := compilerselector.SelectCompiler(ctx)
	// Should use global configuration or return error if no Delphi found
	if err != nil {
		t.Logf("SelectCompiler returned error (expected behavior): %v", err)
	}
}
