package paths_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/internal/core/services/paths"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/env"
)

func TestEnsureCacheDir(t *testing.T) {
	// Create a temp directory for BOSS_HOME
	tempDir := t.TempDir()
	t.Setenv("BOSS_HOME", tempDir)

	// Create the boss home folder structure
	bossHome := filepath.Join(tempDir, consts.FolderBossHome)
	if err := os.MkdirAll(bossHome, 0755); err != nil {
		t.Fatalf("Failed to create boss home: %v", err)
	}

	// Create a dependency
	dep := domain.ParseDependency("github.com/hashload/horse", "^1.0.0")

	// Ensure cache dir (should not panic)
	paths.EnsureCacheDir(dep)

	// Verify the cache dir was created if GitEmbedded is true
	config := env.GlobalConfiguration()
	if config.GitEmbedded {
		cacheDir := filepath.Join(bossHome, "cache", dep.HashName())
		if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
			t.Error("EnsureCacheDir() should create cache directory when GitEmbedded is true")
		}
	}
}

func TestEnsureCleanModulesDir_CreatesDir(t *testing.T) {
	// Create a temp directory for workspace
	tempDir := t.TempDir()

	// Save original state and set not global
	originalGlobal := env.GetGlobal()
	defer env.SetGlobal(originalGlobal)
	env.SetGlobal(false)

	// Change to temp directory
	t.Chdir(tempDir)

	// Create empty dependencies and lock
	deps := []domain.Dependency{}
	lock := domain.PackageLock{
		Installed: map[string]domain.LockedDependency{},
	}

	// EnsureCleanModulesDir should create the modules directory
	paths.EnsureCleanModulesDir(deps, lock)

	// Verify modules directory was created
	modulesDir := filepath.Join(tempDir, consts.FolderDependencies)
	if _, err := os.Stat(modulesDir); os.IsNotExist(err) {
		t.Error("EnsureCleanModulesDir() should create modules directory")
	}

	// Verify default paths were created
	for _, path := range consts.DefaultPaths() {
		pathDir := filepath.Join(modulesDir, path)
		if _, err := os.Stat(pathDir); os.IsNotExist(err) {
			t.Errorf("EnsureCleanModulesDir() should create default path: %s", path)
		}
	}
}

func TestEnsureCleanModulesDir_RemovesOldDependencies(t *testing.T) {
	// Create a temp directory for workspace
	tempDir := t.TempDir()

	// Save original state and set not global
	originalGlobal := env.GetGlobal()
	defer env.SetGlobal(originalGlobal)
	env.SetGlobal(false)

	// Change to temp directory
	t.Chdir(tempDir)

	// Create modules directory with old dependency
	modulesDir := filepath.Join(tempDir, consts.FolderDependencies)
	if err := os.MkdirAll(modulesDir, 0755); err != nil {
		t.Fatalf("Failed to create modules dir: %v", err)
	}

	// Create an old dependency directory that should be removed
	oldDepDir := filepath.Join(modulesDir, "old-dependency")
	if err := os.MkdirAll(oldDepDir, 0755); err != nil {
		t.Fatalf("Failed to create old dependency dir: %v", err)
	}

	// Create a current dependency directory that should be kept
	currentDepDir := filepath.Join(modulesDir, "horse")
	if err := os.MkdirAll(currentDepDir, 0755); err != nil {
		t.Fatalf("Failed to create current dependency dir: %v", err)
	}

	// Define current dependencies
	dep := domain.ParseDependency("github.com/hashload/horse", "^1.0.0")
	deps := []domain.Dependency{dep}
	lock := domain.PackageLock{
		Installed: map[string]domain.LockedDependency{},
	}

	// EnsureCleanModulesDir should remove old dependency
	paths.EnsureCleanModulesDir(deps, lock)

	// Verify old dependency was removed
	if _, err := os.Stat(oldDepDir); !os.IsNotExist(err) {
		t.Error("EnsureCleanModulesDir() should remove old dependency directories")
	}

	// Verify current dependency was kept
	if _, err := os.Stat(currentDepDir); os.IsNotExist(err) {
		t.Error("EnsureCleanModulesDir() should keep current dependency directories")
	}
}
