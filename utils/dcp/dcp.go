// Package dcp provides functionality for managing Delphi DCP (Delphi Compiled Package) files.
// It handles injection of DCP dependencies into project files (.dpr, .dpk).
package dcp

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashload/boss/internal/core/domain"
	"github.com/hashload/boss/pkg/consts"
	"github.com/hashload/boss/pkg/msg"
	"github.com/hashload/boss/utils/librarypath"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

var (
	reRequires   = regexp.MustCompile(`(?m)^(requires)([^;]+)(;)`)
	reWhitespace = regexp.MustCompile(`[\r\n ]+`)
)

// InjectDpcs injects DCP dependencies into project files.
func InjectDpcs(pkg *domain.Package, lock domain.PackageLock) {
	dprojNames := librarypath.GetProjectNames(pkg)

	for _, value := range dprojNames {
		if fileName, exists := getDprDpkFromDproj(filepath.Base(value)); exists {
			InjectDpcsFile(fileName, pkg, lock)
		}
	}
}

// InjectDpcsFile injects DCP dependencies into a specific file.
func InjectDpcsFile(fileName string, pkg *domain.Package, lock domain.PackageLock) {
	dprDpkFileName, exists := getDprDpkFromDproj(fileName)
	if !exists {
		return
	}

	file := readFile(dprDpkFileName)
	requiresList := getRequiresList(pkg, lock)

	if processedFile, needWrite := processFile(file, requiresList); needWrite {
		writeFile(dprDpkFileName, processedFile)
	}
}

// readFile reads a file with Windows1252 encoding.
func readFile(filename string) string {
	f, err := os.Open(filename) // #nosec G304 -- Reading DCP files from controlled package directories
	if err != nil {
		msg.Die(err.Error())
	}
	r := transform.NewReader(f, charmap.Windows1252.NewDecoder())

	bytes, err := io.ReadAll(r)
	if err != nil {
		msg.Die(err.Error())
	}

	return string(bytes)
}

// writeFile writes a file with Windows1252 encoding.
func writeFile(filename string, content string) {
	f, err := os.Create(filename) // #nosec G304 -- Writing DCP files to controlled package directories
	if err != nil {
		msg.Die(err.Error())
	}
	w := transform.NewWriter(f, charmap.Windows1252.NewEncoder())
	_, err = fmt.Fprint(w, content)
	if err != nil {
		msg.Die(err.Error())
	}
	if err = f.Close(); err != nil {
		msg.Die(err.Error())
	}
}

// getDprDpkFromDproj returns the DPR or DPK file name from a DPROJ file name.
func getDprDpkFromDproj(dprojName string) (string, bool) {
	baseName := strings.TrimSuffix(dprojName, filepath.Ext(dprojName))
	dpkName := baseName + consts.FileExtensionDpk

	if _, err := os.Stat(dpkName); !os.IsNotExist(err) {
		return dpkName, true
	}
	return "", false
}

// CommentBoss is the marker for Boss injected dependencies.
const CommentBoss = "{BOSS}"

// getDcpString returns the DCP requires string formatted for injection.
func getDcpString(dcps []string) string {
	var dcpRequiresLine = "\n"

	for _, dcp := range dcps {
		dcpRequiresLine += "  " + filepath.Base(dcp) + CommentBoss + ",\n"
	}
	return dcpRequiresLine[:len(dcpRequiresLine)-2]
}

// injectDcps injects DCP dependencies into the file content while preserving original formatting, comments, and conditionals.
func injectDcps(filecontent string, dcps []string) (string, bool) {
	resultRegex := reRequires.FindAllStringSubmatch(filecontent, -1)
	if len(resultRegex) == 0 {
		return filecontent, false
	}

	resultRegexIndexes := reRequires.FindAllStringSubmatchIndex(filecontent, -1)

	// Group 2 (indices 4 and 5) is the body of the requires section
	body := filecontent[resultRegexIndexes[0][4]:resultRegexIndexes[0][5]]

	// Find all existing boss-injected dependencies in the body.
	// They are marked with {BOSS}. We match them as \b([\w\.\-]+)\{BOSS\}
	reBossDep := regexp.MustCompile(`(?i)\b([\w\.\-]+)\{BOSS\}`)
	bossDepsMatches := reBossDep.FindAllStringSubmatch(body, -1)

	existingBossDeps := make(map[string]bool)
	for _, match := range bossDepsMatches {
		existingBossDeps[strings.ToLower(match[1])] = true
	}

	// Prepare a map of the target dcps (lowercase for case-insensitive comparison)
	targetDcpsMap := make(map[string]bool)
	for _, dcp := range dcps {
		targetDcpsMap[strings.ToLower(filepath.Base(dcp))] = true
	}

	// 1. Remove old boss deps that are no longer in the target dcps list
	modifiedBody := body
	for oldDcp := range existingBossDeps {
		if !targetDcpsMap[oldDcp] {
			// Remove this dependency, its {BOSS} comment, and any trailing comma/whitespace
			escapedDcp := regexp.QuoteMeta(oldDcp)
			reRemove := regexp.MustCompile(`(?i)\s*\b` + escapedDcp + `\{BOSS\}\s*,?`)
			modifiedBody = reRemove.ReplaceAllString(modifiedBody, "")
		}
	}

	// 2. Add new dcps that are not already present in the requires section
	for _, dcp := range dcps {
		dcpName := filepath.Base(dcp)
		dcpNameLower := strings.ToLower(dcpName)

		// Check if it's already in the requires section (case-insensitive)
		// We match it as a whole word: \b<dcpName>\b
		escapedDcp := regexp.QuoteMeta(dcpNameLower)
		reCheck := regexp.MustCompile(`(?i)\b` + escapedDcp + `\b`)
		if reCheck.MatchString(modifiedBody) {
			continue // Already exists (either user-defined or already injected), skip
		}

		// Append the new dependency
		trimmed := strings.TrimSpace(modifiedBody)
		if trimmed != "" && !strings.HasSuffix(trimmed, ",") {
			modifiedBody = trimmed + ",\n  " + dcpName + CommentBoss
		} else {
			if trimmed == "" {
				modifiedBody = "\n  " + dcpName + CommentBoss
			} else {
				modifiedBody = modifiedBody + "\n  " + dcpName + CommentBoss
			}
		}
	}

	// Reconstruct the file content by replacing the body
	result := filecontent[:resultRegexIndexes[0][4]] + modifiedBody + filecontent[resultRegexIndexes[0][5]:]
	return result, true
}

// processFile processes the file content to inject DCP dependencies.
// Returns the modified content and a boolean indicating if the file was changed.
func processFile(content string, dcps []string) (string, bool) {
	if len(dcps) == 0 {
		return content, false
	}
	if injectedContent, success := injectDcps(content, dcps); success {
		return injectedContent, true
	}

	lines := strings.Split(content, "\n")

	var dcpRequiresLine = getDcpString(dcps)
	var containsLineIndex = 1

	for key, value := range lines {
		if strings.TrimSpace(strings.ToLower(value)) == "contains" {
			containsLineIndex = key - 1
			break
		}
	}

	content = strings.Join(lines[:containsLineIndex], "\n\n") +
		"requires" + dcpRequiresLine + ";\n\n" + strings.Join(lines[containsLineIndex:], "\n")
	return content, true
}
