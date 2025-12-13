package compiler_selector

import (
	"errors"
	"path/filepath"
	"strings"

	registryadapter "github.com/hashload/boss/internal/adapters/secondary/registry"
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/env"
)

type SelectionContext struct {
	CliCompilerVersion string
	CliPlatform        string
	Package            *domain.Package
}

type SelectedCompiler struct {
	Version string
	Path    string
	Arch    string
	BinDir  string
}

func SelectCompiler(ctx SelectionContext) (*SelectedCompiler, error) {
	installations := registryadapter.GetDetectedDelphis()
	if len(installations) == 0 {
		return nil, errors.New("no Delphi installation found")
	}

	if ctx.CliCompilerVersion != "" {
		return findCompiler(installations, ctx.CliCompilerVersion, ctx.CliPlatform)
	}

	if ctx.Package != nil && ctx.Package.Toolchain != nil {
		tc := ctx.Package.Toolchain

		platform := tc.Platform
		if platform == "" {
			platform = "Win32"
		}

		if tc.Compiler != "" {
			return findCompiler(installations, tc.Compiler, platform)
		}

		if tc.Delphi != "" {
			return findCompiler(installations, tc.Delphi, platform)
		}
	}

	globalPath := env.GlobalConfiguration().DelphiPath
	if globalPath != "" {

		for _, inst := range installations {
			instDir := filepath.Dir(inst.Path)
			if strings.EqualFold(instDir, globalPath) {
				return createSelectedCompiler(inst), nil
			}
		}

		return &SelectedCompiler{
			Path:   filepath.Join(globalPath, "dcc32.exe"),
			BinDir: globalPath,
			Arch:   "Win32",
		}, nil
	}

	if len(installations) > 0 {
		latest := installations[0]
		for _, inst := range installations[1:] {
			if inst.Version > latest.Version {
				latest = inst
			}
		}
		return createSelectedCompiler(latest), nil
	}

	return nil, errors.New("could not determine compiler")
}

func findCompiler(installations []registryadapter.DelphiInstallation, version string, platform string) (*SelectedCompiler, error) {
	if platform == "" {
		platform = "Win32"
	}

	for _, inst := range installations {
		if inst.Version == version && strings.EqualFold(inst.Arch, platform) {
			return createSelectedCompiler(inst), nil
		}
	}
	return nil, errors.New("compiler version " + version + " for platform " + platform + " not found")
}

func createSelectedCompiler(inst registryadapter.DelphiInstallation) *SelectedCompiler {
	binDir := filepath.Dir(inst.Path)
	exeName := "dcc32.exe"
	if inst.Arch == "Win64" {
		exeName = "dcc64.exe"
	}
	return &SelectedCompiler{
		Version: inst.Version,
		Path:    filepath.Join(binDir, exeName),
		Arch:    inst.Arch,
		BinDir:  binDir,
	}
}
