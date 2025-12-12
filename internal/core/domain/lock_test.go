package domain_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashload/boss/internal/core/domain"
)

func TestDependencyArtifacts_Clean(t *testing.T) {
	artifacts := domain.DependencyArtifacts{
		Bin: []string{"file1.exe", "file2.exe"},
		Dcp: []string{"file1.dcp"},
		Dcu: []string{"file1.dcu", "file2.dcu"},
		Bpl: []string{"file1.bpl"},
	}

	artifacts.Clean()

	if len(artifacts.Bin) != 0 {
		t.Errorf("Bin should be empty after Clean(), got %d items", len(artifacts.Bin))
	}
	if len(artifacts.Dcp) != 0 {
		t.Errorf("Dcp should be empty after Clean(), got %d items", len(artifacts.Dcp))
	}
	if len(artifacts.Dcu) != 0 {
		t.Errorf("Dcu should be empty after Clean(), got %d items", len(artifacts.Dcu))
	}
	if len(artifacts.Bpl) != 0 {
		t.Errorf("Bpl should be empty after Clean(), got %d items", len(artifacts.Bpl))
	}
}

func TestLockedDependency_GetArtifacts(t *testing.T) {
	tests := []struct {
		name     string
		locked   domain.LockedDependency
		expected int
	}{
		{
			name: "all artifact types",
			locked: domain.LockedDependency{
				Artifacts: domain.DependencyArtifacts{
					Bin: []string{"a.exe", "b.exe"},
					Dcp: []string{"c.dcp"},
					Dcu: []string{"d.dcu", "e.dcu"},
					Bpl: []string{"f.bpl"},
				},
			},
			expected: 6,
		},
		{
			name: "empty artifacts",
			locked: domain.LockedDependency{
				Artifacts: domain.DependencyArtifacts{},
			},
			expected: 0,
		},
		{
			name: "only bin",
			locked: domain.LockedDependency{
				Artifacts: domain.DependencyArtifacts{
					Bin: []string{"only.exe"},
				},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.locked.GetArtifacts()
			if len(result) != tt.expected {
				t.Errorf("GetArtifacts() returned %d items, want %d", len(result), tt.expected)
			}
		})
	}
}

func TestPackageLock_GetInstalled(t *testing.T) {
	lock := domain.PackageLock{
		Installed: map[string]domain.LockedDependency{
			"github.com/hashload/boss": {
				Name:    "boss",
				Version: "1.0.0",
				Hash:    "abc123",
			},
			"github.com/hashload/horse": {
				Name:    "horse",
				Version: "2.0.0",
				Hash:    "def456",
			},
		},
	}

	t.Run("get existing dependency", func(t *testing.T) {
		dep := domain.Dependency{Repository: "github.com/hashload/boss"}
		result := lock.GetInstalled(dep)

		if result.Name != "boss" {
			t.Errorf("Name = %q, want %q", result.Name, "boss")
		}
		if result.Version != "1.0.0" {
			t.Errorf("Version = %q, want %q", result.Version, "1.0.0")
		}
	})

	t.Run("get non-existing dependency", func(t *testing.T) {
		dep := domain.Dependency{Repository: "github.com/hashload/notexists"}
		result := lock.GetInstalled(dep)

		if result.Name != "" {
			t.Errorf("Name should be empty for non-existing dependency, got %q", result.Name)
		}
	})

	t.Run("case insensitive lookup", func(t *testing.T) {
		dep := domain.Dependency{Repository: "GITHUB.COM/HASHLOAD/BOSS"}
		result := lock.GetInstalled(dep)

		if result.Name != "boss" {
			t.Errorf("Should find dependency case-insensitively, got Name = %q", result.Name)
		}
	})
}

