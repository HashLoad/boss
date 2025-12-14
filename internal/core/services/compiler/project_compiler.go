package compiler

import (
	"github.com/hashload/boss/internal/core/domain"
)

// DefaultProjectCompiler implements ProjectCompiler.
type DefaultProjectCompiler struct{}

// Compile compiles a dproj file.
func (d *DefaultProjectCompiler) Compile(dprojPath string, dep *domain.Dependency, rootLock domain.PackageLock) bool {
	return compile(dprojPath, dep, rootLock, nil, nil)
}
