package models_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/models"
)

func TestCacheRepositoryDetails_And_RepoData(t *testing.T) {
	// Create a temp directory for BOSS_HOME
	tempDir := t.TempDir()
	t.Setenv("BOSS_HOME", tempDir)

	// Create the boss home folder structure
	bossHome := filepath.Join(tempDir, consts.FolderBossHome)
	cacheDir := filepath.Join(bossHome, "cache")
	infoDir := filepath.Join(cacheDir, "info")
	if err := os.MkdirAll(infoDir, 0755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	// Create a dependency
	dep := models.ParseDependency("github.com/hashload/horse", "^1.0.0")
	versions := []string{"1.0.0", "1.1.0", "1.2.0"}

	// Cache the repository details
	models.CacheRepositoryDetails(dep, versions)

	// Verify the file was created
	hashName := dep.HashName()
	jsonPath := filepath.Join(infoDir, hashName+".json")
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Error("CacheRepositoryDetails() should create JSON file")
	}

	// Read back the data
	repoInfo, err := models.RepoData(hashName)
	if err != nil {
		t.Errorf("RepoData() error = %v", err)
	}

	if repoInfo.Name != "horse" {
		t.Errorf("RepoData().Name = %q, want %q", repoInfo.Name, "horse")
	}

	if len(repoInfo.Versions) != 3 {
		t.Errorf("RepoData().Versions count = %d, want 3", len(repoInfo.Versions))
	}
}

func TestRepoData_NonExistent(t *testing.T) {
	// Create a temp directory for BOSS_HOME
	tempDir := t.TempDir()
	t.Setenv("BOSS_HOME", tempDir)

	// Create the boss home folder structure
	bossHome := filepath.Join(tempDir, consts.FolderBossHome)
	cacheDir := filepath.Join(bossHome, "cache")
	infoDir := filepath.Join(cacheDir, "info")
	if err := os.MkdirAll(infoDir, 0755); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	// Try to read non-existent data
	_, err := models.RepoData("nonexistent")
	if err == nil {
		t.Error("RepoData() should return error for non-existent key")
	}
}

func TestRepoInfo_Struct(t *testing.T) {
	info := models.RepoInfo{
		Key:      "abc123",
		Name:     "test-repo",
		Versions: []string{"1.0.0", "2.0.0"},
	}

	if info.Key != "abc123" {
		t.Errorf("Key = %q, want %q", info.Key, "abc123")
	}
	if info.Name != "test-repo" {
		t.Errorf("Name = %q, want %q", info.Name, "test-repo")
	}
	if len(info.Versions) != 2 {
		t.Errorf("Versions count = %d, want 2", len(info.Versions))
	}
}
