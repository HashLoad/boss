package ports

import "github.com/hashload/boss/internal/core/domain"

// Compiler defines the contract for compiling Delphi projects.
type Compiler interface {
	// Compile compiles a Delphi project file.
	Compile(dprojPath string, dep *domain.Dependency, rootLock domain.PackageLock) bool

	// GetCompilerParameters returns the MSBuild parameters for compilation.
	GetCompilerParameters(rootPath string, dep *domain.Dependency, platform string) string

	// BuildSearchPath builds the search path for a dependency.
	BuildSearchPath(dep *domain.Dependency) string
}

// ArtifactManager defines the contract for managing build artifacts.
type ArtifactManager interface {
	// EnsureArtifacts collects artifacts for a locked dependency.
	EnsureArtifacts(lockedDependency *domain.LockedDependency, dep domain.Dependency, rootPath string)

	// MoveArtifacts moves artifacts to the shared folder.
	MoveArtifacts(dep domain.Dependency, rootPath string)

	// CollectArtifacts collects artifact files from a path.
	CollectArtifacts(artifactList []string, path string) []string
}
