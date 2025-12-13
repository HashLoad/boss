package repository

import (
	"encoding/json"
	"fmt"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/ports"
	"github.com/hashload/boss/internal/infra"
	"github.com/hashload/boss/utils/parser"
)

// Compile-time check that FilePackageRepository implements ports.PackageRepository.
var _ ports.PackageRepository = (*FilePackageRepository)(nil)

// FilePackageRepository implements PackageRepository using the filesystem.
type FilePackageRepository struct {
	fs infra.FileSystem
}

// NewFilePackageRepository creates a new FilePackageRepository.
func NewFilePackageRepository(fs infra.FileSystem) *FilePackageRepository {
	return &FilePackageRepository{fs: fs}
}

// Load loads a package from the given path.
func (r *FilePackageRepository) Load(packagePath string) (*domain.Package, error) {
	data, err := r.fs.ReadFile(packagePath)
	if err != nil {
		return nil, err
	}

	pkg := domain.NewPackage(packagePath)

	if err := json.Unmarshal(data, pkg); err != nil {
		return nil, fmt.Errorf("error unmarshaling package %s: %w", packagePath, err)
	}

	return pkg, nil
}

// Save persists the package to the given path.
func (r *FilePackageRepository) Save(pkg *domain.Package, packagePath string) error {
	data, err := parser.JSONMarshal(pkg, true)
	if err != nil {
		return err
	}

	return r.fs.WriteFile(packagePath, data, 0600)
}

// Exists checks if a package file exists at the given path.
func (r *FilePackageRepository) Exists(packagePath string) bool {
	return r.fs.Exists(packagePath)
}
