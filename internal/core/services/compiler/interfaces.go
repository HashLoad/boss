package compiler

import (
	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/infra"
)

// PackageLoader abstracts loading package information.
type PackageLoader interface {
	LoadPackage(path string) (*domain.Package, error)
}

// LockManager abstracts lock file operations.
type LockManager interface {
	Save() error
	GetInstalled(dep domain.Dependency) domain.LockedDependency
	SetInstalled(dep domain.Dependency, locked domain.LockedDependency)
}

// GraphBuilder abstracts dependency graph construction.
type GraphBuilder interface {
	LoadOrderGraph(pkg *domain.Package) *domain.NodeQueue
	LoadOrderGraphAll(pkg *domain.Package) *domain.NodeQueue
}

// ProjectCompiler abstracts project compilation.
type ProjectCompiler interface {
	Compile(dprojPath string, dep *domain.Dependency, rootLock domain.PackageLock) bool
}

// ArtifactManager abstracts artifact operations.
type ArtifactManager interface {
	EnsureArtifacts(lockedDependency *domain.LockedDependency, dep domain.Dependency, rootPath string)
	MoveArtifacts(dep domain.Dependency, rootPath string)
}

// DefaultGraphBuilder implements GraphBuilder using the real graph functions.
type DefaultGraphBuilder struct{}

// LoadOrderGraph loads the dependency graph for changed packages only.
func (d *DefaultGraphBuilder) LoadOrderGraph(pkg *domain.Package) *domain.NodeQueue {
	return loadOrderGraph(pkg)
}

// LoadOrderGraphAll loads the complete dependency graph.
func (d *DefaultGraphBuilder) LoadOrderGraphAll(pkg *domain.Package) *domain.NodeQueue {
	return LoadOrderGraphAll(pkg)
}

// DefaultProjectCompiler implements ProjectCompiler.
type DefaultProjectCompiler struct{}

// Compile compiles a dproj file.
func (d *DefaultProjectCompiler) Compile(dprojPath string, dep *domain.Dependency, rootLock domain.PackageLock) bool {
	return compile(dprojPath, dep, rootLock, nil, nil)
}

// DefaultArtifactManager implements ArtifactManager.
type DefaultArtifactManager struct {
	service *ArtifactService
}

// NewDefaultArtifactManager creates a default artifact manager with OS filesystem.
func NewDefaultArtifactManager(fs infra.FileSystem) *DefaultArtifactManager {
	return &DefaultArtifactManager{
		service: NewArtifactService(fs),
	}
}

// EnsureArtifacts collects artifacts for a dependency.
func (d *DefaultArtifactManager) EnsureArtifacts(
	lockedDependency *domain.LockedDependency,
	dep domain.Dependency,
	rootPath string,
) {
	d.service.ensureArtifacts(lockedDependency, dep, rootPath)
}

// MoveArtifacts moves artifacts to the shared folder.
func (d *DefaultArtifactManager) MoveArtifacts(dep domain.Dependency, rootPath string) {
	d.service.moveArtifacts(dep, rootPath)
}