func TestPackageLock_CleanRemoved(t *testing.T) {
	lock := domain.PackageLock{
		Installed: map[string]domain.LockedDependency{
			"github.com/hashload/boss": {
				Name:    "boss",
				Version: "1.0.0",
			},
			"github.com/hashload/horse": {
				Name:    "horse",
				Version: "2.0.0",
			},
			"github.com/hashload/old": {
				Name:    "old",
				Version: "1.0.0",
			},
		},
	}

	currentDeps := []domain.Dependency{
		{Repository: "github.com/hashload/boss"},
		{Repository: "github.com/hashload/horse"},
	}

	lock.CleanRemoved(currentDeps)

	if len(lock.Installed) != 2 {
		t.Errorf("Installed count = %d, want 2", len(lock.Installed))
	}

	for key := range lock.Installed {
		if strings.Contains(key, "old") {
			t.Error("'old' dependency should have been removed")
		}
	}
}

func TestPackageLock_GetArtifactList(t *testing.T) {
	lock := domain.PackageLock{
		Installed: map[string]domain.LockedDependency{
			"github.com/hashload/boss": {
				Artifacts: domain.DependencyArtifacts{
					Bin: []string{"boss.exe"},
					Bpl: []string{"boss.bpl"},
				},
			},
			"github.com/hashload/horse": {
				Artifacts: domain.DependencyArtifacts{
					Dcu: []string{"horse.dcu"},
					Dcp: []string{"horse.dcp"},
				},
			},
		},
	}

	result := lock.GetArtifactList()

	if len(result) != 4 {
		t.Errorf("GetArtifactList() returned %d items, want 4", len(result))
	}

	expected := map[string]bool{
		"boss.exe":  false,
		"boss.bpl":  false,
		"horse.dcu": false,
		"horse.dcp": false,
	}

	for _, artifact := range result {
		if _, exists := expected[artifact]; exists {
			expected[artifact] = true
		}
	}

	for artifact, found := range expected {
		if !found {
			t.Errorf("Expected artifact %q not found in result", artifact)
		}
	}
}

func TestPackageLock_SetInstalled(t *testing.T) {
	lock := domain.PackageLock{
		Installed: map[string]domain.LockedDependency{},
	}

	dep := domain.Dependency{Repository: "github.com/hashload/boss"}
	locked := domain.LockedDependency{
		Name:    "boss",
		Version: "1.0.0",
	}

	lock.SetInstalled(dep, locked)

	result := lock.GetInstalled(dep)
	if result.Name != "boss" {
		t.Errorf("SetInstalled did not store dependency correctly, got Name = %q", result.Name)
	}
	if result.Version != "1.0.0" {
		t.Errorf("SetInstalled did not store version correctly, got Version = %q", result.Version)
	}
}

