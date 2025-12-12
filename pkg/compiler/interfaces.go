package compiler

import (
	"github.com/hashload/boss/pkg/compiler/graphs"
	"github.com/hashload/boss/pkg/models"
)

// PackageLoader abstracts loading package information.
type PackageLoader interface {
	LoadPackage(path string) (*models.Package, error)
}

// LockManager abstracts lock file operations.
type LockManager interface {
	Save() error
	GetInstalled(dep models.Dependency) models.LockedDependency
	SetInstalled(dep models.Dependency, locked models.LockedDependency)
}

// GraphBuilder abstracts dependency graph construction.
type GraphBuilder interface {
	LoadOrderGraph(pkg *models.Package) *graphs.NodeQueue
	LoadOrderGraphAll(pkg *models.Package) *graphs.NodeQueue
}

// ProjectCompiler abstracts project compilation.
type ProjectCompiler interface {
	Compile(dprojPath string, dep *models.Dependency, rootLock models.PackageLock) bool
}

// ArtifactManager abstracts artifact operations.
type ArtifactManager interface {
	EnsureArtifacts(lockedDependency *models.LockedDependency, dep models.Dependency, rootPath string)
	MoveArtifacts(dep models.Dependency, rootPath string)
}

// FileSystem abstracts file system operations for testability.
type FileSystem interface {
	WriteFile(name string, data []byte, perm int) error
	ReadDir(name string) ([]FileInfo, error)
	Rename(oldpath, newpath string) error
	RemoveAll(path string) error
	ReadFile(name string) ([]byte, error)
}

// FileInfo abstracts file information.
type FileInfo interface {
	Name() string
	IsDir() bool
}

// DefaultGraphBuilder implements GraphBuilder using the real graph functions.
type DefaultGraphBuilder struct{}

// LoadOrderGraph loads the dependency graph for changed packages only.
func (d *DefaultGraphBuilder) LoadOrderGraph(pkg *models.Package) *graphs.NodeQueue {
	return loadOrderGraph(pkg)
}

// LoadOrderGraphAll loads the complete dependency graph.
func (d *DefaultGraphBuilder) LoadOrderGraphAll(pkg *models.Package) *graphs.NodeQueue {
	return LoadOrderGraphAll(pkg)
}

// DefaultProjectCompiler implements ProjectCompiler.
type DefaultProjectCompiler struct{}

// Compile compiles a dproj file.
func (d *DefaultProjectCompiler) Compile(dprojPath string, dep *models.Dependency, rootLock models.PackageLock) bool {
	return compile(dprojPath, dep, rootLock)
}

// DefaultArtifactManager implements ArtifactManager.
type DefaultArtifactManager struct{}

// EnsureArtifacts collects artifacts for a dependency.
func (d *DefaultArtifactManager) EnsureArtifacts(
	lockedDependency *models.LockedDependency,
	dep models.Dependency,
	rootPath string,
) {
	ensureArtifacts(lockedDependency, dep, rootPath)
}

// MoveArtifacts moves artifacts to the shared folder.
func (d *DefaultArtifactManager) MoveArtifacts(dep models.Dependency, rootPath string) {
	moveArtifacts(dep, rootPath)
}
