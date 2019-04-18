package dcp

import (
	"fmt"
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/env"
	"github.com/hashload/boss/msg"
	"github.com/hashload/boss/utils/librarypath"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

var encode = charmap.Windows1252

func InjectDpcs() {
	dprojNames := librarypath.GetDprojNames()

	for _, value := range dprojNames {
		if fileName, exists := getDprDpkFromDproj(filepath.Base(value)); exists {
			file := readFile(fileName)
			file = processFile(file, getDpcNames())
			writeFile(fileName, file)

		}
	}
}

func readFile(filename string) string {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	r := transform.NewReader(f, encode.NewDecoder())

	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	return string(bytes)
}

func writeFile(filename string, content string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	w := transform.NewWriter(f, encode.NewEncoder())
	_, err = fmt.Fprint(w, content)
	if err != nil {
		log.Fatal(err)
	}
	if err = f.Close(); err != nil {
		log.Fatal(err)
	}
}

func getDprDpkFromDproj(dprojName string) (filename string, find bool) {
	baseName := strings.TrimSuffix(dprojName, filepath.Ext(dprojName))
	dpkName := baseName + ".dpk"

	if _, err := os.Stat(dpkName); !os.IsNotExist(err) {
		return dpkName, true
	}
	return "", false
}

func getDpcNames() []string {
	fileInfos, err := ioutil.ReadDir(filepath.Join(env.GetModulesDir(), consts.DcpFolder))
	if err != nil {
		msg.Err("Falied on load dcps")
		return []string{}
	}

	dpcs := []string{}

	for _, value := range fileInfos {
		if strings.ToLower(filepath.Ext(value.Name())) == ".dcp" {
			dpcs = append(dpcs, strings.TrimSuffix(value.Name(), path.Ext(value.Name())))
		}
	}
	return dpcs
}

const COMMENT_BOSS = "{BOSS}"

func getDcpString(dcps []string) string {
	var dpsLine = "\n"

	for _, dcp := range dcps {
		dpsLine += "  " + filepath.Base(dcp) + COMMENT_BOSS + ",\n"
	}
	return dpsLine[:len(dpsLine)-2]
}

func injectDcps(filecontent string, dcps []string) (string, bool) {
	regexRequires := regexp.MustCompile(`(?m)^(requires)([\n\r \w,{}\\.]+)(;)`)

	resultRegex := regexRequires.FindAllStringSubmatch(filecontent, -1)
	if len(resultRegex) == 0 {
		return filecontent, false
	}

	resultRegexIndexes := regexRequires.FindAllStringSubmatchIndex(filecontent, -1)

	currentRequiresString := regexp.MustCompile("[\r\n ]+").ReplaceAllString(resultRegex[0][2], "")

	currentRequires := strings.Split(currentRequiresString, ",")

	var result = filecontent[:resultRegexIndexes[0][3]]

	for _, value := range currentRequires {
		if strings.Contains(value, COMMENT_BOSS) {
			continue
		}
		result += "\n  " + value + ","
	}

	result = result + getDcpString(dcps) + ";" + filecontent[resultRegexIndexes[0][7]:]
	return result, true
}

func processFile(content string, dcps []string) string {
	if len(dcps) == 0 {
		return content
	}
	if content, success := injectDcps(content, dcps); success {
		return content
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
	return content

}
