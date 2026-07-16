// Package normalizer provides line ending normalization for Delphi source files.
package normalizer

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashload/boss/pkg/msg"
)

//nolint:gochecknoglobals // Immutable lookup map for Delphi file extensions
var delphiExtensions = map[string]bool{
	".pas":   true,
	".inc":   true,
	".dfm":   true,
	".dpk":   true,
	".dproj": true,
}

// NormalizeDirectoryLineEndings walks the given directory and converts LF to CRLF in all Delphi-related files.
func NormalizeDirectoryLineEndings(dir string) error {
	msg.Debug("  🧹 Normalizing line endings to CRLF in %s", dir)
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !delphiExtensions[ext] {
			return nil
		}

		return NormalizeFileLineEndings(path)
	})
}

// NormalizeFileLineEndings converts all LF to CRLF in a specific file.
func NormalizeFileLineEndings(filePath string) error {
	content, err := os.ReadFile(filePath) // #nosec G304 -- Reading Delphi files from controlled package directories
	if err != nil {
		return err
	}

	// Check if there are any raw LFs (LFs not preceded by CR)
	hasLF := false
	for i, b := range content {
		if b == '\n' {
			if i == 0 || content[i-1] != '\r' {
				hasLF = true
				break
			}
		}
	}

	if !hasLF {
		return nil // Already CRLF or has no LFs
	}

	msg.Debug("     Normalizing line endings for %s", filepath.Base(filePath))

	// Normalize: replace all CRLF with LF, then replace all LF with CRLF
	normalized := bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
	crlfContent := bytes.ReplaceAll(normalized, []byte("\n"), []byte("\r\n"))

	// Write back
	//nolint:lll // #nosec comment makes line long
	return os.WriteFile(filePath, crlfContent, 0600) // #nosec G304,G703 -- Writing Delphi files from controlled package directories
}
