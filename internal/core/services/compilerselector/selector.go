// Package compilerselector provides functionality for selecting the appropriate Delphi compiler
// based on project configuration, CLI arguments, or system defaults.
package compilerselector

import (
	"errors"
	"path/filepath"
	"strings"

	registryadapter "github.com/hashload/boss/internal/adapters/secondary/registry"
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
)

// SelectionContext holds the context for compiler selection.
type SelectionContext struct {
	CliCompilerVersion string
	CliPlatform        string
	Package            *domain.Package
}

// SelectedCompiler represents the selected compiler configuration.
type SelectedCompiler struct {
	Version string
	Path    string
	Arch    string
	BinDir  string
}

// Service provides compiler selection functionality.
type Service struct {
	registry RegistryAdapter
	config   env.ConfigProvider
}

// RegistryAdapter defines the interface for registry operations needed by the service.
type RegistryAdapter interface {
	GetDetectedDelphis() []registryadapter.DelphiInstallation
}

// DefaultRegistryAdapter wraps the registry adapter.
type DefaultRegistryAdapter struct{}

// GetDetectedDelphis returns detected Delphi installations.
func (d *DefaultRegistryAdapter) GetDetectedDelphis() []registryadapter.DelphiInstallation {
	return registryadapter.GetDetectedDelphis()
}

// NewService creates a new compiler selector service.
func NewService(registry RegistryAdapter, config env.ConfigProvider) *Service {
	return &Service{
		registry: registry,
		config:   config,
	}
}

// NewDefaultService creates a service with default dependencies.
func NewDefaultService() *Service {
	return NewService(&DefaultRegistryAdapter{}, env.GlobalConfiguration())
}

// SelectCompiler selects the appropriate compiler based on the context.
func (s *Service) SelectCompiler(ctx SelectionContext) (*SelectedCompiler, error) {
	installations := s.registry.GetDetectedDelphis()
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
			platform = consts.PlatformWin32.String()
		}

		if tc.Compiler != "" {
			return findCompiler(installations, tc.Compiler, platform)
		}

		if tc.Delphi != "" {
			return findCompiler(installations, tc.Delphi, platform)
		}
	}

	globalPath := s.config.GetDelphiPath()
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
			Arch:   consts.PlatformWin32.String(),
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
	return &SelectedCompiler{
		Version: inst.Version,
		Path:    inst.Path,
		Arch:    inst.Arch,
		BinDir:  filepath.Dir(inst.Path),
	}
}

// SelectCompiler is a convenience function that uses the default service.
// For better testability, inject Service directly in new code.
func SelectCompiler(ctx SelectionContext) (*SelectedCompiler, error) {
	return NewDefaultService().SelectCompiler(ctx)
}
