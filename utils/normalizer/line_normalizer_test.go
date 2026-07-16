package normalizer_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashload/boss/utils/normalizer"
)

func TestNormalizeDirectoryLineEndings(t *testing.T) {
	tempDir := t.TempDir()

	// 1. Create a Delphi .pas file with LF line endings
	pasPath := filepath.Join(tempDir, "unit.pas")
	pasContent := []byte("unit Test;\ninterface\n\nimplementation\nend.\n")
	if err := os.WriteFile(pasPath, pasContent, 0600); err != nil {
		t.Fatalf("failed to write pas file: %v", err)
	}

	// 2. Create a non-Delphi .txt file with LF line endings
	txtPath := filepath.Join(tempDir, "readme.txt")
	txtContent := []byte("hello\nworld\n")
	if err := os.WriteFile(txtPath, txtContent, 0600); err != nil {
		t.Fatalf("failed to write txt file: %v", err)
	}

	// Run the normalizer
	if err := normalizer.NormalizeDirectoryLineEndings(tempDir); err != nil {
		t.Fatalf("NormalizeDirectoryLineEndings failed: %v", err)
	}

	// 3. Verify .pas file was converted to CRLF
	newPasContent, err := os.ReadFile(pasPath)
	if err != nil {
		t.Fatalf("failed to read pas file: %v", err)
	}
	if !bytes.Contains(newPasContent, []byte("\r\n")) {
		t.Error("expected pas file to contain CRLF line endings")
	}
	if bytes.Contains(newPasContent, []byte("\n")) && !bytes.Contains(newPasContent, []byte("\r\n")) {
		t.Error("expected pas file to have all LFs preceded by CR")
	}

	// 4. Verify .txt file was NOT converted (remains LF)
	newTxtContent, err := os.ReadFile(txtPath)
	if err != nil {
		t.Fatalf("failed to read txt file: %v", err)
	}
	if bytes.Contains(newTxtContent, []byte("\r\n")) {
		t.Error("expected txt file to NOT contain CRLF line endings")
	}
}
