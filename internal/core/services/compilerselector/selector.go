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

	// Determine target platform
	var targetPlatform string
	if ctx.CliPlatform != "" {
		targetPlatform = ctx.CliPlatform
	} else if ctx.Package != nil && ctx.Package.Toolchain != nil && ctx.Package.Toolchain.Platform != "" {
		targetPlatform = ctx.Package.Toolchain.Platform
	} else {
		targetPlatform = consts.PlatformWin32.String()
	}

	// 1. If CLI compiler version is specified
	if ctx.CliCompilerVersion != "" {
		return findCompiler(installations, ctx.CliCompilerVersion, targetPlatform)
	}

	// 2. If toolchain compiler version is specified
	if ctx.Package != nil && ctx.Package.Toolchain != nil && ctx.Package.Toolchain.Compiler != "" {
		return findCompiler(installations, ctx.Package.Toolchain.Compiler, targetPlatform)
	}

	// 3. Fallback to global path
	globalPath := s.config.GetDelphiPath()
	if globalPath != "" {
		// Try to find a matching installation for the target platform in the global path
		for _, inst := range installations {
			instDir := filepath.Dir(inst.Path)
			if strings.EqualFold(instDir, globalPath) && strings.EqualFold(inst.Arch, targetPlatform) {
				return createSelectedCompiler(inst), nil
			}
		}

		// Fallback if not found in registered installations
		var compilerBinary string
		switch targetPlatform {
		case consts.PlatformWin64.String():
			compilerBinary = "dcc64.exe"
		default:
			compilerBinary = "dcc32.exe"
		}
		return &SelectedCompiler{
			Path:   filepath.Join(globalPath, compilerBinary),
			BinDir: globalPath,
			Arch:   targetPlatform,
		}, nil
	}

	// 4. Fallback to the latest installation that matches the target platform
	var bestMatch *registryadapter.DelphiInstallation
	for _, inst := range installations {
		if strings.EqualFold(inst.Arch, targetPlatform) {
			if bestMatch == nil || inst.Version > bestMatch.Version {
				instCopy := inst
				bestMatch = &instCopy
			}
		}
	}
	if bestMatch != nil {
		return createSelectedCompiler(*bestMatch), nil
	}

	// If no installation matching the target platform is found, fallback to the latest overall installation
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

//nolint:lll // Function signature cannot be easily shortened
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
