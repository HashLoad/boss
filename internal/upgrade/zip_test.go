//nolint:testpackage // Testing internal functions
package upgrade

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestGetAssetFromFile_InvalidFile tests error handling for invalid file.
func TestGetAssetFromFile_InvalidFile(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "invalid.zip")

	// Create an empty file (not a valid zip)
	f, err := os.Create(tempFile)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	f.Close()

	file, err := os.Open(tempFile)
	if err != nil {
		t.Fatalf("Failed to open temp file: %v", err)
	}
	defer file.Close()

	_, err = getAssetFromFile(file, "test.zip")
	if err == nil {
		t.Error("getAssetFromFile() should return error for invalid zip")
	}
}

// TestReadFileFromZip_ValidZip tests reading from a valid zip file.
func TestReadFileFromZip_ValidZip(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "test.zip")

	// Create a valid zip file with expected structure
	expectedContent := []byte("test content")
	assetPath := fmt.Sprintf("%s-%s/boss", runtime.GOOS, runtime.GOARCH)

	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}

	w := zip.NewWriter(zipFile)
	f, err := w.Create(assetPath)
	if err != nil {
		t.Fatalf("Failed to create file in zip: %v", err)
	}
	_, err = f.Write(expectedContent)
	if err != nil {
		t.Fatalf("Failed to write to zip: %v", err)
	}
	w.Close()
	zipFile.Close()

	// Now read from it
	file, err := os.Open(zipPath)
	if err != nil {
		t.Fatalf("Failed to open zip: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	content, err := readFileFromZip(file, "test.zip", stat)
	if err != nil {
		t.Fatalf("readFileFromZip() error: %v", err)
	}

	if string(content) != string(expectedContent) {
		t.Errorf("readFileFromZip() content mismatch: got %s, want %s", content, expectedContent)
	}
}

// TestReadFileFromZip_AssetNotFound tests error when asset is not in zip.
func TestReadFileFromZip_AssetNotFound(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "test.zip")

	// Create a zip without the expected asset
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("Failed to create zip file: %v", err)
	}

	w := zip.NewWriter(zipFile)
	f, err := w.Create("other-file.txt")
	if err != nil {
		t.Fatalf("Failed to create file in zip: %v", err)
	}
	_, _ = f.Write([]byte("other content"))
	w.Close()
	zipFile.Close()

	file, err := os.Open(zipPath)
	if err != nil {
		t.Fatalf("Failed to open zip: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	_, err = readFileFromZip(file, "test.zip", stat)
	if err == nil {
		t.Error("readFileFromZip() should return error when asset not found")
	}
}

// TestReadFileFromTargz_InvalidFile tests error handling for invalid targz.
func TestReadFileFromTargz_InvalidFile(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "invalid.tar.gz")

	// Create an empty file (not a valid tar.gz)
	f, err := os.Create(tempFile)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	f.Close()

	file, err := os.Open(tempFile)
	if err != nil {
		t.Fatalf("Failed to open temp file: %v", err)
	}
	defer file.Close()

	_, err = readFileFromTargz(file, "test.tar.gz")
	if err == nil {
		t.Error("readFileFromTargz() should return error for invalid targz")
	}
}