func TestLockedDependency_CheckArtifactsType(t *testing.T) {
	tempDir := t.TempDir()

	// Create test artifact files
	artifactFiles := []string{"test.bpl", "test2.bpl"}
	for _, f := range artifactFiles {
		path := filepath.Join(tempDir, f)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	locked := domain.LockedDependency{
		Artifacts: domain.DependencyArtifacts{
			Bpl: artifactFiles,
		},
	}

	// Test with existing files - should return true
	result := locked.CheckArtifactsType(tempDir, locked.Artifacts.Bpl)
	if !result {
		t.Error("CheckArtifactsType should return true when all artifacts exist")
	}

	// Test with non-existing files - should return false
	result = locked.CheckArtifactsType(tempDir, []string{"nonexistent.bpl"})
	if result {
		t.Error("CheckArtifactsType should return false when artifacts don't exist")
	}

	// Test with empty artifacts - should return true
	result = locked.CheckArtifactsType(tempDir, []string{})
	if !result {
		t.Error("CheckArtifactsType should return true for empty artifact list")
	}
}

func TestLockedDependency_Failed_And_Changed_Flags(t *testing.T) {
	locked := domain.LockedDependency{
		Failed:  false,
		Changed: false,
	}

	// Verify initial state
	if locked.Failed {
		t.Error("Failed flag should be false initially")
	}
	if locked.Changed {
		t.Error("Changed flag should be false initially")
	}

	// Test setting Failed flag
	locked.Failed = true
	if !locked.Failed {
		t.Error("Failed flag should be true after setting")
	}

	// Test setting Changed flag
	locked.Changed = true
	if !locked.Changed {
		t.Error("Changed flag should be true after setting")
	}
}

func TestPackageLock_EmptyInstalled(t *testing.T) {
	lock := domain.PackageLock{
		Installed: map[string]domain.LockedDependency{},
	}

	// GetArtifactList on empty installed should return nil/empty
	artifacts := lock.GetArtifactList()
	if len(artifacts) != 0 {
		t.Errorf("GetArtifactList() on empty lock should return empty, got %d items", len(artifacts))
	}

	// CleanRemoved on empty should not panic
	lock.CleanRemoved([]domain.Dependency{})
}

func TestDependencyArtifacts_AllTypes(t *testing.T) {
	artifacts := domain.DependencyArtifacts{
		Bin: []string{"a.exe", "b.exe"},
		Dcp: []string{"c.dcp"},
		Dcu: []string{"d.dcu", "e.dcu", "f.dcu"},
		Bpl: []string{"g.bpl"},
	}

	// Verify each type has correct count
	if len(artifacts.Bin) != 2 {
		t.Errorf("Bin count = %d, want 2", len(artifacts.Bin))
	}
	if len(artifacts.Dcp) != 1 {
		t.Errorf("Dcp count = %d, want 1", len(artifacts.Dcp))
	}
	if len(artifacts.Dcu) != 3 {
		t.Errorf("Dcu count = %d, want 3", len(artifacts.Dcu))
	}
	if len(artifacts.Bpl) != 1 {
		t.Errorf("Bpl count = %d, want 1", len(artifacts.Bpl))
	}

	// Clean should reset all
	artifacts.Clean()

	if len(artifacts.Bin) != 0 || len(artifacts.Dcp) != 0 ||
		len(artifacts.Dcu) != 0 || len(artifacts.Bpl) != 0 {
		t.Error("Clean() should empty all artifact slices")
	}
}

func TestPackageLock_MultipleOperations(t *testing.T) {
	lock := domain.PackageLock{
		Installed: map[string]domain.LockedDependency{},
	}

	// Add multiple dependencies
	deps := []domain.Dependency{
		{Repository: "github.com/hashload/boss"},
		{Repository: "github.com/hashload/horse"},
		{Repository: "github.com/hashload/dataset"},
	}

	for i, dep := range deps {
		locked := domain.LockedDependency{
			Name:    dep.Name(),
			Version: "1.0." + string(rune('0'+i)),
			Hash:    "hash" + string(rune('0'+i)),
		}
		lock.SetInstalled(dep, locked)
	}

	// Verify all were added
	if len(lock.Installed) != 3 {
		t.Errorf("Installed count = %d, want 3", len(lock.Installed))
	}

	// Get each one
	for _, dep := range deps {
		result := lock.GetInstalled(dep)
		if result.Name == "" {
			t.Errorf("GetInstalled(%s) returned empty", dep.Repository)
		}
	}

	// Clean removed - keep only first two
	lock.CleanRemoved(deps[:2])

	if len(lock.Installed) != 2 {
		t.Errorf("After CleanRemoved, Installed count = %d, want 2", len(lock.Installed))
	}
}

func TestLockedDependency_GetArtifacts_Order(t *testing.T) {
	locked := domain.LockedDependency{
		Artifacts: domain.DependencyArtifacts{
			Dcp: []string{"first.dcp"},
			Dcu: []string{"second.dcu"},
			Bin: []string{"third.exe"},
			Bpl: []string{"fourth.bpl"},
		},
	}

	result := locked.GetArtifacts()

	// Should contain all 4 artifacts
	if len(result) != 4 {
		t.Errorf("GetArtifacts() returned %d items, want 4", len(result))
	}

	// Verify all expected artifacts are present
	expected := map[string]bool{
		"first.dcp":  false,
		"second.dcu": false,
		"third.exe":  false,
		"fourth.bpl": false,
	}

	for _, artifact := range result {
		expected[artifact] = true
	}

	for name, found := range expected {
		if !found {
			t.Errorf("Artifact %q not found in result", name)
		}
	}
}
