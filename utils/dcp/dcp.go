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
	"github.com/hashload/boss/utils"
	"github.com/hashload/boss/utils/librarypath"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

var (
	reRequires   = regexp.MustCompile(`(?m)^(requires)([\n\r \w,{}\\.]+)(;)`)
	reWhitespace = regexp.MustCompile(`[\r\n ]+`)
)

// InjectDpcs injects DCP dependencies into project files
func InjectDpcs(pkg *domain.Package, lock domain.PackageLock) {
	dprojNames := librarypath.GetProjectNames(pkg)

	for _, value := range dprojNames {
		if fileName, exists := getDprDpkFromDproj(filepath.Base(value)); exists {
			InjectDpcsFile(fileName, pkg, lock)
		}
	}
}

// InjectDpcsFile injects DCP dependencies into a specific file
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

// readFile reads a file with Windows1252 encoding
func readFile(filename string) string {
	f, err := os.Open(filename)
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

// writeFile writes a file with Windows1252 encoding
func writeFile(filename string, content string) {
	f, err := os.Create(filename)
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

// getDprDpkFromDproj returns the DPR or DPK file name from a DPROJ file name
func getDprDpkFromDproj(dprojName string) (string, bool) {
	baseName := strings.TrimSuffix(dprojName, filepath.Ext(dprojName))
	dpkName := baseName + consts.FileExtensionDpk

	if _, err := os.Stat(dpkName); !os.IsNotExist(err) {
		return dpkName, true
	}
	return "", false
}

// CommentBoss is the marker for Boss injected dependencies
const CommentBoss = "{BOSS}"

// getDcpString returns the DCP requires string
func getDcpString(dcps []string) string {
	var dpsLine = "\n"

	for _, dcp := range dcps {
		dpsLine += "  " + filepath.Base(dcp) + CommentBoss + ",\n"
	}
	return dpsLine[:len(dpsLine)-2]
}

// injectDcps injects DCP dependencies into the file content
func injectDcps(filecontent string, dcps []string) (string, bool) {
	resultRegex := reRequires.FindAllStringSubmatch(filecontent, -1)
	if len(resultRegex) == 0 {
		return filecontent, false
	}

	resultRegexIndexes := reRequires.FindAllStringSubmatchIndex(filecontent, -1)

	currentRequiresString := reWhitespace.ReplaceAllString(resultRegex[0][2], "")

	currentRequires := strings.Split(currentRequiresString, ",")

	var result = filecontent[:resultRegexIndexes[0][3]]

	for _, value := range currentRequires {
		if strings.Contains(value, CommentBoss) || utils.Contains(dcps, value) {
			continue
		}
		result += "\n  " + value + ","
	}

	result = result + getDcpString(dcps) + ";" + filecontent[resultRegexIndexes[0][7]:]
	return result, true
}

// processFile processes the file content to inject DCP dependencies
func processFile(content string, dcps []string) (string, bool) {
	if len(dcps) == 0 {
		return content, false
	}
	if injectedContent, success := injectDcps(content, dcps); success {
		return injectedContent, true
	}

	lines := strings.Split(content, "\n")

	var dpcLine = getDcpString(dcps)
	var containsindex = 1

	for key, value := range lines {
		if strings.TrimSpace(strings.ToLower(value)) == "contains" {
			containsindex = key - 1
			break
		}
	}

	content = strings.Join(lines[:containsindex], "\n\n") +
		"requires" + dpcLine + ";\n\n" + strings.Join(lines[containsindex:], "\n")
	return content, true
}
