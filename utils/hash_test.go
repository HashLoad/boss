package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hashload/boss/utils"
)

func TestHashDir_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	emptyDir := filepath.Join(tempDir, "empty")

	err := os.MkdirAll(emptyDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create empty dir: %v", err)
	}

	hash := utils.HashDir(emptyDir)
	if hash == "" {
		t.Error("HashDir returned empty string for empty directory")
	}
}

func TestHashDir_SingleFile(t *testing.T) {
	tempDir := t.TempDir()
	singleFileDir := filepath.Join(tempDir, "single")

	err := os.MkdirAll(singleFileDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}

	filePath := filepath.Join(singleFileDir, "test.txt")
	err = os.WriteFile(filePath, []byte("hello world"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	hash := utils.HashDir(singleFileDir)
	if hash == "" {
		t.Error("HashDir returned empty string")
	}
	if len(hash) != 32 {
		t.Errorf("HashDir returned invalid hash length: got %d, want 32", len(hash))
	}
}

func TestHashDir_SameContentSameHash(t *testing.T) {
	tempDir := t.TempDir()
	dir1 := filepath.Join(tempDir, "dir1")
	dir2 := filepath.Join(tempDir, "dir2")

	for _, dir := range []string{dir1, dir2} {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		err = os.WriteFile(filepath.Join(dir, "file.txt"), []byte("same content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	hash1 := utils.HashDir(dir1)
	hash2 := utils.HashDir(dir2)

	if hash1 != hash2 {
		t.Errorf("Same content should produce same hash: got %s and %s", hash1, hash2)
	}
}

func TestHashDir_DifferentContentDifferentHash(t *testing.T) {
	tempDir := t.TempDir()
	dir1 := filepath.Join(tempDir, "diff1")
	dir2 := filepath.Join(tempDir, "diff2")

	setupDir(t, dir1, "content A")
	setupDir(t, dir2, "content B")

	hash1 := utils.HashDir(dir1)
	hash2 := utils.HashDir(dir2)

	if hash1 == hash2 {
		t.Error("Different content should produce different hash")
	}
}

func TestHashDir_NestedDirectories(t *testing.T) {
	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "nested", "sub1", "sub2")

	err := os.MkdirAll(nestedDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create nested dir: %v", err)
	}

	err = os.WriteFile(filepath.Join(nestedDir, "deep.txt"), []byte("deep file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	hash := utils.HashDir(filepath.Join(tempDir, "nested"))
	if hash == "" {
		t.Error("HashDir returned empty string for nested directory")
	}
	if len(hash) != 32 {
		t.Errorf("HashDir returned invalid hash length: got %d, want 32", len(hash))
	}
}

func setupDir(t *testing.T, dir, content string) {
	t.Helper()
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}
	err = os.WriteFile(filepath.Join(dir, "file.txt"), []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
}
